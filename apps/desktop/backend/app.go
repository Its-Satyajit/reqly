package main

// AppService is a Wails v3 service exposing the Go core to the frontend.
// It should stay thin: business logic belongs in the Go core (internal/).
type AppService struct{}

// NewAppService creates a new AppService.
func NewAppService() *AppService {
	return &AppService{}
}

// Greet is a placeholder binding proving the Go ↔ JavaScript bridge works.
// Remove once real core bindings exist.
func (s *AppService) Greet(name string) string {
	return "Hello, " + name
}