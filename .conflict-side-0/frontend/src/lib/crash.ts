export const APP_VERSION = "0.1.0";

export type CrashKind =
  | "render"
  | "window-error"
  | "unhandled-rejection";

export interface Breadcrumb {
  at: number;
  kind: string;
  detail?: string;
}

export interface CrashEntry {
  error: unknown;
  kind: CrashKind;
  componentStack?: string;
}

const MAX_BREADCRUMBS = 20;
let breadcrumbs: Breadcrumb[] = [];

export interface GoLogLine {
  at: number;
  level: string;
  message: string;
}

const MAX_GO_LOGS = 30;
let goLogs: GoLogLine[] = [];

export function addGoLog(level: string, message: string): void {
  goLogs.push({ at: Date.now(), level, message });
  if (goLogs.length > MAX_GO_LOGS) {
    goLogs = goLogs.slice(-MAX_GO_LOGS);
  }
}

export function getGoLogs(): GoLogLine[] {
  return [...goLogs];
}

export function addBreadcrumb(kind: string, detail?: string): void {
  breadcrumbs.push({ at: Date.now(), kind, detail });
  if (breadcrumbs.length > MAX_BREADCRUMBS) {
    breadcrumbs = breadcrumbs.slice(-MAX_BREADCRUMBS);
  }
}

export function getBreadcrumbs(): Breadcrumb[] {
  return [...breadcrumbs];
}

export function platformLabel(): string {
  const ua = navigator.userAgent;
  const os = (() => {
    if (ua.includes("Win")) return "Windows";
    if (ua.includes("Mac")) return "macOS";
    if (ua.includes("Linux")) return "Linux";
    return "Unknown OS";
  })();
  const engineMatch = ua.match(/Chrome\/(\S+)/);
  const engine = engineMatch ? ` WebView/Chrome ${engineMatch[1]}` : "";
  return `${os}${engine} (${navigator.platform || "unknown platform"})`;
}

interface ParsedError {
	name: string;
	message: string;
	stack?: string;
}

/** describeThrown renders any thrown value defensively — window handlers
 * deliver arbitrary host objects, so nothing here trusts the input's shape. */
function describeThrown(thrown: CrashEntry["error"]): ParsedError {
	if (thrown instanceof Error) {
		return { name: thrown.name, message: thrown.message, stack: thrown.stack };
	}
	try {
		return { name: "Error", message: JSON.stringify(thrown) };
	} catch {
		// SAFETY: last-resort rendering for unserializable thrown values
		return { name: "Error", message: Object.prototype.toString.call(thrown) };
	}
}

const MAX_STACK_LENGTH = 4000;

function truncate(value: string, max: number): string {
  return value.length > max ? `${value.slice(0, max)}\n… (truncated)` : value;
}

/**
 * Builds the shareable report from explicitly allowlisted fields only.
 * Request/response payloads, headers, URLs, environment values, and
 * credentials can never enter this text because they are never passed in.
 */
export function formatReport(entry: CrashEntry): string {
  const { name, message, stack } = describeThrown(entry.error);
  const lines: string[] = [
    "Reqly Crash Report",
    "==================",
    `Version:  ${APP_VERSION}`,
    `Platform: ${platformLabel()}`,
    `Time:     ${new Date().toISOString()}`,
    `Kind:     ${entry.kind}`,
    "",
    `[Error] ${name}`,
    message,
  ];
  if (stack) {
    lines.push("", "[Stack]", truncate(stack, MAX_STACK_LENGTH));
  }
  if (entry.componentStack) {
    lines.push(
      "",
      "[Component stack]",
      truncate(entry.componentStack.trim(), MAX_STACK_LENGTH),
    );
  }
  const crumbs = getBreadcrumbs();
  if (crumbs.length > 0) {
    lines.push("", `[Recent actions] (last ${crumbs.length})`);
    for (const crumb of crumbs) {
      const time = new Date(crumb.at).toLocaleTimeString();
      lines.push(`  ${time}  ${crumb.kind}${crumb.detail ? ` — ${crumb.detail}` : ""}`);
    }
  }
  const logs = getGoLogs();
  if (logs.length > 0) {
    lines.push("", `[Backend log] (last ${logs.length})`);
    for (const entry of logs) {
      const time = new Date(entry.at).toLocaleTimeString();
      lines.push(`  ${time}  ${entry.level}  ${entry.message}`);
    }
  }
  return lines.join("\n");
}
