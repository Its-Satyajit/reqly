# ADR 0011: Desktop Request Auth Editing

## Status
Accepted

## Context
The desktop request editor can edit builder fields (url/method/headers/query/body) but not auth: `mergedBuilderRequest` preserves a file's `Auth` verbatim on save, and sends apply the open-time inherited auth silently (features.md: "no auth editing UI yet"). The roadmap's open follow-up across M16–M18 is a request-level auth editor. The design question is how auth becomes editable without breaking the Git-native file contract, the ADR 0009 draft model, or secret handling.

## Decision
1. **Auth is an editable draft field.** The **Auth** tab joins Params/Headers/Body/Variables in the request editor. `TabDraft` carries the request's own auth; it is dirty-tracked and written to the request file on save, so `mergedBuilderRequest` and the bridge save path are amended to pass it.
2. **Inherit is distinct from No Auth.** The scheme picker exposes **Inherit** (no auth block — the request inherits from its container chain) separately from **No Auth** (`auth.type: none`, which explicitly disables inherited auth). Saving **Inherit** *removes* any existing auth block from the file, so the file truly declares none; saving **No Auth** writes `auth.type: none`. This keeps the editor WYSIWYG with the CLI/file model.
3. **Sends use the live draft auth.** File-backed tabs pass the draft's auth through the existing `ResolveSend(path, draft)` seam, which re-resolves inheritance against it (Inherit → inherited auth applies; No Auth/typed → overrides). The open-time resolved auth retires from the send path but remains the read-only **Inherited Auth** display.
4. **OAuth split.** The Auth tab edits the request's OAuth grant config (grant_type, token_url, authorization_url, device_authorization_url, client_id, client_secret, redirect_uri, scope, audience, token_name) on the file; token acquisition/login stays in the sidebar Auth Panel and the token store. The tab links to the panel rather than embedding token lifecycle.
5. **Plaintext, sensitive-flagged inputs.** Auth config values are plaintext editable fields (the Git-native file already holds them); fields mirroring the core schemes' `SecretKeys()` (token, password, apikey.value, jwt secret, client_secret) are flagged visually. Masking at send/output remains the core's job.
6. **Non-blocking validation.** Missing required config produces save warnings (extending the existing save-warnings pattern) and surfaces the core's per-scheme errors at send; nothing hard-blocks a save.
7. **Scheme parity.** Every core-registered scheme is editable: none, basic, bearer, apikey, jwt (HS256/384/512), digest, and oauth2 (client_credentials / authorization_code / device_code), via per-scheme typed field forms; oauth2 fields are driven by `grant_type`.

## Considered Options
- **Keep auth uneditable; separate management surface** — rejected: the recurring follow-up is a request-level editor, and a separate surface would fragment the file-editing model.
- **Masked secret inputs, blank = unchanged (Environment-Draft style)** — rejected: request auth values are visibly in the Git-tracked file, so plaintext editing is honest; the masked pattern exists for environment secrets that should never read back into the UI.
- **Re-resolve from disk at send** — already rejected in ADR 0009; unchanged here.

## Consequences
- **Positive:** request-level auth is finally editable end-to-end with full CLI parity; the WYSIWYG contract extends to auth (choosing Inherit deletes the block, the diff shows it); secrets stay masked in output via the existing core masking.
- **Trade-off:** a file that held an auth block loses it when the user saves with Inherit — intended, but a visible Git diff; OAuth grant config edits are plaintext in the file like every other scheme.