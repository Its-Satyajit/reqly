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

package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/variables"
)

func TestComputeDigestResponseMD5RFC2617(t *testing.T) {
	got, err := computeDigestResponse("GET", "/dir/index.html", "Mufasa", "Circle Of Life",
		"testrealm@host.com", "", "dcd98b7102dd2f0e8b11d0f600bfb0c093",
		"0a4f113b", "00000001", "auth", "")
	if err != nil {
		t.Fatal(err)
	}
	if want := "6629fae49393a05397450978507c4ef1"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestComputeDigestResponseSHA256(t *testing.T) {
	got, err := computeDigestResponse("GET", "/dir/index.html", "Mufasa", "Circle of Life",
		"http-auth@example.org", "SHA-256", "7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v",
		"f2/wE4q74E6zIJEtWaHKaf5wv/H5QzzpXusqGemxURZJ", "00000001", "auth", "")
	if err != nil {
		t.Fatal(err)
	}
	if want := "753927fa0e85d155564e2e272a28d1802ca10daf4496794697cf8db5856cb6c1"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestComputeDigestResponseWithoutQop(t *testing.T) {
	got, err := computeDigestResponse("GET", "/dir/index.html", "Mufasa", "Circle Of Life",
		"testrealm@host.com", "", "dcd98b7102dd2f0e8b11d0f600bfb0c093",
		"", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	// RFC 2617 legacy response (no qop): H(HA1:nonce:HA2).
	HA1 := md5Hex("Mufasa:testrealm@host.com:Circle Of Life")
	HA2 := md5Hex("GET:/dir/index.html")
	want := md5Hex(HA1 + ":dcd98b7102dd2f0e8b11d0f600bfb0c093:" + HA2)
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestDigestChallengeSetsHeader(t *testing.T) {
	old := newCNonce
	newCNonce = func() string { return "0a4f113b" }
	defer func() { newCNonce = old }()

	req := httptest.NewRequest(http.MethodGet, "https://example.com/dir/index.html", nil)
	s := digestScheme{}
	challenge := `Digest realm="testrealm@host.com", qop="auth", nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093", opaque="5ccc069c403ebaf9f0171e9517f40e41"`
	err := s.Challenge(req, challenge, map[string]string{
		"username": "Mufasa",
		"password": "Circle Of Life",
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	hdr := req.Header.Get("Authorization")
	if !strings.HasPrefix(hdr, "Digest ") {
		t.Fatalf("expected Digest header, got %q", hdr)
	}
	for _, want := range []string{
		`username="Mufasa"`,
		`realm="testrealm@host.com"`,
		`nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093"`,
		`uri="/dir/index.html"`,
		`qop=auth`,
		`nc=00000001`,
		`cnonce="0a4f113b"`,
		`opaque="5ccc069c403ebaf9f0171e9517f40e41"`,
		`response="6629fae49393a05397450978507c4ef1"`,
	} {
		if !strings.Contains(hdr, want) {
			t.Fatalf("header missing %q: %q", want, hdr)
		}
	}
}

func TestDigestApplyValidatesConfig(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	s := digestScheme{}
	if err := s.Apply(req, nil, variables.NewSet()); err == nil {
		t.Fatal("expected error when username/password missing")
	}
	if err := s.Apply(req, map[string]string{"username": "u"}, variables.NewSet()); err == nil {
		t.Fatal("expected error when password missing")
	}
	if !strings.Contains(errFor(s.Apply(req, map[string]string{"username": "u"}, variables.NewSet())).Error(), "password") {
		t.Fatal("expected error to mention password")
	}
}

func errFor(err error) error { return err }
