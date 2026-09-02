# 07: Mocks Page

**What to build:** The Mocks page for managing mock servers. Shows mock servers with status, start/stop/restart controls, and route configuration.

**Blocked by:** #01 (Shell Foundation)

**Status:** ready-for-agent

## Acceptance Criteria

- [ ] Mocks Page renders when activeView is 'mocks'
- [ ] Mock servers list shows: Name, Status (Running/Stopped), URL
- [ ] Green dot for running, red dot for stopped/crashed
- [ ] Start button starts the mock server (spawns separate process via backend)
- [ ] Stop button stops the mock server
- [ ] Restart button appears when mock is crashed
- [ ] Route configuration table shows: Method, Path, Status, Body file
- [ ] Add route button adds a new row
- [ ] Delete route removes the row
- [ ] Route editor shows: Method, Path, Status, Headers, Body, Latency, Scenario
- [ ] Mock server runs as separate process (not in GUI)
- [ ] GUI polls backend for mock process status
- [ ] Error message shows when mock crashes
- [ ] All components use theme tokens
