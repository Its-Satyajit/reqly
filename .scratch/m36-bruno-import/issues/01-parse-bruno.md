# 01 — Bruno JSON parser (internal/importer)

**What to build:** Parse a Bruno collection export JSON: items tree (folder/http/graphql), requests with method/url/headers/params, body per mode (json/xml/text/sparql/formUrlEncoded/multipartForm/graphql/file), auth block mapping, collection-level root.request defaults.

**Blocked by:** None

**Status:** done

- [x] Folder/request recursion preserving array order; unknown item types warned
- [x] Body modes per spec; file entries warned + skipped
- [x] Auth: basic/bearer/apikey (placement→in)/digest mapped; others warned; request overrides collection root
- [x] root.request.auth + headers → result-level collection descriptor values
- [x] Environments: variables with secret flag split into secrets/variables; disabled skipped
- [x] Scripts/assertions/tests/docs warned + skipped
- [x] Table-driven tests incl. bruno-testbench fixture
- [x] go vet/gofmt/go test green
