# Phase 3: Power-User Features (P2)

## Phase 3 — Power-User Features (P2)

Advanced functionality for experienced developers and teams.
**Spec:** [`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](docs/Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) §57

### §57.1 API Monitoring Dashboard

- [ ] Scheduled requests/collections, health checks, latency/availability, alerts

### §57.2 Performance Testing

- [ ] RPS, latency, P95/P99, error rate, status distribution

### §57.3 MQTT / Socket.IO

- [x] MQTT publish/subscribe, topics, QoS (0,1,2), retain, TLS, CLI `reqly mqtt pub/sub`, and Goja binding `reqly.mqtt` (`internal/mqtt`, [M57](docs/spec/m57-mqtt-protocol-engine.md), [ADR 0041](docs/adr/0041-mqtt-protocol-engine.md)) shipped
- [x] Socket.IO connections, events, rooms, namespaces, CLI `reqly socketio connect/emit`, and Goja binding `reqly.socketio` (`internal/socketio`, [M58](docs/spec/m58-socketio-protocol-engine.md), [ADR 0042](docs/adr/0042-socketio-protocol-engine.md)) shipped

### §57.4 Dependency Graph

- [ ] API dependency graph visualization

### §57.5 Request Replay

- [ ] Exact / modified vars / other env / captured traffic replay

### §57.6 In-app Developer Tools / Debugger

- [ ] Request/auth/variables/script/runtime/network inspection

### §57.7 Git GUI

- [ ] Init/commit/branch/diff/history/pull/push/merge/conflicts

### §57.8 Network Interception / Timeline Debugging

- [x] Request timeline debugging (DNS/connect/TLS/request/server/response/transfer) — `response.Timings` (`internal/response`), `internal/request.Client` `httptrace` synthesis, CLI `reqly run --timeline`, and Goja binding `reqly.response.timings` ([M59](docs/spec/m59-request-timeline-debugging.md), [ADR 0043](docs/adr/0043-request-timeline-debugging.md)) shipped
- [ ] Capture/inspect/import/modify/replay network traffic

### Other P2 Items

- [ ] API changelog (from specs + Git changes)
- [ ] Browser integrations (DevTools import, cURL copy, Chrome/Firefox/Safari)
- [ ] Advanced mock state (multi-scenario state machines)
- [ ] Visual workflow builder
- [ ] Self-hosted automation

---

