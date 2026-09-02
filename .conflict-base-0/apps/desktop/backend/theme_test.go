// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
)

func TestThemeList_Desktop(t *testing.T) {
	svc := NewAppService()
	themes := svc.ThemeList()
	if len(themes) != 14 {
		t.Fatalf("want 14 themes, got %d", len(themes))
	}
}

func TestThemeExport_Desktop(t *testing.T) {
	svc := NewAppService()
	yamlStr, err := svc.ThemeExport("atlas-light")
	if err != nil {
		t.Fatalf("ThemeExport: %v", err)
	}
	if !strings.Contains(yamlStr, "atlas-light") {
		t.Fatalf("unexpected yaml %s", yamlStr)
	}
}

func TestThemeImport_Desktop(t *testing.T) {
	svc := NewAppService()
	yamlStr := "id: desk-custom\nlabel: Desk\nappearance: dark\ntokens:\n  primary: \"#abcdef\"\n"
	css, err := svc.ThemeImport(yamlStr)
	if err != nil {
		t.Fatalf("ThemeImport: %v", err)
	}
	if !strings.Contains(css, "desk-custom") {
		t.Fatalf("unexpected css %s", css)
	}
}

func TestThemeImport_DesktopInvalid(t *testing.T) {
	svc := NewAppService()
	if _, err := svc.ThemeImport("not: [valid"); err == nil {
		t.Fatalf("expected error for invalid")
	}
}
