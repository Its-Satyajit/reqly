// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package mocking

import (
	"strings"
	"sync"
)

// MockResponse represents a fixed mock payload and status.
type MockResponse struct {
	Status  int               `json:"status" yaml:"status"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body    string            `json:"body,omitempty" yaml:"body,omitempty"`
}

// TransitionRule specifies a rule to match in a given state, the response to return, and optional next state.
type TransitionRule struct {
	Method      string       `json:"method" yaml:"method"`
	Path        string       `json:"path" yaml:"path"`
	TargetState string       `json:"target_state,omitempty" yaml:"target_state,omitempty"`
	Response    MockResponse `json:"response" yaml:"response"`
}

// StateConfig defines available transitions and rules for a named state.
type StateConfig struct {
	Transitions []TransitionRule `json:"transitions" yaml:"transitions"`
}

// Scenario represents a multi-scenario state machine configuration.
type Scenario struct {
	Name         string                 `json:"name,omitempty" yaml:"name,omitempty"`
	InitialState string                 `json:"initial_state" yaml:"initial_state"`
	States       map[string]StateConfig `json:"states" yaml:"states"`
}

// StateMachine manages multi-scenario state transitions concurrently.
type StateMachine struct {
	mu           sync.RWMutex
	scenario     *Scenario
	currentState string
}

// NewStateMachine creates a new StateMachine from a Scenario definition.
func NewStateMachine(scenario *Scenario) *StateMachine {
	initial := ""
	if scenario != nil {
		initial = scenario.InitialState
	}
	return &StateMachine{
		scenario:     scenario,
		currentState: initial,
	}
}

// CurrentState returns the active state name.
func (sm *StateMachine) CurrentState() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}

// SetState resets or updates the active state.
func (sm *StateMachine) SetState(state string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.currentState = state
}

// Handle matches incoming request method and path against the current state's transitions.
// If matched, performs the transition (if TargetState is set) and returns the configured response.
func (sm *StateMachine) Handle(method, path string) (*MockResponse, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.scenario == nil || sm.scenario.States == nil {
		return nil, false
	}

	stateCfg, ok := sm.scenario.States[sm.currentState]
	if !ok {
		return nil, false
	}

	for _, rule := range stateCfg.Transitions {
		methodMatches := rule.Method == "" || strings.EqualFold(rule.Method, method)
		pathMatches := rule.Path == path || strings.TrimSuffix(rule.Path, "/") == strings.TrimSuffix(path, "/")

		if methodMatches && pathMatches {
			if rule.TargetState != "" {
				sm.currentState = rule.TargetState
			}
			resp := rule.Response
			if resp.Status == 0 {
				resp.Status = 200
			}
			return &resp, true
		}
	}

	return nil, false
}
