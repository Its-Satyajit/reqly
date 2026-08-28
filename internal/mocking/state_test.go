package mocking

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestStateMachine_Transitions(t *testing.T) {
	scenario := &Scenario{
		InitialState: "logged_out",
		States: map[string]StateConfig{
			"logged_out": {
				Transitions: []TransitionRule{
					{
						Method:      "POST",
						Path:        "/api/login",
						TargetState: "logged_in",
						Response: MockResponse{
							Status: 200,
							Body:   `{"token": "xyz123"}`,
						},
					},
					{
						Method: "GET",
						Path:   "/api/profile",
						Response: MockResponse{
							Status: 401,
							Body:   `{"error": "unauthorized"}`,
						},
					},
				},
			},
			"logged_in": {
				Transitions: []TransitionRule{
					{
						Method: "GET",
						Path:   "/api/profile",
						Response: MockResponse{
							Status: 200,
							Body:   `{"user": "alice"}`,
						},
					},
					{
						Method:      "POST",
						Path:        "/api/logout",
						TargetState: "logged_out",
						Response: MockResponse{
							Status: 204,
						},
					},
				},
			},
		},
	}

	sm := NewStateMachine(scenario)
	if sm.CurrentState() != "logged_out" {
		t.Fatalf("want logged_out, got %s", sm.CurrentState())
	}

	// 1. In logged_out: GET /api/profile -> 401
	resp, ok := sm.Handle("GET", "/api/profile")
	if !ok || resp.Status != 401 {
		t.Fatalf("expected 401 in logged_out, got %v (%d)", ok, resp.Status)
	}

	// 2. In logged_out: POST /api/login -> 200, transitions to logged_in
	resp, ok = sm.Handle("POST", "/api/login")
	if !ok || resp.Status != 200 || sm.CurrentState() != "logged_in" {
		t.Fatalf("expected login transition to logged_in, got %s (%d)", sm.CurrentState(), resp.Status)
	}

	// 3. In logged_in: GET /api/profile -> 200
	resp, ok = sm.Handle("GET", "/api/profile")
	if !ok || resp.Status != 200 || resp.Body != `{"user": "alice"}` {
		t.Fatalf("expected 200 profile in logged_in, got %v (%d)", ok, resp.Status)
	}

	// 4. In logged_in: POST /api/logout -> 204, transitions to logged_out
	resp, ok = sm.Handle("POST", "/api/logout")
	if !ok || resp.Status != 204 || sm.CurrentState() != "logged_out" {
		t.Fatalf("expected logout transition to logged_out, got %s (%d)", sm.CurrentState(), resp.Status)
	}
}

func TestServer_StateMachineIntegration(t *testing.T) {
	scenario := &Scenario{
		InitialState: "cart_empty",
		States: map[string]StateConfig{
			"cart_empty": {
				Transitions: []TransitionRule{
					{
						Method:      "POST",
						Path:        "/cart/add",
						TargetState: "cart_with_items",
						Response: MockResponse{
							Status: 201,
							Body:   `{"count": 1}`,
						},
					},
					{
						Method: "GET",
						Path:   "/cart",
						Response: MockResponse{
							Status: 200,
							Body:   `{"items": []}`,
						},
					},
				},
			},
			"cart_with_items": {
				Transitions: []TransitionRule{
					{
						Method: "GET",
						Path:   "/cart",
						Response: MockResponse{
							Status: 200,
							Body:   `{"items": ["item-1"]}`,
						},
					},
				},
			},
		},
	}

	sm := NewStateMachine(scenario)
	srv, err := NewServer(nil, WithStateMachine(sm))
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}

	// Check initial state
	req := httptest.NewRequest("GET", "/cart", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Body.String() != `{"items": []}` {
		t.Errorf("want empty items, got %s", rec.Body.String())
	}

	// Check control endpoint GET /__reqly/state
	reqCtrl := httptest.NewRequest("GET", "/__reqly/state", nil)
	recCtrl := httptest.NewRecorder()
	srv.ServeHTTP(recCtrl, reqCtrl)
	if recCtrl.Code != 200 || !bytes.Contains(recCtrl.Body.Bytes(), []byte("cart_empty")) {
		t.Errorf("unexpected state control get: %s", recCtrl.Body.String())
	}

	// Add to cart -> trigger state change
	reqAdd := httptest.NewRequest("POST", "/cart/add", nil)
	recAdd := httptest.NewRecorder()
	srv.ServeHTTP(recAdd, reqAdd)
	if recAdd.Code != 201 {
		t.Errorf("want 201, got %d", recAdd.Code)
	}

	// Now GET /cart should return items
	reqCart := httptest.NewRequest("GET", "/cart", nil)
	recCart := httptest.NewRecorder()
	srv.ServeHTTP(recCart, reqCart)
	if recCart.Body.String() != `{"items": ["item-1"]}` {
		t.Errorf("want item-1, got %s", recCart.Body.String())
	}
}
