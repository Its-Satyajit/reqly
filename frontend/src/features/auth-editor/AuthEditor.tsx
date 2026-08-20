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
  ORDERED_AUTH_SCHEMES,
  AUTH_SCHEME_LABELS,
  authForScheme,
  schemeFieldValue,
  schemeFor,
  isSensitiveKey,
  type AuthSchemeId,
} from "../../lib/authSchemes"
import type { RequestAuth } from "../../lib/request"

const selectClass =
  "rounded-md border border-input bg-background px-2 py-1.5 text-xs text-foreground"

export interface AuthEditorProps {
  auth: RequestAuth | undefined
  onChange: (auth: RequestAuth | undefined) => void
}

export function AuthEditor({ auth, onChange }: AuthEditorProps) {
  const scheme = schemeFor(auth)
  const fields = AUTH_SCHEMES.find((s) => s.id === scheme)?.fields ?? []

  const setScheme = (next: AuthSchemeId) => onChange(authForScheme(next, auth))

  const setField = (key: string, value: string) => {
    onChange({ type: scheme, config: { ...(auth?.config ?? {}), [key]: value } })
  }

  return (
    <div className="flex flex-col gap-4">
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
        <p className="text-xs text-muted-foreground">
          This request declares no own auth and will inherit from its
          workspace / collection / folder. No auth block is written on save.
        </p>
      ) : scheme === "none" ? (
        <p className="text-xs text-muted-foreground">
          Sends unauthenticated, even under an auth-bearing collection or
          folder. Saves <code className="rounded bg-muted/40 px-1">auth.type: none</code>.
        </p>
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