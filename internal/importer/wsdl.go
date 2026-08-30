// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package importer

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

const (
	wsdlNS       = "http://schemas.xmlsoap.org/wsdl/"
	xsdNS        = "http://www.w3.org/2001/XMLSchema"
	soap11BindNS = "http://schemas.xmlsoap.org/wsdl/soap/"
	soap12BindNS = "http://schemas.xmlsoap.org/wsdl/soap12/"
	soap11Env    = "http://schemas.xmlsoap.org/soap/envelope/"
	soap12Env    = "http://www.w3.org/2003/05/soap-envelope"
	maxBodyDepth = 3
)

// WSDLResult is the in-memory result of importing a WSDL document.
type WSDLResult struct {
	Title       string
	Collections []*WSDLCollection
}

// WSDLCollection groups generated requests per wsdl:service.
type WSDLCollection struct {
	Name    string
	Request []*wsdlEntry
}

// Write writes the imported result to disk as a Git-native workspace:
// reqly.yaml + collections/<service>/<operation>.yaml request files.
func (r *WSDLResult) Write(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create workspace dir: %w", err)
	}
	wsCfg := map[string]any{"name": r.Title}
	if err := writeYAMLFile(filepath.Join(dir, "reqly.yaml"), wsCfg); err != nil {
		return err
	}
	for _, coll := range r.Collections {
		collDir := filepath.Join(dir, "collections", sanitizeName(coll.Name))
		if err := os.MkdirAll(collDir, 0o755); err != nil {
			return fmt.Errorf("create collection dir: %w", err)
		}
		if err := writeYAMLFile(filepath.Join(collDir, "reqly.yaml"), map[string]any{"name": coll.Name}); err != nil {
			return err
		}
		used := map[string]int{}
		for _, entry := range coll.Request {
			base := sanitizeName(entry.Filename)
			used[base]++
			name := base
			if used[base] > 1 {
				name = fmt.Sprintf("%s-%d", base, used[base])
			}
			if err := writeYAMLFile(filepath.Join(collDir, name+".yaml"), entry.File); err != nil {
				return err
			}
		}
	}
	return nil
}

// wsdlEntry couples a generated filename with its request file content.
type wsdlEntry struct {
	Filename string
	File     *requestfile.File
}

// wsdlNode is a namespace-resolved element from the document tree.
type wsdlNode struct {
	Space, Local string
	Attrs        []xml.Attr
	Text         string
	parent       *wsdlNode
	Children     []*wsdlNode
}

func (n *wsdlNode) get(local string) *wsdlNode {
	for _, c := range n.Children {
		if c.Local == local {
			return c
		}
	}
	return nil
}

func (n *wsdlNode) all(local string) []*wsdlNode {
	var out []*wsdlNode
	for _, c := range n.Children {
		if c.Local == local {
			out = append(out, c)
		}
	}
	return out
}

func (n *wsdlNode) attr(local string) string {
	for _, a := range n.Attrs {
		if a.Name.Local == local && a.Name.Space != "xmlns" {
			return a.Value
		}
	}
	return ""
}

// xmlns resolves an xmlns prefix declared on this node or an ancestor.
func (n *wsdlNode) xmlns(prefix string) string {
	for cur := n; cur != nil; cur = cur.parent {
		for _, a := range cur.Attrs {
			if prefix == "" {
				if a.Name.Space == "" && a.Name.Local == "xmlns" {
					return a.Value
				}
			} else if a.Name.Space == "xmlns" && a.Name.Local == prefix {
				return a.Value
			}
		}
	}
	return ""
}

// qname resolves a QName attribute value against in-scope namespaces.
func (n *wsdlNode) qname(value string) QName {
	i := strings.Index(value, ":")
	if i < 0 {
		return QName{NS: n.xmlns(""), Local: value}
	}
	return QName{NS: n.xmlns(value[:i]), Local: value[i+1:]}
}

// QName is a namespace-resolved name.
type QName struct{ NS, Local string }

func (q QName) String() string { return "{" + q.NS + "}" + q.Local }

// parseTree builds the node tree; encoding/xml already resolves element and
// attribute namespaces into Name.Space.
func parseTree(data []byte) (*wsdlNode, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	var root *wsdlNode
	var stack []*wsdlNode
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			node := &wsdlNode{
				Space: t.Name.Space,
				Local: t.Name.Local,
				Attrs: append([]xml.Attr(nil), t.Attr...),
			}
			if len(stack) > 0 {
				node.parent = stack[len(stack)-1]
				node.parent.Children = append(node.parent.Children, node)
			} else if root == nil {
				root = node
			} else {
				return nil, fmt.Errorf("multiple root elements")
			}
			stack = append(stack, node)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].Text += strings.TrimSpace(string(t))
			}
		}
	}
	if root == nil {
		return nil, fmt.Errorf("empty XML document")
	}
	return root, nil
}

