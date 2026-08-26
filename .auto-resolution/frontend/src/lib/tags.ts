export const DYNAMIC_TAGS = ["uuid", "timestamp", "isoTimestamp", "randomInt", "randomString"] as const
export type DynamicTag = typeof DYNAMIC_TAGS[number]

export const TAG_REGEX = /\{\{\$([^}\s]+)(?:\s+([^}]*))?\}\}/g

export function unknownTags(input: string): string[] {
  const re = new RegExp(TAG_REGEX.source, "g")
  const seen = new Set<string>()
  const out: string[] = []
  let m: RegExpExecArray | null
  while ((m = re.exec(input)) !== null) {
    const tag = m[1]
    // SAFETY: DYNAMIC_TAGS is the closed set; unknown tags are exactly those not in it.
    if (!DYNAMIC_TAGS.includes(tag as DynamicTag) && !seen.has(tag)) {
      seen.add(tag)
      out.push(tag)
    }
  }
  return out
}

export function tagWarnings(input: string): string[] {
  const unknowns = unknownTags(input)
  return unknowns.map((t) => "Unknown dynamic tag {{$" + t + "}} will be sent as-is.")
}
