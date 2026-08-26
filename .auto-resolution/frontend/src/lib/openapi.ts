export interface OpenapiEndpointView {
  method: string;
  path: string;
  operationId?: string;
  tags?: string[];
  summary?: string;
  requestSchema?: string;
  responseSchemas?: Record<string, string>;
}

export interface OpenapiExploreResultView {
  title: string;
  version?: string | undefined;
  endpoints: OpenapiEndpointView[];
}

export interface OpenapiGenerateResultView {
  targetDir: string;
  created: string[];
  warnings?: string[];
}

export interface OpenapiAdapter {
  explore(specPath: string): Promise<OpenapiExploreResultView>;
  generate(input: {
    specPath: string;
    selections: { method: string; path: string }[];
    dirName: string;
  }): Promise<OpenapiGenerateResultView>;
}

// Bridge registry, same pattern as the other feature adapters.
let bridge: OpenapiAdapter | null = null;

/** setOpenapiBridge installs the host adapter (called once from the bridge). */
export function setOpenapiBridge(adapter: OpenapiAdapter): void {
  bridge = adapter;
}

export function getOpenapiBridge(): OpenapiAdapter {
  if (!bridge) throw new Error("openapi explorer is not available in this build");
  return bridge;
}
