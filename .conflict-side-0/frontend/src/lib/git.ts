export interface GitAdapter {
  status(): Promise<string[]>;
  diff(staged?: boolean): Promise<string>;
  log(limit?: number, offset?: number): Promise<string[]>;
  commit(message: string, files: string[]): Promise<void>;
}

let gitBridge: GitAdapter | null = null;

export function setGitBridge(adapter: GitAdapter): void {
  gitBridge = adapter;
}

export function getGitBridge(): GitAdapter {
  if (!gitBridge) {
    return {
      async status() {
        return [];
      },
      async diff() {
        return "";
      },
      async log() {
        return [];
      },
      async commit() {},
    };
  }
  return gitBridge;
}
