package scripting

import (
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/response"
)

func TestScripting_AI(t *testing.T) {
	resp := &response.Response{
		StatusCode: 200,
		StatusText: "OK",
		Duration:   45 * time.Millisecond,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: []byte(`{"message": "success"}`),
	}

	var generatedTests string
	var explanation string
	var diagnosis string

	sb := NewSandbox(SandboxOptions{
		Response: resp,
		SetVariable: func(key, val string) {
			switch key {
			case "tests":
				generatedTests = val
			case "explanation":
				explanation = val
			case "diagnosis":
				diagnosis = val
			}
		},
	})

	script := `
		reqly.setVariable("tests", reqly.ai.generateTests());
		reqly.setVariable("explanation", reqly.ai.explain());
		reqly.setVariable("diagnosis", reqly.ai.diagnose());
	`

	if err := sb.Run(script); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !strings.Contains(generatedTests, "Status code is 200") {
		t.Errorf("missing generated tests: %s", generatedTests)
	}
	if !strings.Contains(explanation, "response 200 OK") {
		t.Errorf("missing explanation: %s", explanation)
	}
	if !strings.Contains(diagnosis, "Success") {
		t.Errorf("missing diagnosis: %s", diagnosis)
	}
}
