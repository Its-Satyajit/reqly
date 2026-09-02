// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ValidationOptions configures XML/XSD validation execution.
type ValidationOptions struct {
	WorkspaceDir string `json:"workspaceDir,omitempty"`
}

// XSDValidationError describes an individual XML vs XSD validation error.
type XSDValidationError struct {
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Path     string `json:"path"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "error" | "warning"
}

// XSDValidationResult contains full status, errors, and warnings for validation.
type XSDValidationResult struct {
	Valid    bool                 `json:"valid"`
	Errors   []XSDValidationError `json:"errors,omitempty"`
	Warnings []string             `json:"warnings,omitempty"`
}

// Simple XML DOM representation for validation
type xmlNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr
	Children []xmlNode
	Content  string
}

func (n *xmlNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.XMLName = start.Name
	n.Attrs = start.Attr

	for {
		token, err := d.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			var child xmlNode
			if err := child.UnmarshalXML(d, t); err != nil {
				return err
			}
			n.Children = append(n.Children, child)
		case xml.CharData:
			n.Content += string(t)
		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
}

// Schema definitions
type xsdSchema struct {
	XMLName  xml.Name     `xml:"schema"`
	Elements []xsdElement `xml:"element"`
	Imports  []xsdImport  `xml:"import"`
	Includes []xsdInclude `xml:"include"`
}

type xsdImport struct {
	SchemaLocation string `xml:"schemaLocation,attr"`
	Namespace      string `xml:"namespace,attr"`
}

type xsdInclude struct {
	SchemaLocation string `xml:"schemaLocation,attr"`
}

type xsdElement struct {
	Name        string          `xml:"name,attr"`
	Type        string          `xml:"type,attr"`
	MinOccurs   string          `xml:"minOccurs,attr"`
	MaxOccurs   string          `xml:"maxOccurs,attr"`
	ComplexType *xsdComplexType `xml:"complexType"`
}

type xsdComplexType struct {
	Sequence   *xsdSequence   `xml:"sequence"`
	Attributes []xsdAttribute `xml:"attribute"`
}

type xsdSequence struct {
	Elements []xsdElement `xml:"element"`
}

type xsdAttribute struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	Use  string `xml:"use,attr"`
}

// ValidateXMLAgainstXSD validates an XML byte slice against an XSD schema byte slice.
func ValidateXMLAgainstXSD(xmlContent, xsdContent []byte, opts ValidationOptions) (*XSDValidationResult, error) {
	var schema xsdSchema
	if err := xml.Unmarshal(xsdContent, &schema); err != nil {
		return nil, fmt.Errorf("invalid xsd schema: %w", err)
	}

	var warnings []string
	for _, imp := range append(schema.Imports, xsdImport{}) {
		if strings.HasPrefix(imp.SchemaLocation, "http://") || strings.HasPrefix(imp.SchemaLocation, "https://") {
			warnings = append(warnings, fmt.Sprintf("Warning: Remote XSD import '%s' skipped for offline safety. Save schema locally into workspace to resolve.", imp.SchemaLocation))
		}
	}

	var rootNode xmlNode
	decoder := xml.NewDecoder(bytes.NewReader(xmlContent))
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("invalid xml content: %w", err)
		}
		if se, ok := token.(xml.StartElement); ok {
			if err := rootNode.UnmarshalXML(decoder, se); err != nil {
				return nil, fmt.Errorf("error parsing xml dom: %w", err)
			}
			break
		}
	}

	var errs []XSDValidationError

	// Match root element definition
	var targetElem *xsdElement
	for i := range schema.Elements {
		if schema.Elements[i].Name == rootNode.XMLName.Local {
			targetElem = &schema.Elements[i]
			break
		}
	}

	if targetElem == nil {
		errs = append(errs, XSDValidationError{
			Path:     rootNode.XMLName.Local,
			Message:  fmt.Sprintf("root element '%s' not defined in XSD schema", rootNode.XMLName.Local),
			Severity: "error",
		})
		return &XSDValidationResult{Valid: false, Errors: errs, Warnings: warnings}, nil
	}

	// Validate attributes and complex sequence
	if targetElem.ComplexType != nil {
		// Check attributes
		attrMap := make(map[string]string)
		for _, attr := range rootNode.Attrs {
			attrMap[attr.Name.Local] = attr.Value
		}

		for _, reqAttr := range targetElem.ComplexType.Attributes {
			if reqAttr.Use == "required" {
				if _, ok := attrMap[reqAttr.Name]; !ok {
					errs = append(errs, XSDValidationError{
						Path:     fmt.Sprintf("%s/@%s", rootNode.XMLName.Local, reqAttr.Name),
						Message:  fmt.Sprintf("missing required attribute '%s'", reqAttr.Name),
						Severity: "error",
					})
				}
			}
		}

		// Check sequence elements
		if targetElem.ComplexType.Sequence != nil {
			childMap := make(map[string]int)
			for _, child := range rootNode.Children {
				childMap[child.XMLName.Local]++
			}

			for _, seqElem := range targetElem.ComplexType.Sequence.Elements {
				minOccurs := 1
				if seqElem.MinOccurs == "0" {
					minOccurs = 0
				}
				count := childMap[seqElem.Name]
				if count < minOccurs {
					errs = append(errs, XSDValidationError{
						Path:     fmt.Sprintf("%s/%s", rootNode.XMLName.Local, seqElem.Name),
						Message:  fmt.Sprintf("missing required element '%s'", seqElem.Name),
						Severity: "error",
					})
				}
			}
		}
	}

	return &XSDValidationResult{
		Valid:    len(errs) == 0,
		Errors:   errs,
		Warnings: warnings,
	}, nil
}
