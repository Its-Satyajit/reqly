// Auth scheme metadata shared by every Reqly front-end.
//
// The scheme set mirrors the core's auth registry (`internal/auth`): each
// scheme's `fields` are the flat `config` keys its Apply/SecretKeys expect,
// so the Auth tab renders the same shape a request file carries. `inherit`
// and `none` are the two non-credential states: Inherit declares no own auth
// (the request receives its containers'), No Auth writes `auth.type: none`
// to explicitly disable inherited auth.

import type { RequestAuth } from "./request"

export type AuthSchemeId =
	| "inherit"
	| "none"
	| "basic"
	| "bearer"
	| "apikey"
	| "jwt"
	| "digest"
	| "oauth2"

export interface AuthField {
	/** The flat config key the scheme's core Apply reads. */
	key: string
	label: string
	placeholder?: string
	/** Sensitive value (mirrors the core scheme's SecretKeys) — flagged in the
	 * UI and masked by the core at send/output. */
	secret?: boolean
	/** Enum select options; a plain text input otherwise. */
	options?: string[]
	help?: string
}

export interface AuthScheme {
	id: Exclude<AuthSchemeId, "inherit" | "none">
	label: string
	fields: AuthField[]
}

export const AUTH_SCHEMES: AuthScheme[] = [
	{
		id: "basic",
		label: "Basic",
		fields: [
			{ key: "username", label: "Username", placeholder: "user@example.com" },
			{ key: "password", label: "Password", placeholder: "••••••••", secret: true },
		],
	},
	{
		id: "bearer",
		label: "Bearer",
		fields: [
			{ key: "token", label: "Token", placeholder: "token", secret: true },
		],
	},
	{
		id: "apikey",
		label: "API Key",
		fields: [
			{ key: "key", label: "Key", placeholder: "X-API-Key" },
			{ key: "value", label: "Value", placeholder: "api key", secret: true },
			{
				key: "in",
				label: "Send as",
				options: ["header", "query"],
				help: "Header or query parameter; defaults to header.",
			},
		],
	},
	{
		id: "jwt",
		label: "JWT",
		fields: [
			{ key: "secret", label: "Secret", placeholder: "shared secret", secret: true },
			{
				key: "algorithm",
				label: "Algorithm",
				options: ["HS256", "HS384", "HS512"],
				help: "Defaults to HS256.",
			},
			{
				key: "claims",
				label: "Claims (JSON)",
				placeholder: '{"sub":"user","iss":"reqly"}',
			},
			{ key: "expiresIn", label: "Expires in", placeholder: "5m", help: "Go duration, e.g. 5m or 1h." },
		],
	},
	{
		id: "digest",
		label: "Digest",
		fields: [
			{ key: "username", label: "Username", placeholder: "user@example.com" },
			{ key: "password", label: "Password", placeholder: "••••••••", secret: true },
			{ key: "realm", label: "Realm", placeholder: "challenge realm (optional)" },
			{
				key: "algorithm",
				label: "Algorithm",
				options: ["MD5", "SHA-256"],
				help: "Fallback when the server's challenge omits one.",
			},
		],
	},
]

/** AUTH_SCHEME_LABELS maps every picker state (including Inherit and No Auth)
 * to its display label. */
export const AUTH_SCHEME_LABELS: Record<AuthSchemeId, string> = {
	inherit: "Inherit",
	none: "No Auth",
	basic: "Basic",
	bearer: "Bearer",
	apikey: "API Key",
	jwt: "JWT",
	digest: "Digest",
	oauth2: "OAuth 2.0",
}

/** ORDERED_AUTH_SCHEMES is the picker order: the two non-credential states
 * first, then the typed schemes. */
export const ORDERED_AUTH_SCHEMES: AuthSchemeId[] = [
	"inherit",
	"none",
	"basic",
	"bearer",
	"apikey",
	"jwt",
	"digest",
	"oauth2",
]

/** schemeFor maps a request's own auth onto the picker state. */
export const schemeFor = (auth?: RequestAuth): AuthSchemeId => {
	if (!auth || !auth.type) return "inherit"
	return auth.type as AuthSchemeId
}

/** authForScheme builds the request's own auth for the given picker state,
 * preserving an existing config where the config keys still apply. Inherit
 * yields no own auth; No Auth yields the explicit `none` block. */
export const authForScheme = (
	scheme: AuthSchemeId,
	prev?: RequestAuth,
): RequestAuth | undefined => {
	if (scheme === "inherit") return undefined
	if (scheme === "none") return { type: "none" }
	return {
		type: scheme,
		config: prev?.type === scheme ? prev.config : {},
	}
}

/** schemeFieldValue reads a config key, normalizing undefined to "". */
export const schemeFieldValue = (
	auth: RequestAuth | undefined,
	key: string,
): string => auth?.config?.[key] ?? ""

/** isSensitiveKey reports whether a config key is a secret for the given
 * scheme (mirrors the core scheme's SecretKeys). */
export const isSensitiveKey = (scheme: AuthSchemeId, key: string): boolean => {
	switch (scheme) {
		case "basic":
		case "digest":
			return key === "password"
		case "bearer":
			return key === "token"
		case "apikey":
			return key === "value"
		case "jwt":
			return key === "secret"
		case "oauth2":
			return key === "client_secret"
		default:
			return false
	}
}