export interface JwtExpiry {
  status: "expired" | "not_yet_valid" | "valid" | "no_expiry";
  remaining: number;
  exp?: number;
  nbf?: number;
  iat?: number;
}

/** JwtClaim is one decoded claim rendered as a display string by the bridge. */
export interface JwtClaim {
  key: string;
  value: string;
}

export interface JwtTokenView {
  header: JwtClaim[];
  payload: JwtClaim[];
  signature: string;
  alg: string;
  expiry: JwtExpiry;
}

export interface JwtAdapter {
  decode(token: string): Promise<JwtTokenView>;
}

// Bridge registry, same pattern as the diff adapter.
let jwtBridge: JwtAdapter | null = null;

/** setJwtBridge installs the host adapter (called once from the bridge). */
export function setJwtBridge(adapter: JwtAdapter): void {
  jwtBridge = adapter;
}

/** getJwtBridge returns the installed adapter or a throwing fallback. */
export function getJwtBridge(): JwtAdapter {
  return jwtBridge ?? fallbackJwtAdapter;
}

export const fallbackJwtAdapter: JwtAdapter = {
  async decode() {
    throw new Error("jwt inspector is not available in this build");
  },
};

/** formatRemaining renders seconds as a coarse human duration. */
export function formatRemaining(seconds: number): string {
  if (seconds <= 0) return "expired";
  const units: [number, string][] = [
    [86400, "d"],
    [3600, "h"],
    [60, "m"],
  ];
  for (const [size, suffix] of units) {
    if (seconds >= size) return `${Math.floor(seconds / size)}${suffix}`;
  }
  return `${seconds}s`;
}

/** extractBearer pulls a bearer token from an Authorization header value. */
export function extractBearer(value?: string | null): string | null {
  if (!value) return null;
  const trimmed = value.trim();
  if (/^bearer /i.test(trimmed)) {
    const token = trimmed.slice(7).trim();
    return token === "" ? null : token;
  }
  return null;
}
