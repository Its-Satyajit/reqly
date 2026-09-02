package scim

import (
	"testing"
)

func TestUserValidation(t *testing.T) {
	if err := ValidateUser(User{UserName: "alice"}); err != nil {
		t.Fatalf("valid: %v", err)
	}
	if err := ValidateUser(User{}); err == nil {
		t.Fatalf("expected error for missing userName")
	}
	if err := ValidateUser(User{UserName: "bob", Email: "not-an-email"}); err == nil {
		t.Fatalf("expected error for bad email")
	}
}

func TestStore_UserLifecycle(t *testing.T) {
	s := NewStore()
	u, err := s.CreateUser(User{UserName: "alice", Email: "alice@example.com"})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID == "" || !u.Active {
		t.Fatalf("unexpected user %+v", u)
	}
	// Duplicate userName should fail
	if _, err := s.CreateUser(User{UserName: "alice"}); err == nil {
		t.Fatalf("expected duplicate error")
	}
	got, err := s.GetUser(u.ID)
	if err != nil || got.UserName != "alice" {
		t.Fatalf("GetUser: %v, %+v", err, got)
	}
	if err := s.DeactivateUser(u.ID); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	got2, _ := s.GetUser(u.ID)
	if got2.Active {
		t.Fatalf("should be inactive")
	}
}

func TestStore_GroupAndMembership(t *testing.T) {
	s := NewStore()
	u, _ := s.CreateUser(User{UserName: "bob"})
	g, err := s.CreateGroup(Group{DisplayName: "eng"})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if err := s.AddUserToGroup(u.ID, g.ID); err != nil {
		t.Fatalf("AddUserToGroup: %v", err)
	}
	// Check membership
	u2, _ := s.GetUser(u.ID)
	if len(u2.Groups) != 1 || u2.Groups[0] != g.ID {
		t.Fatalf("unexpected groups %v", u2.Groups)
	}
	// Duplicate add is no-op
	if err := s.AddUserToGroup(u.ID, g.ID); err != nil {
		t.Fatalf("duplicate add should be no-op: %v", err)
	}
}

func TestStore_ListUsers(t *testing.T) {
	s := NewStore()
	s.CreateUser(User{UserName: "a"})
	s.CreateUser(User{UserName: "b"})
	if len(s.ListUsers()) != 2 {
		t.Fatalf("want 2")
	}
}
