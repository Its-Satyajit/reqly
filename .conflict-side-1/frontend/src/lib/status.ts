/** Status/method → class mappings shared by the status pill components and
 * non-UI callers (runners, history). Kept out of status.tsx so that file only
 * exports components. */

export type StatusTier = "info" | "ok" | "redirect" | "warn" | "error";

/** Tailwind keeps literal classes only — the ramp is enumerated, not interpolated. */
export const STATUS_CLASSES = {
	info: "text-status-info bg-status-info/10 border-status-info/25",
	ok: "text-status-ok bg-status-ok/10 border-status-ok/25",
	redirect: "text-status-redirect bg-status-redirect/10 border-status-redirect/25",
	warn: "text-status-warn bg-status-warn/10 border-status-warn/25",
	error: "text-status-error bg-status-error/10 border-status-error/25",
} as const;

export const DOT_CLASSES = {
	info: "bg-status-info",
	ok: "bg-status-ok",
	redirect: "bg-status-redirect",
	warn: "bg-status-warn",
	error: "bg-status-error",
} as const;

export const METHOD_CLASSES = {
	GET: "text-method-get",
	POST: "text-method-post",
	PUT: "text-method-put",
	PATCH: "text-method-put",
	DELETE: "text-method-delete",
	HEAD: "text-status-info",
	OPTIONS: "text-status-info",
} as const;

export function statusTier(status: number | null | undefined): StatusTier {
	if (status == null || status === 0) return "error";
	if (status < 200) return "info";
	if (status < 300) return "ok";
	if (status < 400) return "redirect";
	if (status < 500) return "warn";
	return "error";
}

/** Semantic tint class for an HTTP method, per the GitHub REST-doc ramp. */
export function methodTintClass(method: string): string {
	return (
		// SAFETY: unknown methods miss the lookup and fall back to muted text
		METHOD_CLASSES[method.toUpperCase() as keyof typeof METHOD_CLASSES] ??
		"text-muted-foreground"
	);
}
