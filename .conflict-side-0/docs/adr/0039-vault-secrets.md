# ADR 0039: Vault Secrets — HashiCorp Vault KV v2 Store (M72)

## Status
Accepted — grill Q1 (core + backend, AWS/Azure deferred)

## Context
Enterprise §58.5 secret management needs Vault without CGO. Existing FileStore and KeychainStore are local.

## Decision
`internal/secrets.VaultStore` (M72): `VaultConfig{Addr,Token,Mount,Prefix}` + `NewVaultStore` (addr/token required, mount default `secret`, prefix `reqly/`) + `Get`/`Set`/`Delete`/`Keys` via `X-Vault-Token` (`secret/data/prefix/<key>`, `secret/metadata/prefix` LIST), `REQLY_TOKEN_STORE=vault` with `VAULT_ADDR`/`TOKEN`/`MOUNT`/`PREFIX` env, fallback to file store.

## Consequences
Q1: Vault KV v2 only — KV v1 and AWS/Azure deferred.
Q2: No Vault agent — token is env var, not auto-renewed.
