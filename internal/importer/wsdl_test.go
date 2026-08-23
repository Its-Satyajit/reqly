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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

const simpleWSDL = `<?xml version="1.0" encoding="UTF-8"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/"
  xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
  xmlns:tns="http://example.com/test"
  xmlns:xsd="http://www.w3.org/2001/XMLSchema"
  name="TestService" targetNamespace="http://example.com/test">
  <wsdl:types>
    <xsd:schema targetNamespace="http://example.com/test" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
      <xsd:element name="SayHelloRequest">
        <xsd:complexType>
          <xsd:sequence>
            <xsd:element name="name" type="xsd:string"/>
            <xsd:element name="age" type="xsd:int"/>
            <xsd:element name="nickname" type="xsd:string" minOccurs="0"/>
          </xsd:sequence>
        </xsd:complexType>
      </xsd:element>
    </xsd:schema>
  </wsdl:types>
  <wsdl:message name="SayHelloRequestMessage">
    <wsdl:part name="parameters" element="tns:SayHelloRequest"/>
  </wsdl:message>
  <wsdl:portType name="TestPortType">
    <wsdl:operation name="SayHello">
      <wsdl:documentation>Say hello to someone</wsdl:documentation>
      <wsdl:input message="tns:SayHelloRequestMessage"/>
    </wsdl:operation>
  </wsdl:portType>
  <wsdl:binding name="TestBinding" type="tns:TestPortType">
    <soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
    <wsdl:operation name="SayHello">
      <soap:operation soapAction="http://example.com/test/SayHello"/>
      <wsdl:input><soap:body use="literal"/></wsdl:input>
    </wsdl:operation>
  </wsdl:binding>
  <wsdl:service name="TestSvc">
    <wsdl:port name="TestPort" binding="tns:TestBinding">
      <soap:address location="http://example.com/soap"/>
    </wsdl:port>
  </wsdl:service>
</wsdl:definitions>
`

// firstReq returns the first generated request of the first collection.
func firstReq(res *WSDLResult) *requestfile.File { return res.Collections[0].Request[0].File }

func mustParseWSDL(t *testing.T, data string) *WSDLResult {
	t.Helper()
	res, warnings, err := ParseWSDL([]byte(data))
	if err != nil {
		t.Fatalf("ParseWSDL() error = %v", err)
	}
	t.Cleanup(func() {})
	_ = warnings
	return res
}

func TestWSDLDiscovery(t *testing.T) {
	res := mustParseWSDL(t, simpleWSDL)
	if len(res.Collections) != 1 || res.Collections[0].Name != "TestSvc" {
		t.Fatalf("collections = %+v, want one named TestSvc", res.Collections)
	}
	reqs := res.Collections[0].Request
	if len(reqs) != 1 {
		t.Fatalf("got %d requests, want 1", len(reqs))
	}
	f := reqs[0].File
	if f.Request.Method != "POST" || f.Request.URL != "http://example.com/soap" {
		t.Errorf("method/url = %s %s", f.Request.Method, f.Request.URL)
	}
	hasCT, hasAction := false, false
	for _, h := range f.Request.Headers {
		if h.Key == "Content-Type" && h.Value == "text/xml; charset=utf-8" {
			hasCT = true
		}
		if h.Key == "SOAPAction" && h.Value == `"http://example.com/test/SayHello"` {
			hasAction = true
		}
	}
	if !hasCT || !hasAction {
		t.Errorf("headers missing Content-Type/SOAPAction: %+v", f.Request.Headers)
	}
}

func TestWSDLEnvelopeSkeleton(t *testing.T) {
	res := mustParseWSDL(t, simpleWSDL)
	body := firstReq(res).Request.Body
	for _, want := range []string{
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">`,
		"<soapenv:Header/>",
		"<soapenv:Body>",
		`<SayHelloRequest xmlns="http://example.com/test">`,
		"<name></name>",
		"<age>0</age>",
		"</soapenv:Envelope>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("envelope missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "nickname") {
		t.Errorf("optional minOccurs=0 element should be omitted:\n%s", body)
	}
}

func TestWSDLOperationNamingAndDocs(t *testing.T) {
	res := mustParseWSDL(t, simpleWSDL)
	if got := firstReq(res).Name; got != "SayHello — Say hello to someone" {
		t.Errorf("name = %q, want documentation carried over", got)
	}
}

func TestWSDLSOAP12Binding(t *testing.T) {
	soap12 := strings.NewReplacer(
		`xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"`,
		`xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap12/"`,
	).Replace(simpleWSDL)
	res := mustParseWSDL(t, soap12)
	f := firstReq(res)
	if !strings.Contains(f.Request.Body, "http://www.w3.org/2003/05/soap-envelope") {
		t.Errorf("SOAP 1.2 envelope namespace missing:\n%s", f.Request.Body)
	}
	for _, h := range f.Request.Headers {
		if h.Key == "Content-Type" && h.Value != "application/soap+xml; charset=utf-8" {
			t.Errorf("Content-Type = %q for SOAP 1.2", h.Value)
		}
	}
}

