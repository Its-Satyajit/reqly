// JSONPath querying for the response viewer. Dependency-free evaluator
// matching the core's assertion JSONPath subset (internal/response/jsonpath.go)
// plus wildcard support: `$`, dot and bracket segments, `*`, and array indexes.
// Kept as plain functions so they can be exercised without a component tree.

import { isRecord, type JsonObject, type JsonValue } from "./typeGuards"

export interface JSONPathMatch {
  /** Canonical path of the match, e.g. `$.users[0].name`. */
  path: string
  value: JsonValue
}

export interface JSONPathResult {
  matches: JSONPathMatch[]
  /** Specific, user-facing error; undefined when the path parsed cleanly.
   * Zero matches is not an error. */
  error?: string
}

type Step =
  | { kind: 'key'; key: string; at: number }
  | { kind: 'wildcard'; at: number }

/** queryJSONPath evaluates a JSONPath against a parsed JSON value. Returns
 * every node the path selects (wildcards fan out) with its canonical path, or
 * a specific error for malformed paths. Missing values yield zero matches,
 * never an error. */
export function queryJSONPath(root: JsonValue, path: string): JSONPathResult {
  const parsed = parsePath(path)
  if ('error' in parsed) return { matches: [], error: parsed.error }

  const matches: JSONPathMatch[] = []
  walk(root, parsed.steps, 0, '$', matches)
  return { matches }
}

function walk(
  node: JsonValue,
  steps: Step[],
  i: number,
  currentPath: string,
  matches: JSONPathMatch[],
): void {
  if (i >= steps.length) {
    matches.push({ path: currentPath, value: node })
    return
  }

  const step = steps[i]
  if (step.kind === 'wildcard') {
    if (Array.isArray(node)) {
      node.forEach((item, idx) => {
        walk(item, steps, i + 1, `${currentPath}[${idx}]`, matches)
      })
      return
    }
    if (isRecord(node)) {
      // SAFETY: isRecord narrows JsonValue to JsonObject with concrete JsonValue values
      const obj = node as JsonObject
      for (const key of Object.keys(obj)) {
        // SAFETY: JsonObject values are JsonValue; key existence validated via Object.keys
        walk(
          obj[key] as JsonValue,
          steps,
          i + 1,
          `${currentPath}.${key}`,
          matches,
        )
      }
      return
    }
    return
  }

  if (Array.isArray(node)) {
    const idx = Number(step.key)
    if (!Number.isInteger(idx) || idx < 0 || idx >= node.length) return
    // SAFETY: array bounds checked above; element is JsonValue per JsonValue definition
    walk(node[idx] as JsonValue, steps, i + 1, `${currentPath}[${idx}]`, matches)
    return
  }
  if (isRecord(node)) {
    // SAFETY: isRecord guarantees JsonObject with JsonValue values; missing key yields undefined
    const value = (node as JsonObject)[step.key]
    if (value === undefined) return
    walk(value, steps, i + 1, `${currentPath}.${step.key}`, matches)
    return
  }
  // Primitive node with segments remaining — nothing to descend into.
}

function parsePath(path: string): { steps: Step[] } | { error: string } {
  let p = path.trim()
  if (p === '') return { error: 'path is empty — try `$.users[0].name`' }
  if (p.startsWith('$')) p = p.slice(1)

  const steps: Step[] = []
  while (p.length > 0) {
    p = p.trim().replace(/^\.+/, '').trim()
    if (p === '') break

    const at = steps.length + 1
    if (p.startsWith('[')) {
      const end = p.indexOf(']')
      if (end < 0) {
        return { error: `invalid path at segment ${at}: unclosed \`[\`` }
      }
      const inner = p.slice(1, end).trim()
      if (inner === '') {
        return { error: `invalid path at segment ${at}: empty brackets` }
      }
      const unquoted = stripQuotes(inner)
      if (unquoted === '') {
        return { error: `invalid path at segment ${at}: empty key` }
      }
      if (unquoted === '*') steps.push({ kind: 'wildcard', at })
      else steps.push({ kind: 'key', key: unquoted, at })
      p = p.slice(end + 1)
      continue
    }

    const key = (p.match(/^[^.[]+/)?.[0] ?? '').trim()
    if (key === '') {
      return { error: `invalid path at segment ${at}: expected a key or \`[...]\`` }
    }
    if (key === '*') steps.push({ kind: 'wildcard', at })
    else steps.push({ kind: 'key', key, at })
    p = p.slice(key.length)
  }
  return { steps }
}

function stripQuotes(s: string): string {
  if (
    s.length >= 2 &&
    ((s.startsWith('"') && s.endsWith('"')) ||
      (s.startsWith("'") && s.endsWith("'")))
  ) {
    return s.slice(1, -1)
  }
  return s
}
