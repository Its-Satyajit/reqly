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

// Package websocket provides WebSocket connection management, message
// composing, and incoming/outgoing message inspection for realtime APIs.
package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Status is the lifecycle state of a WebSocket connection.
type Status string

const (
	// StatusDisconnected means no connection is open.
	StatusDisconnected Status = "disconnected"
	// StatusConnecting means Dial is in progress.
	StatusConnecting Status = "connecting"
	// StatusConnected means the connection is open.
	StatusConnected Status = "connected"
	// StatusClosing means Close was requested.
	StatusClosing Status = "closing"
)

// MessageType classifies a WebSocket frame.
type MessageType string

const (
	// MessageText is a UTF-8 text frame.
	MessageText MessageType = "text"
	// MessageBinary is a binary frame.
	MessageBinary MessageType = "binary"
)

// Message is a single WebSocket frame with metadata.
type Message struct {
	// Type is the frame type.
	Type MessageType
	// Data is the payload.
	Data []byte
	// At is when the frame was received or sent.
	At time.Time
	// Sent marks outgoing (true) vs incoming (false) messages.
	Sent bool
}

// Client manages a single WebSocket connection. It is safe for concurrent use.
type Client struct {
	url          string
	headers      http.Header
	subprotocols []string

	mu   sync.RWMutex
	conn *websocket.Conn
	// statusFn receives status transitions (may be nil).
	statusFn func(Status)
}

// Option configures a Client.
type Option func(*Client)

// WithHeaders sets extra headers for the handshake.
func WithHeaders(headers http.Header) Option {
	return func(c *Client) { c.headers = headers }
}

// WithSubprotocols requests the given subprotocols.
func WithSubprotocols(subprotocols []string) Option {
	return func(c *Client) { c.subprotocols = subprotocols }
}

// WithStatusHandler receives every status transition.
func WithStatusHandler(fn func(Status)) Option {
	return func(c *Client) { c.statusFn = fn }
}

// NewClient returns a Client for the given URL.
func NewClient(url string, opts ...Option) *Client {
	c := &Client{
		url:     url,
		headers: http.Header{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// URL returns the connection URL.
func (c *Client) URL() string { return c.url }

// Status returns the current connection state.
func (c *Client) Status() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conn == nil {
		return StatusDisconnected
	}
	return StatusConnected
}

func (c *Client) setStatus(s Status) {
	if c.statusFn != nil {
		c.statusFn(s)
	}
}

// Dial opens the connection. It blocks until the handshake completes.
func (c *Client) Dial(ctx context.Context) error {
	c.mu.Lock()
	if c.conn != nil {
		c.mu.Unlock()
		return fmt.Errorf("websocket: already connected")
	}
	c.mu.Unlock()

	c.setStatus(StatusConnecting)

	opts := &websocket.DialOptions{
		HTTPHeader:   c.headers,
		Subprotocols: c.subprotocols,
	}
	conn, _, err := websocket.Dial(ctx, c.url, opts)
	if err != nil {
		c.setStatus(StatusDisconnected)
		return fmt.Errorf("websocket dial %s: %w", c.url, err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	c.setStatus(StatusConnected)
	return nil
}

// Close closes the connection with a normal close frame.
func (c *Client) Close(ctx context.Context) error {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()
	if conn == nil {
		return nil
	}
	c.setStatus(StatusClosing)
	return conn.Close(websocket.StatusNormalClosure, "")
}

// Send writes a text or binary message.
func (c *Client) Send(ctx context.Context, typ MessageType, data []byte) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return fmt.Errorf("websocket: not connected")
	}
	msgType := websocket.MessageText
	if typ == MessageBinary {
		msgType = websocket.MessageBinary
	}
	return conn.Write(ctx, msgType, data)
}

// SendText writes a UTF-8 text message.
func (c *Client) SendText(ctx context.Context, data string) error {
	return c.Send(ctx, MessageText, []byte(data))
}

// SendBinary writes a binary message.
func (c *Client) SendBinary(ctx context.Context, data []byte) error {
	return c.Send(ctx, MessageBinary, data)
}

// Receive blocks for the next frame and returns it. It returns an error when
// the connection closes or the context is cancelled.
func (c *Client) Receive(ctx context.Context) (Message, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return Message{}, fmt.Errorf("websocket: not connected")
	}

	typ, data, err := conn.Read(ctx)
	if err != nil {
		return Message{}, err
	}

	mt := MessageText
	if typ == websocket.MessageBinary {
		mt = MessageBinary
	}

	return Message{Type: mt, Data: data, At: time.Now()}, nil
}

// Ping sends a ping frame.
func (c *Client) Ping(ctx context.Context) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return fmt.Errorf("websocket: not connected")
	}
	return conn.Ping(ctx)
}

// Subprotocol returns the negotiated subprotocol, or "".
func (c *Client) Subprotocol() string {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return ""
	}
	return conn.Subprotocol()
}
