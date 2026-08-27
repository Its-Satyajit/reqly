package ai

import (
	"fmt"

	"github.com/Its-Satyajit/reqly/internal/response"
)

// ExplainResponse returns a fixed-template explanation for a response.
func ExplainResponse(resp *response.Response) string {
	if resp == nil {
		return "no response"
	}
	timings := ""
	if resp.Timings != nil {
		timings = fmt.Sprintf(" dns=%dms connect=%dms tls=%dms", resp.Timings.DNS, resp.Timings.Connect, resp.Timings.TLS)
	}
	return fmt.Sprintf("response %d %s in %dms (%dB) proto %s%s", resp.StatusCode, resp.StatusText, resp.Duration.Milliseconds(), resp.Size, resp.Proto, timings)
}
