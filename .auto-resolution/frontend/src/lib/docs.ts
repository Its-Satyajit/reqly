export interface DocsFileView {
  name: string;
  content: string;
}

export interface DocsResultView {
  path: string;
  requestCount: number;
  files: DocsFileView[];
}

export interface DocsAdapter {
  generate(input: {
    collections?: string[];
    outName?: string;
  }): Promise<DocsResultView>;
}

export const fallbackDocsAdapter: DocsAdapter = {
  async generate() {
    throw new Error("docs generator is not available in this build");
  },
};
