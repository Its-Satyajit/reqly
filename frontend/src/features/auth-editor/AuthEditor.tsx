// AuthEditor is the request editor's Auth tab: a scheme picker (Inherit /
// No Auth / the typed schemes) with per-scheme field forms. It is a
// controlled form over the request's own auth — every change flows through
// onChange into the tab's draft, so auth is dirty-tracked, saved to the
// file, and used at send like any other builder field.

import { Field } from "@base-ui/react/field"
import { Input } from "../../components/ui/input"
import { Label } from "../../components/ui/field"
import {
  AUTH_SCHEMES,
  OAUTH2_GRANTS,
  ORDERED_AUTH_SCHEMES,
  AUTH_SCHEME_LABELS,
  authForScheme,
  authWarnings,
  schemeFieldValue,
  schemeFor,
  oauth2GrantFor,
  isSensitiveKey,
  type AuthSchemeId,
} from "../../lib/authSchemes"
import type { RequestAuth } from "../../lib/request"

const selectClass =
  "rounded-md border border-input bg-background px-2 py-1.5 text-xs text-foreground"

export interface AuthEditorProps {
  auth: RequestAuth | undefined
  onChange: (auth: RequestAuth | undefined) => void
  /** The effective auth the request will use once inheritance is applied
   * (read-only, shown under the Inherit state). */
  inherited?: RequestAuth
}

/** scrollToAuthPanel focuses the sidebar's OAuth tokens panel. */
const scrollToAuthPanel = () => {
  document
    .getElementById("auth-panel")
    ?.scrollIntoView({ behavior: "smooth", block: "nearest" })
}

export function AuthEditor({ auth, onChange, inherited }: AuthEditorProps) {
  const scheme = schemeFor(auth)
  const fields = AUTH_SCHEMES.find((s) => s.id === scheme)?.fields ?? []
  const warnings = authWarnings(auth)

  const setScheme = (next: AuthSchemeId) => onChange(authForScheme(next, auth))

  const setField = (key: string, value: string) => {
    onChange({ type: scheme, config: { ...(auth?.config ?? {}), [key]: value } })
  }

  return (
    <div className="flex flex-col gap-4">
      {warnings.length > 0 ? (
        <div className="rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-xs text-amber-700">
          <p className="font-medium">Auth needs attention before send</p>
          <ul className="mt-1 list-disc pl-4">
            {warnings.map((w) => (
              <li key={w}>{w}</li>
            ))}
          </ul>
        </div>
      ) : null}

      <div className="flex flex-col gap-1">
        <Label>Auth type</Label>
        <select
          value={scheme}
          onChange={(e) => setScheme(e.target.value as AuthSchemeId)}
          className={`${selectClass} w-full sm:w-48`}
        >
          {ORDERED_AUTH_SCHEMES.map((id) => (
            <option key={id} value={id}>
              {AUTH_SCHEME_LABELS[id]}
            </option>
          ))}
        </select>
      </div>

      {scheme === "inherit" ? (
        <InheritedAuth inherited={inherited} />
      ) : scheme === "none" ? (
        <p className="text-xs text-muted-foreground">
          Sends unauthenticated, even under an auth-bearing collection or
          folder. Saves <code className="rounded bg-muted/40 px-1">auth.type: none</code>.
        </p>
      ) : scheme === "oauth2" ? (
        <OAuth2Fields auth={auth} onChange={onChange} />
      ) : (
        <div className="flex flex-col gap-3">
          {fields.map((field) => (
            <Field.Root key={field.key} className="flex flex-col gap-1">
              <Label className="flex items-center gap-1.5">
                {field.label}
                {isSensitiveKey(scheme, field.key) && (
                  <span className="rounded-sm border border-amber-500/30 bg-amber-500/10 px-1 text-[9px] font-medium uppercase tracking-wide text-amber-600">
                    secret
                  </span>
                )}
              </Label>
              {field.options ? (
                <select
                  value={schemeFieldValue(auth, field.key) || field.options[0]}
                  onChange={(e) => setField(field.key, e.target.value)}
                  className={`${selectClass} w-full sm:w-48`}
                >
                  {field.options.map((opt) => (
                    <option key={opt} value={opt}>
                      {opt}
                    </option>
                  ))}
                </select>
              ) : (
                <Input
                  value={schemeFieldValue(auth, field.key)}
                  onChange={(e) => setField(field.key, e.target.value)}
                  placeholder={field.placeholder}
                  type={field.secret ? "password" : "text"}
                  spellCheck={false}
                  autoComplete="off"
                />
              )}
              {field.help ? (
                <p className="text-[11px] text-muted-foreground/70">{field.help}</p>
              ) : null}
            </Field.Root>
          ))}
        </div>
      )}
    </div>
  )
}

