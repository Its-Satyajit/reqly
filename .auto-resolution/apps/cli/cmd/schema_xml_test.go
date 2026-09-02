// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSchemaValidateCmd_XSD(t *testing.T) {
	resetSchemaFlags()
	dir := t.TempDir()
	xsdPath := filepath.Join(dir, "note.xsd")
	xmlPath := filepath.Join(dir, "note.xml")

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

	xmlContent := `<?xml version="1.0"?><note><to>Tove</to></note>`

	if err := os.WriteFile(xsdPath, []byte(xsdContent), 0644); err != nil {
		t.Fatalf("failed to write xsd: %v", err)
	}
	if err := os.WriteFile(xmlPath, []byte(xmlContent), 0644); err != nil {
		t.Fatalf("failed to write xml: %v", err)
	}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	rootCmd.SetArgs([]string{"schema", "validate", xsdPath, xmlPath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected failure: %v", err)
	}
}
