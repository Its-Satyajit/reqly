export interface WorkspaceStatus {
  found: boolean;
  path?: string;
}

export interface WorkspaceBootstrapAdapter {
  status(): Promise<WorkspaceStatus>;
  restoreLast(): Promise<WorkspaceStatus>;
  pickFolder(): Promise<string>;
  open(dir: string): Promise<void>;
  create(dir: string, name?: string): Promise<void>;
}

export const fallbackWorkspaceBootstrapAdapter: WorkspaceBootstrapAdapter = {
  async status() {
    return { found: false };
  },
  async restoreLast() {
    return { found: false };
  },
  async pickFolder() {
    throw new Error("workspace picker is not available in this build");
  },
  async open() {
    throw new Error("workspace switching is not available in this build");
  },
  async create() {
    throw new Error("workspace creation is not available in this build");
  },
};

export const CREATE_HINT = "is not a Reqly workspace";
