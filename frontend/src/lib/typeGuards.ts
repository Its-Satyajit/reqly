export type JsonValue = string | number | boolean | null | JsonValue[] | JsonObject

export type JsonObject = { [key: string]: JsonValue }

export function isString(cause: unknown): cause is string {
  return typeof cause === "string"
}

export function isRecord(cause: unknown): cause is JsonObject {
  return typeof cause === "object" && cause !== null && !Array.isArray(cause)
}
