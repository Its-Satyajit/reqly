export interface RecentEndpoint {
  url: string;
  kind: "ws" | "sse";
  lastSeen: number;
}

export const RECENTS_CAP = 12;

export function upsertRecents(
  recents: RecentEndpoint[],
  url: string,
  kind: "ws" | "sse",
): RecentEndpoint[] {
  const trimmed = url.trim();
  if (!trimmed) return recents;
  const filtered = recents.filter((r) => !(r.url === trimmed && r.kind === kind));
  const next = [{ url: trimmed, kind, lastSeen: Date.now() }, ...filtered];
  return next.slice(0, RECENTS_CAP);
}

export function recentsForKind(recents: RecentEndpoint[], kind: "ws" | "sse"): RecentEndpoint[] {
  return recents.filter((r) => r.kind === kind);
}
