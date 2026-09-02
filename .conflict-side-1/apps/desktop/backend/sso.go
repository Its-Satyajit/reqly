// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/Its-Satyajit/reqly/internal/scim"
	"github.com/Its-Satyajit/reqly/internal/sso"
)

// SSOValidate validates an OIDC token for the desktop.
func (s *AppService) SSOValidate(issuer, clientID, token, secret string) error {
	cfg := sso.Config{Issuer: issuer, ClientID: clientID}
	return sso.ValidateToken(cfg, token, []byte(secret))
}

// SCIMCreateUser creates a SCIM user (in-memory for M73).
func (s *AppService) SCIMCreateUser(username, email string) (scim.User, error) {
	store := scim.NewStore()
	return store.CreateUser(scim.User{UserName: username, Email: email})
}

// SCIMListUsers lists SCIM users.
func (s *AppService) SCIMListUsers() []scim.User {
	store := scim.NewStore()
	return store.ListUsers()
}
