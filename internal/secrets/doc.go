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

// Package secrets handles sensitive credential storage. It defines the Store
// interface for persisting secret values (tokens) keyed by a stable
// identifier, with two implementations behind it: FileStore writes atomically
// at 0600 permissions to a JSON file, and KeychainStore keeps values in the
// OS credential store (Secret Service / Keychain / WinCred) with a 0600
// key index. Masking of values in output is handled by the environments
// masker.
package secrets
