# Spec: WSDL Import (Milestone 41)

> **Status:** Shipped 2026-08-23 — grill settled 2026-08-23 (Q1–Q8 confirmed)
> **Scope:** Phase 1 §1.8 `ROADMAP.md` — SOAP/WSDL: import, operation discovery, envelope skeleton (the "XML builder" surface ships as generated envelopes, not a runtime builder)
> **Stack:** `internal/importer/wsdl.go` (hand-rolled `encoding/xml`, no new deps) + subcommand in `apps/cli/cmd/import.go`
> **Fixtures:** vendored at `internal/importer/testdata/import-suite/wsdl/fixtures/`

## Problem Statement

Reqly imports REST, GraphQL, and the big three collection formats but cannot consume a WSDL. Developers working against legacy SOAP services must hand-craft envelopes. The roadmap line is "WSDL import, operation discovery, XML builder" — this milestone delivers import with generated envelope skeletons; there is no separate runtime XML builder.

## Solution

* **CLI** — `reqly import wsdl <file> [--output dir]`, file-only like sibling importers:
  * writes a workspace: `reqly.yaml` + one collection per `<wsdl:service>` + one request file per operation
* **Operation discovery**: portType operations joined with their binding (SOAPAction / soap:operation) and service port address (`soap:address location`)
* **Generated request per operation**:
  * POST to the port's address
  * headers: `Content-Type` and `SOAPAction: "<action>"` per binding version
  * body: full envelope skeleton — SOAP 1.1 (`http://schemas.xmlsoap.org/soap/envelope/`, prefix `soapenv`) or SOAP 1.2 (`.../soap12/envelope/`, Content-Type `application/soap+xml; charset=utf-8`) matched to the binding namespace
  * `<soapenv:Header/>` empty; `<soapenv:Body>` contains the input root element (targetNamespace as default xmlns) with children from its inline complexType sequence
  * placeholder values: strings → empty element, numerics → `0`, booleans → empty element; optional (`minOccurs="0"`) children omitted
  * naming: `<operationName>.yaml`; collisions suffixed `-2`; request name carries short `wsdl:documentation` when present
* **Degradations (warn, never fail)**: external `xsd:import`/`xsd:include` not followed → root-only skeleton; rpc/encoded style → operation-named wrapper with message parts as children + warning; additional ports beyond the first SOAP port of a service → one warning each listing them

## User Stories

1. As a developer handed a vendor WSDL, I run `reqly import wsdl service.wsdl -o ws` and get runnable POST requests for every operation.
2. As a tester, I fill two child elements in a generated envelope and send it.
3. As an integrator hitting a SOAP 1.2 endpoint, my imported request already carries the right envelope namespace and Content-Type.

## Implementation Decisions

- Hand-rolled parser over `encoding/xml` following the Swagger 2.x importer precedent; no schema-library dependency.
- Envelope generation is part of the importer output — no new runtime package.
- Collection per service, request per operation using the first SOAP port's address.
- No ADR: additive CLI surface following the established importer pattern.

## Out of Scope

- Fetching WSDLs over HTTP (file-only like all importers today)
- WS-Security, MTOM, attachments
- Response-side XML tooling
