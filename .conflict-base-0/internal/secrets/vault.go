// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
)

// VaultStore is a Store backed by HashiCorp Vault KV v2.
// It stores each key as Vault's `value` field under mount/data/prefix/<key>.
type VaultStore struct {
	addr   string
	token  string
	mount  string
	prefix string
	client *http.Client
}

// VaultConfig configures a VaultStore.
type VaultConfig struct {
	Addr   string // e.g. http://127.0.0.1:8200
	Token  string // Vault token
	Mount  string // KV mount, default "secret"
	Prefix string // key prefix, default "reqly/"
}

// NewVaultStore returns a VaultStore. Addr and Token are required.
func NewVaultStore(cfg VaultConfig) (*VaultStore, error) {
	if strings.TrimSpace(cfg.Addr) == "" {
		return nil, fmt.Errorf("vault addr is required")
	}
	if strings.TrimSpace(cfg.Token) == "" {
		return nil, fmt.Errorf("vault token is required")
	}
	mount := cfg.Mount
	if mount == "" {
		mount = "secret"
	}
	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "reqly/"
	}
	// Ensure prefix ends with /
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	// Validate addr is URL
	if _, err := url.Parse(cfg.Addr); err != nil {
		return nil, fmt.Errorf("vault addr: %w", err)
	}
	client := &http.Client{}
	return &VaultStore{
		addr:   strings.TrimRight(cfg.Addr, "/"),
		token:  cfg.Token,
		mount:  mount,
		prefix: prefix,
		client: client,
	}, nil
}

// vaultData is the KV v2 data envelope.
type vaultData struct {
	Data vaultInner `json:"data"`
}

type vaultInner struct {
	Data map[string]string `json:"data"`
}

// Get returns the value for key, or fs.ErrNotExist when absent.
func (v *VaultStore) Get(key string) (string, error) {
	u := fmt.Sprintf("%s/v1/%s/data/%s%s", v.addr, v.mount, v.prefix, url.PathEscape(key))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Vault-Token", v.token)
	resp, err := v.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("secrets: %w: %q", fs.ErrNotExist, key)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("vault get %q: %s: %s", key, resp.Status, string(body))
	}
	var out vaultData
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	val, ok := out.Data.Data["value"]
	if !ok {
		return "", fmt.Errorf("secrets: %w: %q", fs.ErrNotExist, key)
	}
	return val, nil
}

// Set stores value under key.
func (v *VaultStore) Set(key, value string) error {
	u := fmt.Sprintf("%s/v1/%s/data/%s%s", v.addr, v.mount, v.prefix, url.PathEscape(key))
	payload := map[string]any{
		"data": map[string]string{"value": value},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault set %q: %s: %s", key, resp.Status, string(b))
	}
	return nil
}

// Delete removes a key. Deleting an absent key is not an error.
func (v *VaultStore) Delete(key string) error {
	u := fmt.Sprintf("%s/v1/%s/data/%s%s", v.addr, v.mount, v.prefix, url.PathEscape(key))
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Vault-Token", v.token)
	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault delete %q: %s: %s", key, resp.Status, string(b))
	}
	return nil
}

// Keys lists keys under the prefix via Vault LIST.
// If LIST is not supported or returns 404, it returns nil.
func (v *VaultStore) Keys() ([]string, error) {
	u := fmt.Sprintf("%s/v1/%s/metadata/%s", v.addr, v.mount, v.prefix)
	req, err := http.NewRequest("LIST", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", v.token)
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vault list: %s: %s", resp.Status, string(b))
	}
	var out struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Data.Keys, nil
}
