package importer

import (
	"fmt"
	"strings"
)

// Category classifies which part of an import a report entry concerns.
type Category string

const (
	CategoryAuth        Category = "auth"
	CategoryScript      Category = "script"
	CategoryBody        Category = "body"
	CategoryEnvironment Category = "environment"
	CategorySchema      Category = "schema"
	CategoryOther       Category = "other"
)

// Severity records what happened to the source feature: it was mapped onto
// Reqly-native form, degraded but imported, or skipped entirely.
type Severity string

const (
	SeverityTranslated Severity = "translated"
	SeverityWarned     Severity = "warned"
	SeverityDropped    Severity = "dropped"
)

// ReportEntry is one structured degradation, skip, or translation record.
// ItemPath locates the source item (request/folder/environment name or
// positional label such as "entry 3"); Message carries the same human-readable
// text the free-text warning strings used before reports existed.
type ReportEntry struct {
	ItemPath string
	Category Category
	Severity Severity
	Message  string
}

// ImportReport is the structured record of how an import degraded its source.
// Every parser returns one in place of the former []string warnings; rendering
// belongs to callers (the CLI groups by category, desktop renders later).
type ImportReport struct {
	Importer string
	Entries  []ReportEntry
}

// NewReport returns an empty report attributed to the named importer.
func NewReport(importerName string) *ImportReport {
	return &ImportReport{Importer: importerName}
}

// Add appends one entry with a formatted message.
func (r *ImportReport) Add(itemPath string, cat Category, sev Severity, format string, args ...any) {
	r.Entries = append(r.Entries, ReportEntry{
		ItemPath: itemPath,
		Category: cat,
		Severity: sev,
		Message:  fmt.Sprintf(format, args...),
	})
}

// AddAll records pre-rendered messages (returned by helper functions) under a
// single category and severity. Messages are kept verbatim so existing text
// stays recognizable.
func (r *ImportReport) AddAll(itemPath string, cat Category, sev Severity, msgs []string) {
	for _, m := range msgs {
		r.Entries = append(r.Entries, ReportEntry{ItemPath: itemPath, Category: cat, Severity: sev, Message: m})
	}
}

// Messages returns the raw message of every entry in order — the free-text
// view callers printed before reports existed.
func (r *ImportReport) Messages() []string {
	out := make([]string, len(r.Entries))
	for i, e := range r.Entries {
		out[i] = e.Message
	}
	return out
}

// Tally counts entries per severity for summary rendering.
func (r *ImportReport) Tally() (translated, warned, dropped int) {
	for _, e := range r.Entries {
		switch e.Severity {
		case SeverityTranslated:
			translated++
		case SeverityWarned:
			warned++
		default:
			dropped++
		}
	}
	return translated, warned, dropped
}

// String renders the report as grouped plain text: entries ordered by category
// with item paths where present, followed by a severity tally line.
func (r *ImportReport) String() string {
	if r == nil || len(r.Entries) == 0 {
		return ""
	}
	order := []Category{CategoryAuth, CategoryScript, CategoryBody, CategoryEnvironment, CategorySchema, CategoryOther}
	var sb strings.Builder
	for _, cat := range order {
		first := true
		for _, e := range r.Entries {
			if e.Category != cat {
				continue
			}
			if first {
				fmt.Fprintf(&sb, "%s:\n", cat)
				first = false
			}
			line := e.Message
			if e.ItemPath != "" && !strings.Contains(line, e.ItemPath) {
				line = fmt.Sprintf("%s: %s", e.ItemPath, line)
			}
			fmt.Fprintf(&sb, "  %s\n", line)
		}
	}
	t, w, d := r.Tally()
	fmt.Fprintf(&sb, "%d translated, %d warned, %d dropped\n", t, w, d)
	return sb.String()
}
