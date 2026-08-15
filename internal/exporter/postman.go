// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package exporter serializes Reqly workspaces and requests into shareable
// formats (Postman collections, cURL, ...).
package exporter

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
)

// PostmanCollection is a Postman Collection v2.1 document.
type PostmanCollection struct {
	Info PostmanInfo   `json:"info"`
	Item []PostmanItem `json:"item"`
}

// PostmanInfo describes the collection.
type PostmanInfo struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

// PostmanItem is a request or a folder of items.
type PostmanItem struct {
	Name    string          `json:"name,omitempty"`
	Item    []PostmanItem   `json:"item,omitempty"`
	Request *PostmanRequest `json:"request,omitempty"`
}

// PostmanRequest models a Postman v2.1 request.
type PostmanRequest struct {
	Method string          `json:"method"`
	Header []PostmanHeader `json:"header"`
	Body   *PostmanBody    `json:"body,omitempty"`
	URL    *PostmanURL     `json:"url,omitempty"`
}

// PostmanHeader is a name/value header.
type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanBody carries raw request bodies.
type PostmanBody struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw"`
}

// PostmanURL is a URL split into host, path, and query.
type PostmanURL struct {
	Raw   string             `json:"raw"`
	Query []PostmanQueryPair `json:"query,omitempty"`
}

// PostmanQueryPair is a query parameter.
type PostmanQueryPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ExportToPostman converts a workspace name plus its requests into a Postman
// collection. Requests is a flat list keyed by display name; collection-level
// base URL/headers are applied per request.
func ExportToPostman(name string, requests []request.Request) (*PostmanCollection, error) {
	collection := &PostmanCollection{
		Info: PostmanInfo{
			Name:   name,
			Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Item: []PostmanItem{},
	}

	for _, r := range requests {
		item := PostmanItem{
			Name:    displayName(r),
			Request: toPostmanRequest(r),
		}
		collection.Item = append(collection.Item, item)
	}

	return collection, nil
}

// toPostmanRequest converts a request into the Postman v2.1 shape.
func toPostmanRequest(r request.Request) *PostmanRequest {
	headers := make([]PostmanHeader, 0, len(r.Headers))
	for _, h := range r.Headers {
		headers = append(headers, PostmanHeader{Key: h.Key, Value: h.Value})
	}

	var body *PostmanBody
	if r.Body != "" {
		body = &PostmanBody{Mode: "raw", Raw: r.Body}
	}

	pr := &PostmanRequest{
		Method: string(r.Method),
		Header: headers,
		Body:   body,
		URL:    toPostmanURL(r.URL, r.Query),
	}
	return pr
}

// toPostmanURL splits a URL into raw + query parameters.
func toPostmanURL(rawURL string, query []request.Parameter) *PostmanURL {
	pu := &PostmanURL{Raw: rawURL}
	for _, p := range query {
		pu.Query = append(pu.Query, PostmanQueryPair{Key: p.Key, Value: p.Value})
	}
	return pu
}

// displayName returns a sensible item name.
func displayName(r request.Request) string {
	if r.Name != "" {
		return r.Name
	}
	if r.URL != "" {
		return string(r.Method) + " " + r.URL
	}
	return string(r.Method)
}

// ExportToPostmanJSON is ExportToPostman followed by JSON marshaling.
func ExportToPostmanJSON(name string, requests []request.Request) ([]byte, error) {
	collection, err := ExportToPostman(name, requests)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(collection, "", "  ")
}

// ParsePostman converts a Postman collection v2.1 document back into a slice
// of requests (useful for round-tripping and testing).
func ParsePostman(data []byte) (*PostmanCollection, error) {
	var collection PostmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("parse Postman collection: %w", err)
	}
	if collection.Info.Schema == "" && collection.Info.Name == "" {
		return nil, fmt.Errorf("not a Postman collection")
	}
	return &collection, nil
}

// ToRequests flattens a Postman collection (folders included) into requests.
// Collection-relative URLs are resolved against baseURL when provided.
func (c *PostmanCollection) ToRequests(baseURL string) []request.Request {
	var requests []request.Request
	c.collect(c.Item, baseURL, &requests)
	return requests
}

func (c *PostmanCollection) collect(items []PostmanItem, baseURL string, out *[]request.Request) {
	for _, item := range items {
		if item.Request != nil {
			r := item.Request.toRequest()
			if baseURL != "" && !strings.Contains(r.URL, "://") && r.URL != "" {
				if joined, err := joinURL(baseURL, r.URL); err == nil {
					r.URL = joined
				}
			}
			if item.Name != "" && r.Name == "" {
				r.Name = item.Name
			}
			*out = append(*out, r)
		}
		if len(item.Item) > 0 {
			c.collect(item.Item, baseURL, out)
		}
	}
}

// toRequest converts a Postman request into a Reqly request.
func (pr *PostmanRequest) toRequest() request.Request {
	r := request.Request{
		Method: request.Method(strings.ToUpper(pr.Method)),
		URL:    pr.URL.Raw,
	}
	for _, h := range pr.Header {
		r.Headers = append(r.Headers, request.Header{Key: h.Key, Value: h.Value})
	}
	if pr.Body != nil {
		r.Body = pr.Body.Raw
	}
	if pr.URL != nil {
		for _, q := range pr.URL.Query {
			r.Query = append(r.Query, request.Parameter{Key: q.Key, Value: q.Value})
		}
	}
	return r
}

// joinURL resolves a relative Postman path against a base URL.
func joinURL(base, path string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Path = strings.TrimSuffix(u.Path, "/") + "/" + strings.TrimPrefix(path, "/")
	return u.String(), nil
}
