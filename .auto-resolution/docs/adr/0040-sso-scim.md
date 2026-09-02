# ADR 0040: SSO & SCIM — OIDC Validation and Provisioning (M73)

## Status
Accepted — grill Q1 (core + CLI + desktop, collaboration server deferred)

## Context
Enterprise §58.5 needs SSO/SCIM without cloud. JWT and OAuth2 exist, but not OIDC group checks or SCIM provisioning.

## Decision
`internal/sso` (M73): `Config{Issuer,ClientID,JWKSURL,AllowedGroups}` + `Validate` + `ValidateToken` (HMAC via `jwt.Verify`, issuer check, `IsGroupAllowed`) + `internal/scim` (M73): `User{ID,UserName,Email,Groups,Active}`/`Group{ID,DisplayName,Members}` + `ValidateUser/Group` + `Store{CreateUser/GetUser/ListUsers/DeactivateUser/CreateGroup/AddUserToGroup}` (in-memory, mutex) + CLI `reqly sso validate` + `reqly scim user create/list` + desktop `SSOValidate/SCIMCreateUser/ListUsers`.

## Consequences
Q1: HMAC only — RS256 via JWKS deferred.
Q2: SCIM in-memory — no SQLite persistence, no SCIM HTTP API.
