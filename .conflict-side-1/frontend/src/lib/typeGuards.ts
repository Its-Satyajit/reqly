// Shared domain guards for I/O-boundary parsing. Centralizing `typeof`
// checks here keeps the lint rule `no-runtime-typeof` scoped to type
// predicates (allowed via `allowInTypeGuards`).

export type JsonValue = string | number | boolean | null | JsonValue[] | JsonObject

export type JsonObject = { [key: string]: JsonValue }

export function isString(cause: unknown): cause is string {
  return typeof cause === "string"
}

export function isNumber(cause: unknown): cause is number {
  return typeof cause === "number"
}

export function isRecord(cause: unknown): cause is JsonObject {
  return typeof cause === "object" && cause !== null && !Array.isArray(cause)
}

export function isObject(cause: unknown): cause is JsonObject {
  return typeof cause === "object" && cause !== null && !Array.isArray(cause)
}

export function isDefinedString(cause: string | undefined): cause is string {
  return typeof cause === "string"
}

export type AnyFunction = (...args: never[]) => void;

export function isFunction(cause: unknown): cause is AnyFunction {
  return typeof cause === "function"
}