// xsdChild describes one element occurrence inside a complexType sequence.
type xsdChild struct {
	Name     string
	Type     QName
	Optional bool
}

// xsdShape is the child list of a global element or named complexType.
type xsdShape struct {
	NS       string // target namespace for rendering children
	Children []xsdChild
}

// ParseWSDL parses a WSDL 1.1 document into per-service collections of
// generated requests.
func ParseWSDL(data []byte) (*WSDLResult, *ImportReport, error) {
	return ParseWSDLWithBase(data, "")
}

// ParseWSDLWithBase parses a WSDL document with optional base directory for
// resolving local xsd:import/xsd:include schemaLocation values. When baseDir
// is non-empty, relative schemaLocation paths are resolved against it and
// their file contents are merged into the schema shapes; warnings are
// suppressed for successfully resolved imports.
func ParseWSDLWithBase(data []byte, baseDir string) (*WSDLResult, *ImportReport, error) {
	root, err := parseTree(data)
	if err != nil {
		return nil, nil, fmt.Errorf("parse XML: %w", err)
	}
	if root.Space != wsdlNS || root.Local != "definitions" {
		return nil, nil, fmt.Errorf("not a WSDL 1.1 document (root {%s}%s)", root.Space, root.Local)
	}

	res := &WSDLResult{Title: firstNonEmpty(root.attr("name"), "SOAP Service")}
	rep := NewReport("wsdl")
	targetNS := root.attr("targetNamespace")

	shapes := collectSchemasWithBase(root, rep, baseDir)
	messages := collectMessages(targetNS, root)
	portTypes := collectPortTypes(targetNS, root)
	bindings := collectBindings(targetNS, root)

	for _, svc := range root.all("service") {
		coll := &WSDLCollection{Name: firstNonEmpty(svc.attr("name"), "Service")}
		address, bindingQ, extraPorts := firstSOAPPort(svc)
		if len(extraPorts) > 0 {
			rep.Add(coll.Name, CategorySchema, SeverityWarned, "service %q: additional ports beyond the first SOAP port ignored: %s", coll.Name, strings.Join(extraPorts, ", "))
		}
		if address == "" {
			rep.Add(coll.Name, CategoryOther, SeverityDropped, "service %q has no SOAP port; skipped", coll.Name)
			continue
		}
		binding := bindings[bindingQ.String()]
		if binding == nil {
			rep.Add(coll.Name, CategorySchema, SeverityDropped, "service %q: binding %q not found; skipped", coll.Name, bindingQ.Local)
			continue
		}
		pt := portTypes[binding.portType.String()]
		if pt == nil {
			rep.Add(binding.name, CategorySchema, SeverityDropped, "binding %q: portType %q not found; skipped", binding.name, binding.portType.Local)
			continue
		}
		coll.Request = buildRequests(pt, binding, address, shapes, messages, rep)
		res.Collections = append(res.Collections, coll)
	}
	if len(res.Collections) == 0 {
		return nil, rep, fmt.Errorf("no services with SOAP ports found")
	}
	return res, rep, nil
}

// wsdlBinding captures soap:binding version/style plus per-operation details.
type wsdlBinding struct {
	name       string
	portType   QName
	style      string // document | rpc
	soap12     bool
	operations map[string]*bindingOp
}

type bindingOp struct {
	soapAction string
	use        string // literal | encoded
}

func collectBindings(targetNS string, root *wsdlNode) map[string]*wsdlBinding {
	out := map[string]*wsdlBinding{}
	for _, b := range root.all("binding") {
		bind := &wsdlBinding{name: b.attr("name")}
		bind.portType = b.qname(b.attr("type"))
		if sb := findSoapBinding(b); sb != nil {
			bind.soap12 = sb.Space == soap12BindNS
			bind.style = sb.attr("style")
		}
		bind.operations = map[string]*bindingOp{}
		for _, op := range b.all("operation") {
			bo := &bindingOp{use: "literal"}
			if so := findSoapOperation(op); so != nil {
				bo.soapAction = so.attr("soapAction")
			}
			if input := op.get("input"); input != nil {
				if body := findSoapBody(input); body != nil {
					bo.use = firstNonEmpty(body.attr("use"), "literal")
				}
			}
			bind.operations[op.attr("name")] = bo
		}
		out[QName{targetNS, b.attr("name")}.String()] = bind
	}
	return out
}

