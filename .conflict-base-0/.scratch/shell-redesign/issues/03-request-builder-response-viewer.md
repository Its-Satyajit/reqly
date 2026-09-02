# 03: Request Builder + Response Viewer

**What to build:** The primary workspace for composing and sending requests. Includes request tabs, URL bar, builder panels, and the response viewer with toggleable split.

**Blocked by:** #01 (Shell Foundation)

**Status:** ready-for-agent

## Acceptance Criteria

- [ ] Request tabs render with open/close/pin/duplicate/drag-reorder
- [ ] Tabs persist across app restarts (stored in workspace state)
- [ ] Unsaved indicator shows on modified tabs
- [ ] Request URL bar shows: method selector (GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS/CONNECT/TRACE), URL editor, Send button, Save button
- [ ] Send button triggers request execution via adapter
- [ ] 4 visible tabs render: Params, Headers, Body, Auth
- [ ] Overflow menu (⋮ More) contains: Pre-request, Tests, Docs, Settings
- [ ] Params tab shows query parameters table with enable/disable, key, value, description
- [ ] Headers tab shows headers table with enable/disable, key, value, source indicator
- [ ] Body tab supports: None, JSON, Raw, Text, XML, HTML, Form URL Encoded, Multipart Form, Binary, GraphQL
- [ ] Binary body shows file picker (drag-and-drop or browse)
- [ ] Auth tab supports: Inherit, No Auth, Basic, Bearer, API Key, OAuth 2.0, Digest, AWS Signature, Custom
- [ ] OAuth 2.0 supports all three flows: Client Credentials, Authorization Code + PKCE, Device Flow
- [ ] Response Viewer shows: status code, timing, size, HTTP version
- [ ] Response tabs: Body, Headers, Cookies, Test Results, Timeline
- [ ] Response split is toggleable between vertical and horizontal
- [ ] Draggable divider resizes the Request/Response split
- [ ] All components use theme tokens
