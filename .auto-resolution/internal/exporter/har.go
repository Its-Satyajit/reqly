package exporter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Its-Satyajit/reqly/internal/history"
)

// HarOptions configures HAR export.
type HarOptions struct {
	Env   string
	Limit int
}

// harFile mirrors HAR 1.2 output.
type harLogOut struct {
	Version string        `json:"version"`
	Creator harCreatorOut `json:"creator"`
	Pages   []any         `json:"pages"`
	Entries []harEntryOut `json:"entries"`
}

type harCreatorOut struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Comment string `json:"comment,omitempty"`
}

type harEntryOut struct {
	StartedDateTime string         `json:"startedDateTime"`
	Time            float64        `json:"time"`
	Request         harRequestOut  `json:"request"`
	Response        harResponseOut `json:"response"`
	Cache           harCacheOut    `json:"cache"`
	Timings         harTimingsOut  `json:"timings"`
	Pageref         string         `json:"pageref,omitempty"`
}

type harRequestOut struct {
	Method      string          `json:"method"`
	URL         string          `json:"url"`
	HTTPVersion string          `json:"httpVersion"`
	Headers     []harNVPOut     `json:"headers"`
	Cookies     []any           `json:"cookies"`
	QueryString []harNVPOut     `json:"queryString"`
	PostData    *harPostDataOut `json:"postData,omitempty"`
	HeadersSize int             `json:"headersSize"`
	BodySize    int             `json:"bodySize"`
}

type harResponseOut struct {
	Status      int           `json:"status"`
	StatusText  string        `json:"statusText"`
	HTTPVersion string        `json:"httpVersion"`
	Headers     []harNVPOut   `json:"headers"`
	Content     harContentOut `json:"content"`
	RedirectURL string        `json:"redirectURL"`
	HeadersSize int           `json:"headersSize"`
	BodySize    int           `json:"bodySize"`
}

