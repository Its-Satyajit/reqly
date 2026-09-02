// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"testing"
)

func TestValidateXMLAgainstXSD_Valid(t *testing.T) {
	xsd := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
    <xs:element name="note">
        <xs:complexType>
            <xs:sequence>
                <xs:element name="to" type="xs:string"/>
                <xs:element name="from" type="xs:string"/>
                <xs:element name="heading" type="xs:string"/>
                <xs:element name="body" type="xs:string"/>
            </xs:sequence>
            <xs:attribute name="id" type="xs:string" use="required"/>
        </xs:complexType>
    </xs:element>
</xs:schema>`)

	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<note id="101">
    <to>Tove</to>
    <from>Jani</from>
    <heading>Reminder</heading>
    <body>Don't forget me this weekend!</body>
</note>`)

	res, err := ValidateXMLAgainstXSD(xml, xsd, ValidationOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Valid {
		t.Errorf("expected valid XML instance, got errors: %v", res.Errors)
	}
}

func TestValidateXMLAgainstXSD_MissingAttribute(t *testing.T) {
	xsd := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
    <xs:element name="note">
        <xs:complexType>
            <xs:sequence>
                <xs:element name="to" type="xs:string"/>
            </xs:sequence>
            <xs:attribute name="id" type="xs:string" use="required"/>
        </xs:complexType>
    </xs:element>
</xs:schema>`)

	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<note>
    <to>Tove</to>
</note>`)

	res, err := ValidateXMLAgainstXSD(xml, xsd, ValidationOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid {
		t.Fatalf("expected validation failure due to missing required attribute 'id'")
	}
	if len(res.Errors) == 0 {
		t.Fatalf("expected error details in result")
	}
}

func TestValidateXMLAgainstXSD_MissingRequiredElement(t *testing.T) {
	xsd := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
    <xs:element name="user">
        <xs:complexType>
            <xs:sequence>
                <xs:element name="name" type="xs:string"/>
                <xs:element name="email" type="xs:string"/>
            </xs:sequence>
        </xs:complexType>
    </xs:element>
</xs:schema>`)

	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<user>
    <name>Satyajit</name>
</user>`)

	res, err := ValidateXMLAgainstXSD(xml, xsd, ValidationOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid {
		t.Fatalf("expected validation failure due to missing required element 'email'")
	}
}

func TestValidateXMLAgainstXSD_MismatchedRoot(t *testing.T) {
	xsd := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
    <xs:element name="user">
        <xs:complexType>
            <xs:sequence>
                <xs:element name="name" type="xs:string"/>
            </xs:sequence>
        </xs:complexType>
    </xs:element>
</xs:schema>`)

	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<person>
    <name>Satyajit</name>
</person>`)

	res, err := ValidateXMLAgainstXSD(xml, xsd, ValidationOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid {
		t.Fatalf("expected validation failure due to root tag mismatch ('person' vs 'user')")
	}
}
