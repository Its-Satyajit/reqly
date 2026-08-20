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
	/** Missing required fields surface as non-blocking save warnings. Defaults
	 * to required; mark optional for fields the core defaults or tolerates. */
	optional?: boolean
}

export interface AuthScheme {
	id: Exclude<AuthSchemeId, "inherit" | "none">
	label: string
	fields: AuthField[]
}

export interface OAuth2Grant {
	id: string
	label: string
	fields: AuthField[]
}

export const OAUTH2_GRANTS: OAuth2Grant[] = [
	{
		id: "client_credentials",
		label: "Client Credentials",
		fields: [
			{ key: "token_url", label: "Token URL", placeholder: "https://idp.example.com/token" },
			{ key: "client_id", label: "Client ID", placeholder: "client id" },
			{ key: "client_secret", label: "Client Secret", placeholder: "••••••••", secret: true },
			{ key: "scope", label: "Scope", placeholder: "openid profile", optional: true },
			{ key: "audience", label: "Audience", placeholder: "api://default", optional: true },
			{ key: "token_name", label: "Token name", placeholder: "default", optional: true },
		],
	},
	{
		id: "authorization_code",
		label: "Authorization Code + PKCE",
		fields: [
			{ key: "authorization_url", label: "Authorization URL", placeholder: "https://idp.example.com/authorize" },
			{ key: "token_url", label: "Token URL", placeholder: "https://idp.example.com/token" },
			{ key: "client_id", label: "Client ID", placeholder: "client id" },
			{ key: "client_secret", label: "Client Secret", placeholder: "••••••••", secret: true },
			{ key: "redirect_uri", label: "Redirect URI", placeholder: "https://app.example.com/callback", optional: true },
			{ key: "scope", label: "Scope", placeholder: "openid profile", optional: true },
			{ key: "audience", label: "Audience", placeholder: "api://default", optional: true },
			{ key: "token_name", label: "Token name", placeholder: "default", optional: true },
		],
	},
	{
		id: "device_code",
		label: "Device Code",
		fields: [
			{ key: "device_authorization_url", label: "Device Authorization URL", placeholder: "https://idp.example.com/device" },
			{ key: "token_url", label: "Token URL", placeholder: "https://idp.example.com/token" },
			{ key: "client_id", label: "Client ID", placeholder: "client id" },
			{ key: "client_secret", label: "Client Secret", placeholder: "••••••••", secret: true },
			{ key: "scope", label: "Scope", placeholder: "openid profile", optional: true },
			{ key: "audience", label: "Audience", placeholder: "api://default", optional: true },
			{ key: "token_name", label: "Token name", placeholder: "default", optional: true },
		],
	},
]

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
				optional: true,
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
				optional: true,
			},
			{
				key: "claims",
				label: "Claims (JSON)",
				placeholder: '{"sub":"user","iss":"reqly"}',
				optional: true,
			},
			{ key: "expiresIn", label: "Expires in", placeholder: "3600", help: "Number of seconds the token stays valid (integer).", optional: true },
		],
	},
	{
		id: "digest",
		label: "Digest",
		fields: [
			{ key: "username", label: "Username", placeholder: "user@example.com" },
			{ key: "password", label: "Password", placeholder: "••••••••", secret: true },
			{ key: "realm", label: "Realm", placeholder: "challenge realm (optional)", optional: true },
			{
				key: "algorithm",
				label: "Algorithm",
				options: ["MD5", "SHA-256"],
				help: "Fallback when the server's challenge omits one.",
				optional: true,
			},
		],
	},
	{
		// The oauth2 fields are grant-driven; see OAUTH2_GRANTS.
		id: "oauth2",
		label: "OAuth 2.0",
		fields: [],
	},
]

/** AUTH_SCHEME_LABELS maps every picker state (including Inherit and No Auth)
 * to its display label. */
export const AUTH_SCHEME_LABELS = {
	inherit: "Inherit",
	none: "No Auth",
	basic: "Basic",
	bearer: "Bearer",
	apikey: "API Key",
	jwt: "JWT",
	digest: "Digest",
	oauth2: "OAuth 2.0",
} satisfies Record<AuthSchemeId, string>

/** DEFAULT_OAUTH2_GRANT is used when a request's oauth2 config has not pinned
 * grant_type (the core default). */
export const DEFAULT_OAUTH2_GRANT = "client_credentials"

/** oauth2GrantFor resolves a request's oauth2 config onto a grant. */
export const oauth2GrantFor = (auth: RequestAuth | undefined): string => {
	const granted = auth?.config?.grant_type
	return granted && OAUTH2_GRANTS.some((g) => g.id === granted)
		? granted
		: DEFAULT_OAUTH2_GRANT
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
 * scheme. It derives from the scheme's field metadata (`secret: true`),
 * which mirrors the core scheme's SecretKeys. */
export const isSensitiveKey = (scheme: AuthSchemeId, key: string): boolean => {
	if (scheme === "oauth2") {
		return OAUTH2_GRANTS.some((g) => g.fields.some((f) => f.key === key && f.secret))
	}
	const s = AUTH_SCHEMES.find((x) => x.id === scheme)
	return s?.fields.some((f) => f.key === key && f.secret) ?? false
}

/** authWarnings lists the non-blocking save warnings for a request's own
 * auth: required config fields that are still empty. Inherit and No Auth
 * never warn. */
export const authWarnings = (auth: RequestAuth | undefined): string[] => {
	if (!auth || !auth.type) return []
	if (auth.type === "none") return []
	const scheme = AUTH_SCHEMES.find((s) => s.id === auth.type)
	if (!scheme) return []
	const fields =
		auth.type === "oauth2"
			? (OAUTH2_GRANTS.find((g) => g.id === oauth2GrantFor(auth))?.fields ?? [])
			: scheme.fields
	const warnings: string[] = []
	for (const field of fields) {
		if (field.optional) continue
		if (!auth.config?.[field.key]) {
			warnings.push(`${field.label} is required for ${scheme.label} auth`)
		}
	}
	return warnings
}