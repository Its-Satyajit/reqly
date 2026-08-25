// Request types + senders shared by every Reqly front-end.
//
// The shared UI never imports the Wails-generated bindings directly (they live
// in the host app and are regenerated on build). Instead, request execution is
// injected through a RequestSender: the Wails host injects a sender backed by
// the Go core, while browser dev mode uses fetchSender.

import { serializeBody, type BodyType } from './body'
import type { ResolvedVariable } from './collections'

export interface RequestHeader {
	key: string;
	value: string;
}

/** RequestRetry is the declarative retry policy on a request: automatic
 * re-sends after transient failures (network errors and 429/502/503/504 by
 * default) with computed backoff. Mirrors request.Retry in the Go core. */
export interface RequestRetry {
	/** Retries after the initial attempt; 0/undefined disables retrying. */
	count?: number;
	delayMs?: number;
	strategy?: 'fixed' | 'exponential';
	maxDelayMs?: number;
	retryOn?: number[];
}

/** RequestAuth is the resolved auth attached to an opened request. It is
 * applied silently at send — there is no auth editing UI. */
export interface RequestAuth {
	type?: string;
	config?: Record<string, string>;
}

/** A key-value row in the builder: enabled rows are sent, disabled are kept
 * but skipped; blank keys are dropped. `file` holds a Git-native relative
 * file path for `form-data` file entries (M21); `filename` overrides the
 * uploaded file name. */
export interface KeyValueRow {
	key: string;
	value: string;
	enabled: boolean;
	file?: string;
	filename?: string;
}

export interface RequestInput {
	method?: string;
	url: string;
	headers?: RequestHeader[];
	params?: KeyValueRow[];
	bodyType?: BodyType;
	/** Body text for json/xml/raw/binary body types; serialized form fields for
	 * form-data/urlencoded live in `form`. For graphql, `body` holds the query
	 * when `graphqlQuery` is not set. */
	body?: string;
	form?: KeyValueRow[];
	/** GraphQL query and variables (JSON string) for graphql body type. */
	graphqlQuery?: string;
	graphqlVariables?: string;
	timeout?: number;
	/** Automatic retry policy; absent = no retries. */
	retry?: RequestRetry;
	/** Environment pill (a request file's environment: field) used at send;
	 * empty falls back to the app's active environment. */
	env?: string;
	/** The tab's effective variable chain (scope-tagged snapshot), layered
	 * under the request so file variables win over the environment. */
	vars?: ResolvedVariable[];
	/** Workspace-relative Request Path of the file-backed tab this send belongs
	 * to. When set, the request is treated as the tab's live draft and the full
	 * inheritance chain (base URL, merged headers, inherited auth, variable
	 * scopes) is re-resolved against it at send time; when unset, the request
	 * is sent as-is (scratchpad). */
	requestPath?: string;
	/** Inherited auth resolved when the request was opened; applied silently. */
	auth?: RequestAuth;
	/** Identifies this in-flight send so Stop can cancel it via the bridge's
	 * CancelSend seam; minted per send by the store. */
	sendId?: string;
}

/** sentParams returns the enabled, non-blank rows of a param/header list. */
export function sentRows(rows: KeyValueRow[] | undefined): KeyValueRow[] {
	return (rows ?? []).filter((r) => r.enabled && r.key.trim() !== "");
}

/** appendParams appends params to a URL's query string, preserving any
 * existing query. Mirrors the engine's query merge. */
export function appendParams(url: string, params: KeyValueRow[]): string {
	const rows = sentRows(params);
	if (rows.length === 0) return url;
	const [base, existing = ""] = url.split("?", 2);
	const parts = existing ? existing.split("&") : [];
	for (const r of rows) {
		parts.push(`${encodeURIComponent(r.key)}=${encodeURIComponent(r.value)}`);
	}
	return `${base}?${parts.join("&")}`;
}

export interface ResponseData {
	statusCode: number;
	statusText: string;
	proto: string;
	headers: Record<string, string[]>;
	body: string;
	durationMs: number;
	size: number;
	ok: boolean;
	/** Sends this response took, including retries (1 or undefined = none). */
	attempts?: number;
}

export type RequestSender = (req: RequestInput) => Promise<ResponseData>;

const CONTENT_TYPE = "Content-Type";

/**
 * fetchSender runs a request in the browser (no Wails bridge). Used as the
 * default so the shared UI is usable in plain Vite dev mode.
 */
export const fetchSender: RequestSender = async (req) => {
	const started = performance.now();

	const headers: Record<string, string> = {};
	for (const h of req.headers ?? []) headers[h.key] = h.value;
	const { body: requestBodyText, contentType } = serializeBody(req);
	const hasManualType = Object.keys(headers).some(
		(k) => k.toLowerCase() === "content-type",
	);
	if (contentType && !hasManualType) headers[CONTENT_TYPE] = contentType;

	const method = req.method ?? "GET";
	// Browsers reject a body on GET/HEAD, so drop it for those methods (the Go
	// engine tolerates it; dev mode cannot).
	const requestBody =
		method === "GET" || method === "HEAD"
			? undefined
			: requestBodyText || undefined;

	// Non-2xx is a valid, expected outcome for an API client — the status and
	// body are surfaced to the user verbatim, never treated as a transport
	// failure. Only a thrown fetch error (network/DNS) is an error here.
	// react-doctor-disable-next-line react-doctor/no-fetch-response-used-without-status-check
	const res = await fetch(appendParams(req.url, req.params ?? []), {
		method,
		headers,
		body: requestBody,
	});

	const body = await res.text();
	const responseHeaders: Record<string, string[]> = {};
	res.headers.forEach((value, key) => {
		(responseHeaders[key] ??= []).push(value);
	});

	return {
		statusCode: res.status,
		statusText: res.statusText,
		proto: "",
		headers: responseHeaders,
		body,
		durationMs: Math.round(performance.now() - started),
		size: new TextEncoder().encode(body).length,
		ok: res.ok,
	};
};