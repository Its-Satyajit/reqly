package diffing

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ChangelogItem describes a single categorized change.
type ChangelogItem struct {
	Type     string `json:"type"`     // delete, create, update
	Path     string `json:"path"`     // dot-separated path
	Summary  string `json:"summary"`  // human-readable message
	Severity string `json:"severity"` // breaking, addition, info
}

// Changelog represents categorized changes and suggested SemVer increment.
type Changelog struct {
	SuggestedSemver string          `json:"suggested_semver"` // major, minor, patch, none
	Breaking        []ChangelogItem `json:"breaking"`
	Additions       []ChangelogItem `json:"additions"`
	Info            []ChangelogItem `json:"info"`
}

// GenerateChangelog parses two JSON API specs and returns a structured Changelog.
func GenerateChangelog(oldBytes, newBytes []byte) (*Changelog, error) {
	diff, err := JSON(oldBytes, newBytes)
	if err != nil {
		return nil, fmt.Errorf("diff: %w", err)
	}

	classified := WithSeverity(diff)
	cl := &Changelog{
		SuggestedSemver: "none",
		Breaking:        []ChangelogItem{},
		Additions:       []ChangelogItem{},
		Info:            []ChangelogItem{},
	}

	if classified == nil || len(classified.Changes) == 0 {
		return cl, nil
	}

	for _, c := range classified.Changes {
		pathStr := strings.Join(c.Path, ".")
		item := ChangelogItem{
			Type:     c.Type,
			Path:     pathStr,
			Severity: c.Severity,
		}

		switch c.Type {
		case "create":
			item.Summary = fmt.Sprintf("Added `%s`", pathStr)
		case "delete":
			item.Summary = fmt.Sprintf("Removed `%s`", pathStr)
		default: // update
			item.Summary = fmt.Sprintf("Modified `%s` (%v -> %v)", pathStr, c.From, c.To)
		}

		switch c.Severity {
		case SeverityBreaking:
			cl.Breaking = append(cl.Breaking, item)
		case SeverityNonBreaking:
			cl.Additions = append(cl.Additions, item)
		default:
			cl.Info = append(cl.Info, item)
		}
	}

	if len(cl.Breaking) > 0 {
		cl.SuggestedSemver = "major"
	} else if len(cl.Additions) > 0 {
		cl.SuggestedSemver = "minor"
	} else if len(cl.Info) > 0 {
		cl.SuggestedSemver = "patch"
	}

	return cl, nil
}

// ToMarkdown formats the changelog as standard GitHub-flavored Markdown.
func (c *Changelog) ToMarkdown() string {
	var sb strings.Builder
	sb.WriteString("# API Changelog\n\n")
	sb.WriteString(fmt.Sprintf("**Suggested Version Bump:** `%s`\n\n", c.SuggestedSemver))

	if len(c.Breaking) > 0 {
		sb.WriteString("### 🚨 Breaking Changes\n\n")
		for _, item := range c.Breaking {
			sb.WriteString(fmt.Sprintf("- %s\n", item.Summary))
		}
		sb.WriteString("\n")
	}

	if len(c.Additions) > 0 {
		sb.WriteString("### ✨ Additions\n\n")
		for _, item := range c.Additions {
			sb.WriteString(fmt.Sprintf("- %s\n", item.Summary))
		}
		sb.WriteString("\n")
	}

	if len(c.Info) > 0 {
		sb.WriteString("### 📝 Other Changes\n\n")
		for _, item := range c.Info {
			sb.WriteString(fmt.Sprintf("- %s\n", item.Summary))
		}
		sb.WriteString("\n")
	}

	if len(c.Breaking)+len(c.Additions)+len(c.Info) == 0 {
		sb.WriteString("No changes detected.\n")
	}

	return sb.String()
}

// ToJSON formats the changelog as pretty-printed JSON.
func (c *Changelog) ToJSON() (string, error) {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