func findSoapBinding(b *wsdlNode) *wsdlNode {
	for _, c := range b.Children {
		if c.Local == "binding" && (c.Space == soap11BindNS || c.Space == soap12BindNS) {
			return c
		}
	}
	return nil
}

func findSoapOperation(op *wsdlNode) *wsdlNode {
	for _, c := range op.Children {
		if c.Local == "operation" && (c.Space == soap11BindNS || c.Space == soap12BindNS) {
			return c
		}
	}
	return nil
}

func findSoapBody(n *wsdlNode) *wsdlNode {
	for _, c := range n.Children {
		if c.Local == "body" && (c.Space == soap11BindNS || c.Space == soap12BindNS) {
			return c
		}
	}
	return nil
}

// wsdlPortTypeOp is one operation of a portType.
type wsdlPortTypeOp struct {
	name     string
	doc      string
	inputMsg QName
	inputRef *wsdlNode // for resolving part QNames in scope
}

func collectPortTypes(targetNS string, root *wsdlNode) map[string]*wsdlPortType {
	out := map[string]*wsdlPortType{}
	for _, pt := range root.all("portType") {
		p := &wsdlPortType{name: pt.attr("name")}
		for _, op := range pt.all("operation") {
			po := &wsdlPortTypeOp{name: op.attr("name")}
			if d := op.get("documentation"); d != nil {
				po.doc = d.Text
			}
			if input := op.get("input"); input != nil {
				po.inputMsg = input.qname(input.attr("message"))
				po.inputRef = input
			}
			p.operations = append(p.operations, po)
		}
		out[QName{targetNS, p.name}.String()] = p
	}
	return out
}

type wsdlPortType struct {
	name       string
	operations []*wsdlPortTypeOp
}

func (n *wsdlNode) text() string { return n.Text }

func collectMessages(targetNS string, root *wsdlNode) map[string]*wsdlMessage {
	out := map[string]*wsdlMessage{}
	for _, m := range root.all("message") {
		msg := &wsdlMessage{name: m.attr("name")}
		for _, part := range m.all("part") {
			p := &wsdlPart{name: part.attr("name")}
			if el := part.attr("element"); el != "" {
				p.element = part.qname(el)
			}
			if ty := part.attr("type"); ty != "" {
				p.typ = part.qname(ty)
			}
			p.hasEl = p.element.Local != ""
			msg.parts = append(msg.parts, p)
			_ = part
		}
		out[QName{targetNS, msg.name}.String()] = msg
	}
	return out
}

type wsdlMessage struct {
	name  string
	parts []*wsdlPart
}

type wsdlPart struct {
	name    string
	element QName
	typ     QName
	hasEl   bool
}

func collectSchemas(root *wsdlNode, rep *ImportReport) map[string]*xsdShape {
	return collectSchemasWithBase(root, rep, "")
}

func collectSchemasWithBase(root *wsdlNode, rep *ImportReport, baseDir string) map[string]*xsdShape {
	shapes := map[string]*xsdShape{} // "{ns}Name" → shape
	addSchema := func(schema *wsdlNode) {
		targetNS := schema.attr("targetNamespace")
		for _, imp := range schema.all("import") {
			loc := imp.attr("schemaLocation")
			if loc != "" && baseDir != "" && !isURL(loc) {
				candidate := filepath.Join(baseDir, loc)
				if data, err := os.ReadFile(candidate); err == nil {
					if extShapes := tryParseExternalSchema(data, rep); extShapes != nil {
						for k, v := range extShapes {
							shapes[k] = v
						}
						continue // resolved, no warning
					}
				}
			}
			rep.Add("", CategorySchema, SeverityWarned, "external xsd:import %q not followed; affected elements get skeleton-only bodies", imp.attr("schemaLocation"))
		}
		for _, inc := range schema.all("include") {
			loc := inc.attr("schemaLocation")
			if loc != "" && baseDir != "" && !isURL(loc) {
				candidate := filepath.Join(baseDir, loc)
				if data, err := os.ReadFile(candidate); err == nil {
					if extShapes := tryParseExternalSchema(data, rep); extShapes != nil {
						for k, v := range extShapes {
							shapes[k] = v
						}
						continue // resolved, no warning
					}
				}
			}
			rep.Add("", CategorySchema, SeverityWarned, "external xsd:include %q not followed; affected elements get skeleton-only bodies", inc.attr("schemaLocation"))
		}
		for _, el := range schema.all("element") {
			shape := &xsdShape{NS: targetNS}
			if ct := el.get("complexType"); ct != nil {
				shape.Children = sequenceChildren(ct)
			}
			shapes["{"+targetNS+"}"+el.attr("name")] = shape
		}
		for _, ct := range schema.all("complexType") {
			if name := ct.attr("name"); name != "" {
				shapes["{"+targetNS+"}"+name] = &xsdShape{NS: targetNS, Children: sequenceChildren(ct)}
			}
		}
	}
	for _, types := range root.all("types") {
		for _, c := range types.Children {
			if c.Local == "schema" && c.Space == xsdNS {
				addSchema(c)
			}
		}
	}
	return shapes
}

