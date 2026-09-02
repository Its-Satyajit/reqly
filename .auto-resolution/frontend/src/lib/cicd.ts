export interface CicdPipeline {
  name: string;
  environment: string;
  secrets: string[];
  collectionPath?: string;
  environmentPath?: string;
  exitCodeMap?: { [code: number]: string };
  artifacts?: string[];
}

export interface CicdRunResult {
  passed: number;
  failed: number;
  skipped: number;
  duration: number;
  exitCode: number;
  reportPath?: string;
}

export function generateCliCommand(pipeline: CicdPipeline): string {
  const parts = ["reqly", "collection", "test"];
  if (pipeline.collectionPath) {
    parts.push(pipeline.collectionPath);
  }
  parts.push("--env", pipeline.environment);
  parts.push("--fail-fast");
  return parts.join(" ");
}

export function generateGitHubAction(pipeline: CicdPipeline): string {
  const lines: string[] = [
    `name: ${pipeline.name}`,
    "",
    "on:",
    "  push:",
    "    branches: [main]",
    "  pull_request:",
    "    branches: [main]",
    "",
    "jobs:",
    "  test:",
    "    runs-on: ubuntu-latest",
    "    steps:",
    "      - uses: actions/checkout@v4",
    "      - uses: actions/setup-node@v4",
    "        with:",
    "          node-version: 20",
    "      - run: npm install -g @reqly/cli",
  ];

  if (pipeline.secrets.length > 0) {
    const envLines = pipeline.secrets.map((s) => {
      const secret = "${{ secrets." + s + " }}";
      return "          " + s + ": " + secret;
    });
    lines.push(
      `      - run: ${generateCliCommand(pipeline)}`,
      "        env:",
      ...envLines,
    );
  } else {
    lines.push(`      - run: ${generateCliCommand(pipeline)}`);
  }

  return lines.join("\n");
}

export function parseTestReport(report: string): Omit<CicdRunResult, "duration" | "exitCode"> {
  const passed = Number.parseInt(report.match(/(\d+)\s*passed/)?.[1] ?? "0", 10);
  const failed = Number.parseInt(report.match(/(\d+)\s*failed/)?.[1] ?? "0", 10);
  const skipped = Number.parseInt(report.match(/(\d+)\s*skipped/)?.[1] ?? "0", 10);
  return { passed, failed, skipped };
}
