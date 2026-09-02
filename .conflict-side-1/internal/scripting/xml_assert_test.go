// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package scripting

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSandbox_AssertXSD(t *testing.T) {
	dir := t.TempDir()
	xsdPath := filepath.Join(dir, "note.xsd")
	xsdContent := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
    <xs:element name="note">
        <xs:complexType>
            <xs:sequence>
                <xs:element name="to" type="xs:string"/>
            </xs:sequence>
        </xs:complexType>
    </xs:element>
</xs:schema>`

	if err := os.WriteFile(xsdPath, []byte(xsdContent), 0644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	sb := NewSandbox(SandboxOptions{})
	sb.BindResponse(&responseView{
		Body: `<?xml version="1.0"?><note><to>Tove</to></note>`,
	})

	script := `reqly.test("valid xml", function() { return reqly.assertXSD("` + xsdPath + `"); });`
	if err := sb.Run(script); err != nil {
		t.Fatalf("script run failed: %v", err)
	}

	tests := sb.Tests()
	if len(tests) != 1 {
		t.Fatalf("expected 1 test, got %d", len(tests))
	}
	if !tests[0].Fn() {
		t.Errorf("expected test to pass for valid xml")
	}
}