func isURL(s string) bool { return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "//") }

func tryParseExternalSchema(data []byte, rep *ImportReport) map[string]*xsdShape {
	root, err := parseTree(data)
	if err != nil {
		return nil
	}
	// External file may be a standalone <xsd:schema> or a full document with single schema
	var schemas []*wsdlNode
	if root.Space == xsdNS && root.Local == "schema" {
		schemas = []*wsdlNode{root}
	} else {
		for _, c := range root.Children {
			if c.Space == xsdNS && c.Local == "schema" {
				schemas = append(schemas, c)
			}
		}
		if len(schemas) == 0 && root.Space == xsdNS && root.Local == "schema" {
			schemas = []*wsdlNode{root}
		}
	}
	if len(schemas) == 0 {
		return nil
	}
	out := map[string]*xsdShape{}
	for _, schema := range schemas {
		targetNS := schema.attr("targetNamespace")
		for _, el := range schema.all("element") {
			shape := &xsdShape{NS: targetNS}
			if ct := el.get("complexType"); ct != nil {
				shape.Children = sequenceChildren(ct)
			}
			out["{"+targetNS+"}"+el.attr("name")] = shape
		}
		for _, ct := range schema.all("complexType") {
			if name := ct.attr("name"); name != "" {
				out["{"+targetNS+"}"+name] = &xsdShape{NS: targetNS, Children: sequenceChildren(ct)}
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// sequenceChildren flattens a complexType's sequence (or choice) one level;
// nested complexTypes are expanded by renderChildren up to maxBodyDepth.
func sequenceChildren(ct *wsdlNode) []xsdChild {
	var out []xsdChild
	seq := ct.get("sequence")
	if seq == nil {
		seq = ct.get("choice")
	}
	if seq == nil {
		return nil
	}
	for _, el := range seq.all("element") {
		child := xsdChild{Name: el.attr("name"), Optional: el.attr("minOccurs") == "0"}
		if ref := el.attr("ref"); ref != "" {
			q := el.qname(ref)
			child.Name = q.Local
			child.Type = q
		} else {
			child.Type = el.qname(el.attr("type"))
		}
		out = append(out, child)
	}
	return out
}

// firstSOAPPort returns the address of the first SOAP-carrying port, its
// binding QName, and the names of any further ports that were skipped.
func firstSOAPPort(svc *wsdlNode) (string, QName, []string) {
	location := ""
	var binding QName
	var extras []string
	for _, port := range svc.all("port") {
		addr := findSoapAddress(port)
		if addr == "" {
			continue
		}
		if location == "" {
			location = addr
			binding = port.qname(port.attr("binding"))
			continue
		}
		extras = append(extras, port.attr("name"))
	}
	return location, binding, extras
}

func findSoapAddress(port *wsdlNode) string {
	for _, c := range port.Children {
		if c.Local == "address" && (c.Space == soap11BindNS || c.Space == soap12BindNS) {
			return c.attr("location")
		}
	}
	return ""
}

func buildRequests(pt *wsdlPortType, binding *wsdlBinding, address string, shapes map[string]*xsdShape, messages map[string]*wsdlMessage, rep *ImportReport) []*wsdlEntry {
	envNS, contentType := soap11Env, "text/xml; charset=utf-8"
	if binding.soap12 {
		envNS, contentType = soap12Env, "application/soap+xml; charset=utf-8"
	}
	var files []*wsdlEntry
	for _, op := range pt.operations {
		bo := binding.operations[op.name]
		if bo == nil {
			rep.Add(op.name, CategorySchema, SeverityDropped, "operation %q has no binding entry; skipped", op.name)
			continue
		}
		body := renderBody(op, bo, binding, envNS, shapes, messages, rep)
		name := op.name
		displayName := op.name
		if op.doc != "" {
			displayName = op.name + " — " + summarize(op.doc)
		}
		headers := []request.Header{
			{Key: "Content-Type", Value: contentType},
		}
		if bo.soapAction != "" {
			headers = append(headers, request.Header{Key: "SOAPAction", Value: `"` + bo.soapAction + `"`})
		}
		files = append(files, &wsdlEntry{
			Filename: name,
			File: &requestfile.File{
				Name: displayName,
				Request: request.Request{
					Name:    displayName,
					Method:  request.MethodPost,
					URL:     address,
					Headers: headers,
					Body:    body,
				},
			},
		})
	}
	return files
}

// summarize truncates long documentation to keep generated names readable.
func summarize(doc string) string {
	fields := strings.Fields(doc)
	line := strings.Join(fields, " ")
	if len(line) > 60 {
		line = line[:57] + "…"
	}
	return line
}

// renderBody builds the full envelope for one operation.
func renderBody(op *wsdlPortTypeOp, bo *bindingOp, binding *wsdlBinding, envNS string, shapes map[string]*xsdShape, messages map[string]*wsdlMessage, rep *ImportReport) string {
	var inner string
	if binding.style == "rpc" || bo.use == "encoded" {
		rep.Add(op.name, CategorySchema, SeverityWarned, "operation %q uses rpc/%s style; envelope is best-effort and may need manual editing", op.name, bo.use)
		inner = rpcWrapper(op, messages)
	} else {
		inner = literalWrapper(op, messages, shapes, rep)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "<soapenv:Envelope xmlns:soapenv=%q>\n", envNS)
	sb.WriteString("  <soapenv:Header/>\n")
	sb.WriteString("  <soapenv:Body>\n")
	sb.WriteString(indentXML(inner, 2))
	sb.WriteString("\n  </soapenv:Body>\n</soapenv:Envelope>")
	return sb.String()
}

// literalWrapper renders the document/literal input element tree.
func literalWrapper(op *wsdlPortTypeOp, messages map[string]*wsdlMessage, shapes map[string]*xsdShape, rep *ImportReport) string {
	msg := messages[op.inputMsg.String()]
	if msg == nil || len(msg.parts) == 0 {
		rep.Add(op.name, CategorySchema, SeverityWarned, "operation %q: input message %q not found; empty body used", op.name, op.inputMsg.Local)
		return "<soapenv:Fault>message not resolvable</soapenv:Fault>"
	}
	part := msg.parts[0]
	if !part.hasEl {
		rep.Add(op.name, CategorySchema, SeverityWarned, "operation %q: message part is not element-typed; skeleton-only body", op.name)
		return fmt.Sprintf("<%s/>", part.element.Local)
	}
	shape := shapes[part.element.String()]
	if shape == nil {
		rep.Add(op.name, CategorySchema, SeverityWarned, "operation %q: element %q not defined inline; skeleton-only body", op.name, part.element.Local)
		return fmt.Sprintf("<%s xmlns=%q/>", part.element.Local, part.element.NS)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "<%s xmlns=%q>", part.element.Local, shape.NS)
	renderChildren(&sb, shape, shapes, 1)
	sb.WriteString("</" + part.element.Local + ">")
	return sb.String()
}

// renderChildren writes placeholder child elements recursively.
func renderChildren(sb *strings.Builder, shape *xsdShape, shapes map[string]*xsdShape, depth int) {
	if depth > maxBodyDepth {
		return
	}
	for _, child := range shape.Children {
		if child.Optional {
			continue
		}
		if nested, ok := shapes[child.Type.String()]; ok {
			fmt.Fprintf(sb, "<%s>", child.Name)
			renderChildren(sb, nested, shapes, depth+1)
			sb.WriteString("</" + child.Name + ">")
			continue
		}
		fmt.Fprintf(sb, "<%s>%s</%s>", child.Name, scalarPlaceholder(child.Type), child.Name)
	}
}

// scalarPlaceholder yields a type-appropriate filler value.
func scalarPlaceholder(t QName) string {
	if t.NS != xsdNS {
		return ""
	}
	switch t.Local {
	case "int", "integer", "long", "short", "byte", "decimal", "double", "float",
		"nonNegativeInteger", "positiveInteger", "unsignedInt", "unsignedLong":
		return "0"
	}
	return ""
}

// rpcWrapper renders an operation-named wrapper with message parts as
// children — best-effort for rpc/encoded services.
func rpcWrapper(op *wsdlPortTypeOp, messages map[string]*wsdlMessage) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "<%s>", op.name)
	if msg := messages[op.inputMsg.String()]; msg != nil {
		for _, part := range msg.parts {
			fmt.Fprintf(&sb, "<%s>%s</%s>", part.name, scalarPlaceholder(part.typ), part.name)
		}
	}
	sb.WriteString("</" + op.name + ">")
	return sb.String()
}

func indentXML(s string, level int) string {
	prefix := strings.Repeat("  ", level)
	lines := strings.Split(s, "\n")
	for i := range lines {
		if lines[i] != "" {
			lines[i] = prefix + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
