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
)

// noneScheme applies no credentials. Combined with the inheritance rule that a
// non-empty auth type replaces inherited auth (internal/collections), a
// request with auth.type "none" is sent unauthenticated even under an
// auth-bearing collection or folder.
type noneScheme struct{}

// Apply is a no-op; "none" means no credentials.
func (noneScheme) Apply(req *http.Request, _ map[string]string, _ Interpolator) error {
	return nil
}

func init() {
	Register("none", noneScheme{})
}
