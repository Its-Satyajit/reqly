import {
  addBreadcrumb,
  formatReport,
  type CrashEntry,
} from "#lib/crash";

type CrashListener = (report: string) => void;

const listeners = new Set<CrashListener>();
let installed = false;

function notify(entry: CrashEntry): void {
  addBreadcrumb("crash", entry.kind);
  const report = formatReport(entry);
  for (const listener of listeners) listener(report);
}

export function subscribeToCrashes(listener: CrashListener): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function reportFatal(entry: CrashEntry): void {
  notify(entry);
}

export function installCrashReporter(): void {
  if (installed) return;
  installed = true;
  window.addEventListener("error", (event) => {
    notify({
      error: event.error ?? event.message,
      kind: "window-error",
    });
  });
  window.addEventListener("unhandledrejection", (event) => {
    notify({ error: event.reason, kind: "unhandled-rejection" });
  });
}

export function armDebugCrashTrigger(): () => void {
  const onKey = (event: KeyboardEvent) => {
    if (event.ctrlKey && event.altKey && event.shiftKey && event.key.toLowerCase() === "k") {
      event.preventDefault();
      reportFatal({
        error: new Error("Debug crash trigger — no real failure occurred"),
        kind: "window-error",
      });
    }
  };
  window.addEventListener("keydown", onKey);
  return () => window.removeEventListener("keydown", onKey);
}
