export interface ProxyConfig {
  enabled: boolean;
  type: "http" | "https" | "socks5";
  host: string;
  port: number;
  auth?: { username: string; password: string };
  bypassList: string[];
}

export interface TlsConfig {
  verifyPeer: boolean;
  verifyHostnames: boolean;
  customCaPath?: string;
  clientCertPath?: string;
  clientKeyPath?: string;
  minVersion: "1.0" | "1.1" | "1.2" | "1.3";
  maxVersion?: "1.0" | "1.1" | "1.2" | "1.3";
  cipherSuites?: string[];
}

export const DEFAULT_PROXY: ProxyConfig = {
  enabled: false,
  type: "http",
  host: "",
  port: 8080,
  bypassList: [],
};

export const DEFAULT_TLS: TlsConfig = {
  verifyPeer: true,
  verifyHostnames: true,
  minVersion: "1.2",
};

const VERSION_ORDER = { "1.0": 0, "1.1": 1, "1.2": 2, "1.3": 3 } as const;

export function validateProxy(config: ProxyConfig): string[] {
  if (!config.enabled) return [];
  const errors: string[] = [];
  if (!config.host.trim()) errors.push("Host is required");
  if (config.port < 1 || config.port > 65535) errors.push("Port must be 1–65535");
  return errors;
}

export function validateTls(config: TlsConfig): string[] {
  const errors: string[] = [];
  if (!config.verifyPeer) errors.push("Peer verification is disabled — insecure in production");
  if (!config.verifyHostnames) errors.push("Hostname verification is disabled — insecure in production");
  if (config.maxVersion) {
    // SAFETY: minVersion and maxVersion are constrained to "1.0"|"1.1"|"1.2"|"1.3" — same keys as VERSION_ORDER.
    const min = VERSION_ORDER[config.minVersion as keyof typeof VERSION_ORDER] ?? 0;
    // SAFETY: same constraint as minVersion — both are TLS version string literals.
    const max = VERSION_ORDER[config.maxVersion as keyof typeof VERSION_ORDER] ?? 3;
    if (min > max) errors.push("Min TLS version must be ≤ max TLS version");
  }
  return errors;
}

export function formatProxyUrl(config: ProxyConfig): string {
  const scheme = config.type === "socks5" ? "socks5" : config.type;
  let url = `${scheme}://${config.host}:${config.port}`;
  if (config.auth) {
    url = `${scheme}://${config.auth.username}:${config.auth.password}@${config.host}:${config.port}`;
  }
  return url;
}
