package diffing

import (
	"strings"
)

// Severity classes for OpenAPI changes surfaced by the diff views.
const (
	// SeverityBreaking marks changes that can break existing clients.
	SeverityBreaking = "breaking"
	// SeverityNonBreaking marks purely additive changes.
	SeverityNonBreaking = "addition"
	// SeverityInfo marks modifications that are usually safe.
	SeverityInfo = "info"
)

// breakingKeywords are schema/document keys whose removal or modification
// changes the wire contract clients were built against.
var breakingKeywords = map[string]bool{
	"required":    true,
	"enum":        true,
	"operationId": true,
	"requestBody": true,
	"parameters":  true,
	"servers":     true,
	"security":    true,
	"responses":   true,
}

// WithSeverity returns a copy of the result with every change classified for
// an OpenAPI diff: deletions under paths are breaking, additions are marked,
// updates break when they touch contract-bearing keywords.
func WithSeverity(r *DiffResult) *DiffResult {
	if r == nil {
		return nil
	}
	out := &DiffResult{HasChanges: r.HasChanges}
	for _, c := range r.Changes {
		c.Severity = classify(c)
		out.Changes = append(out.Changes, c)
	}
	return out
}

func classify(c Change) string {
	joined := joinPath(c.Path)
	inPaths := strings.HasPrefix(joined, "paths.")
	switch c.Type {
	case "delete":
		if inPaths {
			return SeverityBreaking
		}
		// Deleting metadata (descriptions, examples) is cosmetic.
		if mentionsAny(c.Path, "description", "example", "summary", "tags", "title") {
			return SeverityInfo
		}
		return SeverityBreaking
	case "create":
		return SeverityNonBreaking
	default: // update
		last := lastSegment(c.Path)
		if breakingKeywords[last] {
			return SeverityBreaking
		}
		if inPaths && typeChanged(c) {
			return SeverityBreaking
		}
		return SeverityInfo
	}
}

func typeChanged(c Change) bool {
	return lastSegment(c.Path) == "type" && valueString(c.From) != valueString(c.To)
}

func valueString(v any) string {
	s, _ := v.(string)
	return s
}

func mentionsAny(path []string, keys ...string) bool {
	for _, seg := range path {
		for _, k := range keys {
			if seg == k {
				return true
			}
		}
	}
	return false
}

func lastSegment(path []string) string {
	if len(path) == 0 {
		return ""
	}
	return path[len(path)-1]
}

func joinPath(path []string) string { return strings.Join(path, ".") }
