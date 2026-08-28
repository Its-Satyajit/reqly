export type ProxyType = "http" | "https" | "socks5";
export function isProxyType(v: string): v is ProxyType {
  return v === "http" || v === "https" || v === "socks5";
}

export type TlsVersion = "1.0" | "1.1" | "1.2" | "1.3";
export function isTlsVersion(v: string): v is TlsVersion {
  return v === "1.0" || v === "1.1" || v === "1.2" || v === "1.3";
}

export interface ProxyConfig {
  enabled: boolean;
  type: ProxyType;
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
  minVersion: TlsVersion;
  maxVersion?: TlsVersion;
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
    const min = VERSION_ORDER[config.minVersion] ?? 0;
    const max = config.maxVersion ? (VERSION_ORDER[config.maxVersion] ?? 3) : 3;
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
