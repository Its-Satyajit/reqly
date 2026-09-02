package importer

import (
	"strings"
	"testing"
)

func TestImportReportMessages(t *testing.T) {
	rep := NewReport("postman")
	rep.Add("Create order", CategoryAuth, SeverityWarned, "%s: hawk auth not supported", "Create order")
	rep.AddAll("Upload file", CategoryBody, SeverityWarned, []string{"file body skipped"})
	if got := len(rep.Entries); got != 2 {
		t.Fatalf("want 2 entries, got %d", got)
	}
	msgs := rep.Messages()
	want := []string{"Create order: hawk auth not supported", "file body skipped"}
	for i, w := range want {
		if msgs[i] != w {
			t.Errorf("Messages()[%d] = %q, want %q", i, msgs[i], w)
		}
	}
}

func TestImportReportTally(t *testing.T) {
	rep := NewReport("bruno")
	rep.Add("", CategorySchema, SeverityTranslated, "a")
	rep.Add("", CategoryBody, SeverityWarned, "b")
	rep.Add("", CategoryOther, SeverityDropped, "c")
	tw, w, d := rep.Tally()
	if tw != 1 || w != 1 || d != 1 {
		t.Fatalf("tally = (%d,%d,%d), want (1,1,1)", tw, w, d)
	}
}

func TestImportReportStringGroupsByCategory(t *testing.T) {
	rep := NewReport("wsdl")
	rep.Add("BackupPort", CategoryOther, SeverityDropped, "extra port listed")
	rep.Add("op", CategoryScript, SeverityDropped, "script dropped")
	rep.Add("", CategorySchema, SeverityWarned, "external xsd not followed")
	out := rep.String()
	order := []string{
		"script:",
		"script dropped",
		"schema:",
		"external xsd not followed",
		"other:",
		"BackupPort: extra port listed",
		"0 translated, 1 warned, 2 dropped",
	}
	last := -1
	for _, want := range order {
		i := strings.Index(out, want)
		if i < 0 {
			t.Errorf("String() missing %q:\n%s", want, out)
			continue
		}
		if i < last {
			t.Errorf("String() ordering: %q appears before previous match\n%s", want, out)
		}
		last = i
	}
}

func TestImportReportStringEmpty(t *testing.T) {
	if s := NewReport("har").String(); s != "" {
		t.Errorf("empty report String() = %q, want \"\"", s)
	}
	var nilRep *ImportReport
	if s := nilRep.String(); s != "" {
		t.Errorf("nil report String() = %q, want \"\"", s)
	}
}