type harContentOut struct {
	Size     int    `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
	Encoding string `json:"encoding,omitempty"`
}

type harCacheOut struct{}

type harTimingsOut struct {
	Blocked float64 `json:"blocked"`
	DNS     float64 `json:"dns"`
	Connect float64 `json:"connect"`
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
	SSL     float64 `json:"ssl"`
}

type harNVPOut struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type harPostDataOut struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

// ExportHAR serializes history entries into HAR JSON (history -> HAR). Pure function.
func ExportHAR(entries []history.Entry, version string, mask func(string) string) ([]byte, error) {
	if version == "" {
		version = "dev"
	}
	applyMask := func(s string) string {
		if mask != nil {
			return mask(s)
		}
		return s
	}
	outEntries := make([]harEntryOut, 0, len(entries))
	for _, e := range entries {
		reqHeaders := nvpFromMap(e.ReqHeaders, applyMask)
		// split Cookie header into cookies array empty for now (keep headers)
		query := nvpFromQuery(e.URL)
		var postData *harPostDataOut
		if len(e.ReqBody) > 0 {
			mime := guessMime(e.ReqHeaders)
			text := string(e.ReqBody)
			// detect binary via mime
			if isBinaryMime(mime) {
				text = base64.StdEncoding.EncodeToString(e.ReqBody)
				postData = &harPostDataOut{MimeType: mime, Text: text}
				// encoding will be set via response content, but for request we keep text base64? HAR spec: postData.text + encoding
				// For simplicity, keep as base64 but mark?
				// We'll set encoding by leaving text base64 and adding comment? Instead we store base64 text and rely on caller to know.
				// For M28, store base64 and set via Content encoding field not available for request; keep base64 text as-is.
			} else {
				postData = &harPostDataOut{MimeType: mime, Text: applyMask(text)}
			}
		}
		respHeaders := nvpFromMap(e.RespHeaders, applyMask)
		respMime := guessMime(e.RespHeaders)
		respText := string(e.RespBody)
		encoding := ""
		if isBinaryMime(respMime) && len(e.RespBody) > 0 {
			respText = base64.StdEncoding.EncodeToString(e.RespBody)
			encoding = "base64"
		} else {
			respText = applyMask(respText)
		}
		// timings synthesized
		ms := float64(e.DurationMS)
		if ms < 0 {
			ms = 0
		}
		timings := harTimingsOut{
			Blocked: 0,
			DNS:     0,
			Connect: 0,
			Send:    ms * 0.1,
			Wait:    ms * 0.8,
			Receive: ms * 0.1,
			SSL:     0,
		}
		// startedDateTime from CreatedAt if available else now
		started := e.CreatedAt
		if started.IsZero() {
			started = time.Now()
		}
		out := harEntryOut{
			StartedDateTime: started.Format(time.RFC3339Nano),
			Time:            ms,
			Request: harRequestOut{
				Method:      e.Method,
				URL:         applyMask(e.URL),
				HTTPVersion: "HTTP/1.1",
				Headers:     reqHeaders,
				Cookies:     []any{},
				QueryString: query,
				PostData:    postData,
				HeadersSize: -1,
				BodySize:    len(e.ReqBody),
			},
			Response: harResponseOut{
				Status:      e.Status,
				StatusText:  httpStatusText(e.Status),
				HTTPVersion: "HTTP/1.1",
				Headers:     respHeaders,
				Content: harContentOut{
					Size:     len(e.RespBody),
					MimeType: respMime,
					Text:     respText,
					Encoding: encoding,
				},
				RedirectURL: "",
				HeadersSize: -1,
				BodySize:    len(e.RespBody),
			},
			Cache:   harCacheOut{},
			Timings: timings,
		}
		outEntries = append(outEntries, out)
	}
	file := map[string]any{
		"log": harLogOut{
			Version: "1.2",
			Creator: harCreatorOut{Name: "reqly", Version: version, Comment: "exported from history"},
			Pages:   []any{},
			Entries: outEntries,
		},
	}
	return json.MarshalIndent(file, "", "  ")
}

func nvpFromMap(m map[string][]string, mask func(string) string) []harNVPOut {
	var out []harNVPOut
	for k, vals := range m {
		for _, v := range vals {
			out = append(out, harNVPOut{Name: k, Value: mask(v)})
		}
	}
	// stable order for determinism
	// sort by name
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Name < out[i].Name {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func nvpFromQuery(rawURL string) []harNVPOut {
	// parse query from URL
	idx := strings.Index(rawURL, "?")
	if idx < 0 {
		return nil
	}
	qs := rawURL[idx+1:]
	// strip fragment
	if hi := strings.Index(qs, "#"); hi >= 0 {
		qs = qs[:hi]
	}
	if qs == "" {
		return nil
	}
	var out []harNVPOut
	for _, pair := range strings.Split(qs, "&") {
		if pair == "" {
			continue
		}
		k, v, _ := strings.Cut(pair, "=")
		out = append(out, harNVPOut{Name: k, Value: v})
	}
	return out
}

func guessMime(headers map[string][]string) string {
	for k, vals := range headers {
		if strings.EqualFold(k, "Content-Type") && len(vals) > 0 {
			return vals[0]
		}
	}
	return "text/plain"
}

func isBinaryMime(mime string) bool {
	mime = strings.ToLower(mime)
	if strings.HasPrefix(mime, "text/") {
		return false
	}
	if strings.Contains(mime, "json") || strings.Contains(mime, "xml") || strings.Contains(mime, "javascript") {
		return false
	}
	// treat image/pdf/octet and others as binary
	return strings.HasPrefix(mime, "image/") || strings.HasPrefix(mime, "application/octet") || strings.HasPrefix(mime, "application/pdf") || strings.Contains(mime, "octet")
}

func httpStatusText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 204:
		return "No Content"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	default:
		return fmt.Sprintf("%d", code)
	}
}
