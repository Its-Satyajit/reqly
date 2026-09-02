export const methodTint = {
	GET: "text-method-get",
	POST: "text-method-post",
	PUT: "text-method-put",
	PATCH: "text-method-put",
	DELETE: "text-method-delete",
} as const satisfies Record<string, string>;

export type MethodTintKey = keyof typeof methodTint;
