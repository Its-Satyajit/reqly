export interface EnvKeyDiff {
  name: string;
  status: string;
  kind: string;
  from: string;
  to: string;
}

export interface EnvIssue {
  severity: string;
  message: string;
}

export interface CrossEnvGap {
  key: string;
  presentIn: string[];
  missingIn: string[];
}

export interface EnvToolsAdapter {
  diff(envA: string, envB: string): Promise<{ envA: string; envB: string; diffs: EnvKeyDiff[] }>;
  validate(name: string): Promise<{ env: string; issues: EnvIssue[] }>;
  crossValidate(): Promise<CrossEnvGap[]>;
}

// Bridge registry, same pattern as the diff and jwt adapters.
let bridge: EnvToolsAdapter | null = null;

/** setEnvToolsBridge installs the host adapter (called once from the bridge). */
export function setEnvToolsBridge(adapter: EnvToolsAdapter): void {
  bridge = adapter;
}

export function getBridge(): EnvToolsAdapter {
  if (!bridge) throw new Error("env tools are not available in this build");
  return bridge;
}
