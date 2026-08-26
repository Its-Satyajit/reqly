// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log/slog"
)

// goLogEventName is the Wails custom event the frontend buffers so backend
// warnings and errors land in pasted crash reports.
const goLogEventName = "reqly.golog"

// newLogMirrorHandler wraps an inner terminal handler and mirrors every
// record at or above minLevel to the frontend as a reqly.golog event. The
// terminal keeps receiving everything the inner handler emits; the mirror is
// additive, so packaged builds lose nothing when no webview is listening.
func newLogMirrorHandler(inner slog.Handler, minLevel slog.Level) *logMirrorHandler {
	return &logMirrorHandler{inner: inner, minLevel: minLevel}
}

type logMirrorHandler struct {
	inner    slog.Handler
	minLevel slog.Level
	attrs    []slog.Attr
	group    string
}

func (h *logMirrorHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *logMirrorHandler) Handle(_ context.Context, record slog.Record) error {
	if h.inner != nil {
		if err := h.inner.Handle(context.Background(), record); err != nil {
			return err
		}
	}
	payload := map[string]any{
		"level":   record.Level.String(),
		"message": record.Message,
	}
	appendAttrs(payload, h.attrs)
	appendRecordAttrs(payload, record)
	emitRunEvent(goLogEventName, payload)
	return nil
}

func appendRecordAttrs(payload map[string]any, record slog.Record) {
	record.Attrs(func(attr slog.Attr) bool {
		payload[attr.Key] = attr.Value.Resolve().Any()
		return true
	})
}

func appendAttrs(payload map[string]any, attrs []slog.Attr) {
	for _, attr := range attrs {
		payload[attr.Key] = attr.Value.Resolve().Any()
	}
}

func (h *logMirrorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr(nil), h.attrs...), attrs...)
	return &clone
}

func (h *logMirrorHandler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.group = name
	return &clone
}
