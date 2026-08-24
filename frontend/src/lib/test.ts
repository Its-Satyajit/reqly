export interface TestAssertion {
  kind: "status" | "header" | "body_contains" | "body_equals" | "json" | "response_time";
  path?: string;
  value?: string;
  expected?: number;
  exact?: boolean;
  not?: boolean;
}

export interface TestAssertionResult {
  assertion: TestAssertion;
  passed: boolean;
  message?: string;
}

export interface TestResultView {
  name: string;
  passed: boolean;
  results: TestAssertionResult[];
}

export interface TestFileRef {
  name: string;
  path: string;
}

export interface TestFileContent {
  path: string;
  content: string;
  format: string;
  version: string;
}

export interface TestRunOutcome {
  passed: boolean;
  passCount: number;
  total: number;
  durationMs: number;
  results: TestResultView[] | null;
  error?: string;
}

export interface TestAdapter {
  list(): Promise<TestFileRef[]>;
  read(path: string): Promise<TestFileContent>;
  write(path: string, content: string): Promise<void>;
  run(input: { path: string; content?: string; env?: string }): Promise<TestRunOutcome>;
}

export const fallbackTestAdapter: TestAdapter = {
  async list() {
    return [];
  },
  async read() {
    throw new Error("test runner is not available in this build");
  },
  async write() {
    throw new Error("test runner is not available in this build");
  },
  async run() {
    throw new Error("test runner is not available in this build");
  },
};

export const TEST_ASSERTION_KIND_OPTIONS = [
  { value: "status", label: "Status code" },
  { value: "header", label: "Header" },
  { value: "body_contains", label: "Body contains" },
  { value: "body_equals", label: "Body equals" },
  { value: "json", label: "JSONPath" },
  { value: "response_time", label: "Response time (ms)" },
];