/** OAuth2Fields is the grant-type-driven form for the oauth2 scheme. The
 * grant config lives in the request's own auth; the token lifecycle (login /
 * refresh / logout) lives in the sidebar Auth Panel. */
function OAuth2Fields({
  auth,
  onChange,
}: {
  auth: RequestAuth | undefined
  onChange: (auth: RequestAuth | undefined) => void
}) {
  const grant = oauth2GrantFor(auth)
  const fields = OAUTH2_GRANTS.find((g) => g.id === grant)?.fields ?? []

  const setField = (key: string, value: string) => {
    onChange({ type: "oauth2", config: { ...(auth?.config ?? {}), [key]: value } })
  }

  return (
    <div className="flex flex-col gap-3">
      <Field.Root className="flex flex-col gap-1">
        <Label>Grant type</Label>
        <select
          value={grant}
          onChange={(e) => setField("grant_type", e.target.value)}
          className={`${selectClass} w-full sm:w-56`}
        >
          {OAUTH2_GRANTS.map((g) => (
            <option key={g.id} value={g.id}>
              {g.label}
            </option>
          ))}
        </select>
      </Field.Root>

      {fields.map((field) => (
        <Field.Root key={field.key} className="flex flex-col gap-1">
          <Label className="flex items-center gap-1.5">
            {field.label}
            {isSensitiveKey("oauth2", field.key) && (
              <span className="rounded-sm border border-amber-500/30 bg-amber-500/10 px-1 text-[9px] font-medium uppercase tracking-wide text-amber-600">
                secret
              </span>
            )}
          </Label>
          <Input
            value={schemeFieldValue(auth, field.key)}
            onChange={(e) => setField(field.key, e.target.value)}
            placeholder={field.placeholder}
            type={field.secret ? "password" : "text"}
            spellCheck={false}
            autoComplete="off"
          />
          {field.help ? (
            <p className="text-[11px] text-muted-foreground/70">{field.help}</p>
          ) : null}
        </Field.Root>
      ))}

      <button
        type="button"
        onClick={scrollToAuthPanel}
        className="self-start text-xs text-muted-foreground underline-offset-2 transition-colors hover:text-foreground hover:underline"
      >
        Log in / manage tokens in the sidebar OAuth panel →
      </button>
    </div>
  )
}

/** InheritedAuth is the read-only view shown under the Inherit state: the
 * auth the request will actually use once its workspace/collection/folder
 * chain is applied. */
function InheritedAuth({ inherited }: { inherited: RequestAuth | undefined }) {
  if (!inherited || !inherited.type) {
    return (
      <p className="text-xs text-muted-foreground">
        No inherited auth — this request will send unauthenticated. No auth
        block is written on save.
      </p>
    )
  }

  const schemeId = inherited.type as AuthSchemeId
  const label = AUTH_SCHEME_LABELS[schemeId] ?? inherited.type
  const hasSecrets = Object.keys(inherited.config ?? {}).some((k) =>
    isSensitiveKey(schemeId, k),
  )
  const publicValues = Object.entries(inherited.config ?? {})
    .filter(([k]) => k !== "grant_type" && !isSensitiveKey(schemeId, k))
    .map(([k, v]) => `${k}: ${v}`)

  return (
    <div className="rounded-md border border-border bg-muted/30 p-3">
      <p className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
        Inherited from workspace / collection / folder
      </p>
      <p className="mt-1 text-xs text-foreground">
        {label}
        {publicValues.length > 0 ? ` · ${publicValues.join(", ")}` : ""}
      </p>
      {hasSecrets ? (
        <p className="mt-1 text-[11px] text-muted-foreground/70">
          Secret values are masked and only applied at send.
        </p>
      ) : null}
    </div>
  )
}