# 02 — Inspector + instance generator (internal/jsonschema)

**Blocked by:** 01 (shares compile step)

**Status:** done

- [x] Inspect: tree walk — type, required marker, inline constraint summary, resolved $ref paths shown once, cycle-safe
- [x] Generate: deterministic synthesis honoring const > enum[0] > example > default > synthesized; min/max/multipleOf; minLength/maxLength; minItems; known formats (email/date-time/uri/uuid/ipv4/hostname); unknown formats fall back
- [x] pattern without hints → "string" + warning; recursive $ref depth cap 8 + warning; allOf shallow merge, anyOf/oneOf first branch, not ignored + warning
- [x] --seed varies choices; --optional includes optional properties deterministically
- [x] Warnings collected, never fatal
- [x] Table-driven tests for every rule above