func TestWSDLExternalImportWarns(t *testing.T) {
	withImport := strings.Replace(simpleWSDL,
		`<xsd:schema targetNamespace="http://example.com/test" xmlns:xsd="http://www.w3.org/2001/XMLSchema">`,
		`<xsd:schema targetNamespace="http://example.com/test" xmlns:xsd="http://www.w3.org/2001/XMLSchema"><xsd:import namespace="http://other" schemaLocation="other.xsd"/>`, 1)
	res, warnings, err := ParseWSDL([]byte(withImport))
	if err != nil {
		t.Fatalf("ParseWSDL() error = %v", err)
	}
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "import") {
		t.Errorf("warnings missing import notice: %v", warnings)
	}
	body := firstReq(res).Request.Body
	if !strings.Contains(body, "<SayHelloRequest") {
		t.Error("root element should still be generated despite skipped import")
	}
}

func TestWSDLRPCStyleWarns(t *testing.T) {
	rpc := `<?xml version="1.0"?>
<wsdl:definitions xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/"
  xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
  xmlns:tns="http://example.com/rpc" xmlns:xsd="http://www.w3.org/2001/XMLSchema"
  name="RpcService" targetNamespace="http://example.com/rpc">
  <wsdl:message name="AddMessage">
    <wsdl:part name="a" type="xsd:int"/>
    <wsdl:part name="b" type="xsd:int"/>
  </wsdl:message>
  <wsdl:portType name="RpcPT">
    <wsdl:operation name="Add"><wsdl:input message="tns:AddMessage"/></wsdl:operation>
  </wsdl:portType>
  <wsdl:binding name="RpcB" type="tns:RpcPT">
    <soap:binding style="rpc" transport="http://schemas.xmlsoap.org/soap/http"/>
    <wsdl:operation name="Add">
      <soap:operation soapAction=""/>
      <wsdl:input><soap:body use="encoded"/></wsdl:input>
    </wsdl:operation>
  </wsdl:binding>
  <wsdl:service name="RpcSvc">
    <wsdl:port name="RpcPort" binding="tns:RpcB">
      <soap:address location="http://example.com/rpc"/>
    </wsdl:port>
  </wsdl:service>
</wsdl:definitions>`
	res, warnings, err := ParseWSDL([]byte(rpc))
	if err != nil {
		t.Fatalf("ParseWSDL() error = %v", err)
	}
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "encoded") {
		t.Errorf("warnings should mention encoded style: %v", warnings)
	}
	body := firstReq(res).Request.Body
	if !strings.Contains(body, "<Add>") || !strings.Contains(body, "<a>0</a>") {
		t.Errorf("rpc wrapper with part children expected:\n%s", body)
	}
}

func TestWSDLExtraPortsWarn(t *testing.T) {
	extra := strings.Replace(simpleWSDL,
		`</wsdl:service>`,
		`<wsdl:port name="BackupPort" binding="tns:TestBinding"><soap:address location="http://backup.example.com/soap"/></wsdl:port></wsdl:service>`, 1)
	res, warnings, err := ParseWSDL([]byte(extra))
	if err != nil {
		t.Fatalf("ParseWSDL() error = %v", err)
	}
	if len(res.Collections[0].Request) != 1 {
		t.Fatalf("expected still 1 request, got %d", len(res.Collections[0].Request))
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "BackupPort") {
		t.Errorf("warnings should list extra port: %v", warnings)
	}
	if firstReq(res).Request.URL != "http://example.com/soap" {
		t.Error("first port's address should win")
	}
}

func TestWSDLWriteCollisionSuffixes(t *testing.T) {
	dir := t.TempDir()
	res := mustParseWSDL(t, simpleWSDL)
	res.Collections[0].Request = append(res.Collections[0].Request,
		&wsdlEntry{Filename: res.Collections[0].Request[0].Filename, File: res.Collections[0].Request[0].File})
	if err := res.Write(dir); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "collections", "TestSvc", "SayHello.yaml")); err != nil {
		t.Errorf("first file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "collections", "TestSvc", "SayHello-2.yaml")); err != nil {
		t.Errorf("collision-suffixed file missing: %v", err)
	}
}

func TestWSDLWriteWorkspace(t *testing.T) {
	dir := t.TempDir()
	res := mustParseWSDL(t, simpleWSDL)
	if err := res.Write(dir); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	wsData, err := os.ReadFile(filepath.Join(dir, "reqly.yaml"))
	if err != nil || !strings.Contains(string(wsData), "TestService") {
		t.Fatalf("workspace descriptor missing/bad: %v", err)
	}
	rf, err := requestfile.LoadFile(filepath.Join(dir, "collections", "TestSvc", "SayHello.yaml"))
	if err != nil {
		t.Fatalf("generated request file does not parse: %v", err)
	}
	if rf.Request.URL != "http://example.com/soap" {
		t.Errorf("round-trip url = %q", rf.Request.URL)
	}
}

func TestWSDLRejectsNonWSDL(t *testing.T) {
	if _, _, err := ParseWSDL([]byte(`<html><body/></html>`)); err == nil {
		t.Fatal("expected error for non-WSDL document")
	}
}
