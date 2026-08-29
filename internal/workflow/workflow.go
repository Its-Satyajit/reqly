// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dop251/goja"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// StepResult records the outcome of one workflow step.
type StepResult struct {
	Name         string             `json:"name"`
	RequestPath  string             `json:"requestPath"`
	Passed       bool               `json:"passed"`
	RequestError string             `json:"requestError,omitempty"`
	Response     *response.Response `json:"response,omitempty"`
	Logs         []string           `json:"logs"`
}

// WorkflowStep defines one step in a visual or programmatic DAG workflow.
type WorkflowStep struct {
	ID        string            `json:"id" yaml:"id"`
	Name      string            `json:"name" yaml:"name"`
	Request   request.Request   `json:"request" yaml:"request"`
	Condition string            `json:"condition,omitempty" yaml:"condition,omitempty"` // Goja JS boolean expression
	Extract   map[string]string `json:"extract,omitempty" yaml:"extract,omitempty"`     // varName -> jsonKey
}

// WorkflowOptions configures options for workflow execution.
type WorkflowOptions struct {
	EnvironmentVars map[string]string
}

// Workflow defines a structured multi-step API execution workflow.
type Workflow struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	Steps       []WorkflowStep    `json:"steps" yaml:"steps"`
}

// WorkflowReport summarizes the execution of a visual workflow.
type WorkflowReport struct {
	WorkflowName  string            `json:"workflowName"`
	Passed        bool              `json:"passed"`
	Duration      time.Duration     `json:"duration"`
	Steps         []StepResult      `json:"steps"`
	ExtractedVars map[string]string `json:"extractedVars"`
}

// WorkflowExecutor executes a Workflow step-by-step with variable extraction and conditional logic.
type WorkflowExecutor struct {
	client *request.Client
}

// NewWorkflowExecutor returns a WorkflowExecutor.
func NewWorkflowExecutor(client *request.Client) *WorkflowExecutor {
	if client == nil {
		client = request.NewClient()
	}
	return &WorkflowExecutor{client: client}
}

// Execute runs the workflow steps in sequence with variable propagation.
func (we *WorkflowExecutor) Execute(ctx context.Context, wf *Workflow, opts WorkflowOptions) (*WorkflowReport, error) {
	if wf == nil {
		return nil, fmt.Errorf("workflow is nil")
	}
	start := time.Now()
	report := &WorkflowReport{
		WorkflowName:  wf.Name,
		Passed:        true,
		ExtractedVars: make(map[string]string),
	}

	varStore := variables.NewSet()
	for k, v := range wf.Variables {
		varStore.Set(variables.ScopeGlobal, k, v)
	}
	if opts.EnvironmentVars != nil {
		for k, v := range opts.EnvironmentVars {
			varStore.Set(variables.ScopeEnvironment, k, v)
		}
	}

	for _, step := range wf.Steps {
		// 1. Evaluate condition if specified
		if step.Condition != "" {
			vm := goja.New()
			reqlyObj := vm.NewObject()
			reqlyObj.Set("getVariable", func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) == 0 {
					return vm.ToValue("")
				}
				vName := call.Arguments[0].String()
				val, _ := varStore.Resolve(vName)
				return vm.ToValue(val)
			})
			_ = vm.Set("reqly", reqlyObj)
			val, err := vm.RunString(step.Condition)
			if err == nil && !val.ToBoolean() {
				// Condition evaluated to false, skip step
				continue
			}
		}

		// 2. Interpolate request with variables — clone slices to avoid mutating the original step.
		reqCopy := step.Request
		if len(reqCopy.Headers) > 0 {
			cloned := make([]request.Header, len(reqCopy.Headers))
			copy(cloned, reqCopy.Headers)
			reqCopy.Headers = cloned
			for i, h := range reqCopy.Headers {
				if v, err := varStore.Interpolate(h.Value); err == nil {
					reqCopy.Headers[i].Value = v
				}
			}
		}
		if len(reqCopy.Query) > 0 {
			cloned := make([]request.Parameter, len(reqCopy.Query))
			copy(cloned, reqCopy.Query)
			reqCopy.Query = cloned
			for i, q := range reqCopy.Query {
				if v, err := varStore.Interpolate(q.Value); err == nil {
					reqCopy.Query[i].Value = v
				}
			}
		}
		if reqCopy.URL != "" {
			if v, err := varStore.Interpolate(reqCopy.URL); err == nil {
				reqCopy.URL = v
			}
		}
		if reqCopy.Body != "" {
			if v, err := varStore.Interpolate(reqCopy.Body); err == nil {
				reqCopy.Body = v
			}
		}

		// 3. Execute HTTP request
		stepStart := time.Now()
		resp, err := we.client.Execute(ctx, &reqCopy, varStore)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		stepRes := StepResult{
			Name:         step.Name,
			RequestPath:  step.ID,
			Response:     resp,
			RequestError: errStr,
			Passed:       err == nil && (resp != nil && resp.StatusCode < 400),
		}

		if err != nil || resp == nil || resp.StatusCode >= 400 {
			report.Passed = false
		}

		// 4. Extract variables from JSON response body
		if resp != nil && len(resp.Body) > 0 && len(step.Extract) > 0 {
			var parsed map[string]any
			if jsonErr := json.Unmarshal(resp.Body, &parsed); jsonErr == nil {
				for varName, jsonKey := range step.Extract {
					if val, ok := parsed[jsonKey]; ok {
						strVal := fmt.Sprintf("%v", val)
						varStore.Set(variables.ScopeGlobal, varName, strVal)
						report.ExtractedVars[varName] = strVal
					}
				}
			}
		}

		if resp != nil {
			stepRes.Logs = append(stepRes.Logs, fmt.Sprintf("%s %s -> %d (%dms)", reqCopy.Method, reqCopy.URL, resp.StatusCode, time.Since(stepStart).Milliseconds()))
		} else if err != nil {
			stepRes.Logs = append(stepRes.Logs, fmt.Sprintf("%s %s -> error: %s", reqCopy.Method, reqCopy.URL, errStr))
		}
		report.Steps = append(report.Steps, stepRes)
	}

	report.Duration = time.Since(start)
	return report, nil
}
