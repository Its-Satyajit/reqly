# 01 — WSDL parser + workspace writer (internal/importer)

**Blocked by:** None

**Status:** done

- [x] encoding/xml model for definitions/types/messages/portTypes/bindings/services
- [x] Inline XSD processing: elements, complexType sequences, minOccurs filtering; xsd:import/include warned + skipped
- [x] Operation join: portType ↔ binding ↔ service port; SOAPAction; document/literal wrapped+bare; rpc/encoded best-effort + warning
- [x] Envelope skeletons matched to binding version (1.1/1.2 ns + Content-Type); placeholders by type
- [x] Workspace writer: reqly.yaml + collections per service + request per op (collision suffixes); warnings list
- [x] Table-driven tests incl. vendored wsdl.xml fixture; go vet/gofmt/go test green
