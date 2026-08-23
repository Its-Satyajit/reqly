package importer

import (
	"fmt"
	"regexp"
	"strings"
)

// ScriptDialect identifies the source scripting API of an imported script.
type ScriptDialect string

const (
	DialectPostman  ScriptDialect = "postman"
	DialectBruno    ScriptDialect = "bruno"
	DialectInsomnia ScriptDialect = "insomnia"
)

const todoMarker = "// TODO(reqly-import): "

// foreignRef matches leftover references to any foreign scripting API.
var foreignRef = regexp.MustCompile(`\b(pm|bru|insomnia)\.`)

// mappingRow rewrites one API pattern onto the reqly sandbox surface.
type mappingRow struct {
	re *regexp.Regexp
	to string
}

// postmanMappings translate the core Postman scripting API. Order matters:
// `.has(` must be rewritten before the generic `.get(` row.
var postmanMappings = []mappingRow{
	{regexp.MustCompile(`pm\.test\(`), "reqly.test("},
	{regexp.MustCompile(`pm\.response\.code\b`), "reqly.response.status"},
	{regexp.MustCompile(`pm\.response\.statusText\b`), "reqly.response.statusText"},
	{regexp.MustCompile(`pm\.response\.json\(\)`), "JSON.parse(reqly.response.body)"},
	{regexp.MustCompile(`pm\.response\.text\(\)`), "String(reqly.response.body)"},
	{regexp.MustCompile(`pm\.response\.headers\.get\(`), "(reqly.response.headers)["},
	{regexp.MustCompile(`pm\.request\.url\b`), "reqly.request.url"},
	{regexp.MustCompile(`pm\.request\.method\b`), "reqly.request.method"},
	{regexp.MustCompile(`pm\.(?:environment|collectionVariables|variables|globals)\.has\(`), "reqly.hasVariable("},
	{regexp.MustCompile(`pm\.(?:environment|collectionVariables|variables|globals)\.get\(`), "reqly.getVariable("},
	{regexp.MustCompile(`pm\.(?:environment|collectionVariables|variables|globals)\.set\(`), "reqly.setVariable("},
}

// brunoMappings translate the Bruno scripting API.
var brunoMappings = []mappingRow{
	{regexp.MustCompile(`^\s*test\(`), "reqly.test("},
	{regexp.MustCompile(`bru\.getEnvVar\(|bru\.getVar\(`), "reqly.getVariable("},
	{regexp.MustCompile(`bru\.setEnvVar\(|bru\.setVar\(`), "reqly.setVariable("},
	{regexp.MustCompile(`bru\.hasEnvVar\(|bru\.hasVar\(`), "reqly.hasVariable("},
}

// insomniaMappings translate the Insomnia plugin-style API.
var insomniaMappings = []mappingRow{
	{regexp.MustCompile(`insomnia\.environment\.has\(`), "reqly.hasVariable("},
	{regexp.MustCompile(`insomnia\.environment\.get\(`), "reqly.getVariable("},
	{regexp.MustCompile(`insomnia\.environment\.set\(`), "reqly.setVariable("},
}

func mappingsFor(d ScriptDialect) []mappingRow {
	switch d {
	case DialectPostman:
		return postmanMappings
	case DialectBruno:
		return brunoMappings
	case DialectInsomnia:
		return insomniaMappings
	default:
		return nil
	}
}

// expectStart detects assertion-library entry points that are deliberately not
// emulated (ADR 0026).
var expectStart = regexp.MustCompile(`\b(?:pm\.expect|expect|chai)\b|\brequire\(\s*['"]chai['"]\s*\)`)

// TranslateScript rewrites a foreign-API script onto Reqly's `reqly.*` sandbox
// surface, one-shot at import time (ADR 0026). Lines that map cleanly are
// translated; unmappable lines are preserved verbatim as TODO(reqly-import)
// comments — never deleted. Each preserved line or multi-line block produces a
// script-category report entry so degradations reach the ImportReport.
func TranslateScript(source string, dialect ScriptDialect) (string, []ReportEntry) {
	if strings.TrimSpace(source) == "" {
		return "", nil
	}
	rows := mappingsFor(dialect)
	var out []string
	var entries []ReportEntry

	commenting := false // inside an unbalanced commented construct
	balance := 0
	for i, line := range strings.Split(source, "\n") {
		num := i + 1
		if commenting {
			out = append(out, todoMarker+line)
			balance += parenDelta(line)
			if balance <= 0 {
				commenting = false
			}
			continue
		}
		if expectStart.MatchString(line) {
			out = append(out, todoMarker+line)
			entries = append(entries, ReportEntry{
				Category: CategoryScript,
				Severity: SeverityWarned,
				Message:  fmt.Sprintf("line %d: assertion library call is not portable; preserved as TODO(reqly-import) comment", num),
			})
			d := parenDelta(line)
			if d > 0 {
				commenting = true
				balance = d
			}
			continue
		}
		for _, m := range rows {
			line = m.re.ReplaceAllString(line, m.to)
		}
		if foreignRef.MatchString(line) {
			out = append(out, todoMarker+line)
			entries = append(entries, ReportEntry{
				Category: CategoryScript,
				Severity: SeverityWarned,
				Message:  fmt.Sprintf("line %d: unsupported scripting API call; preserved as TODO(reqly-import) comment", num),
			})
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n"), entries
}

// parenDelta counts net parentheses on a line, ignoring those in string and
// character literals well enough for generated test scripts.
func parenDelta(line string) int {
	depth := 0
	inStr := byte(0)
	for i := 0; i < len(line); i++ {
		c := line[i]
		if inStr != 0 {
			if c == '\\' {
				i++
			} else if c == inStr {
				inStr = 0
			}
			continue
		}
		switch c {
		case '\'', '"', '`':
			inStr = c
		case '(':
			depth++
		case ')':
			depth--
		}
	}
	return depth
}
