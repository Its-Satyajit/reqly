import { Component, type ErrorInfo, type ReactNode } from "react";
import { addBreadcrumb, formatReport, type CrashEntry } from "#lib/crash";
import { cn } from "#lib/utils";
import { Button } from "#components/ui/button";
import { CrashReportDialog } from "#components/CrashReportDialog";

interface ErrorBoundaryProps {
  children: ReactNode;
  variant?: "panel" | "root";
  label?: string;
}

interface ErrorBoundaryState {
  entry: CrashEntry | null;
  report: string | null;
}

async function copyWithFallback(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    const area = document.createElement("textarea");
    area.value = text;
    area.style.position = "fixed";
    area.style.opacity = "0";
    document.body.appendChild(area);
    area.select();
    let ok = false;
    try {
      ok = document.execCommand("copy");
    } catch {
      ok = false;
    }
    area.remove();
    return ok;
  }
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { entry: null, report: null };

  static getDerivedStateFromError(thrown: CrashEntry["error"]): Partial<ErrorBoundaryState> {
    return { entry: { error: thrown, kind: "render" }, report: null };
  }

  componentDidCatch(thrown: CrashEntry["error"], info: ErrorInfo): void {
    addBreadcrumb("panel-crash", this.props.label);
    this.setState({
      entry: { error: thrown, kind: "render", componentStack: info.componentStack ?? undefined },
      report: formatReport({
        error: thrown,
        kind: "render",
        componentStack: info.componentStack ?? undefined,
      }),
    });
  }

  private reset = (): void => {
    this.setState({ entry: null, report: null });
  };

  render(): ReactNode {
    const { entry, report } = this.state;
    if (!entry) return this.props.children;

    if (this.props.variant === "root") {
      return (
        <CrashReportDialog
          open
          onOpenChange={(open) => {
            if (!open) this.reset();
          }}
          report={report ?? formatReport(entry)}
          onReload={() => window.location.reload()}
        />
      );
    }

    return (
      <div
        role="alert"
        className="flex h-full min-h-32 flex-col items-center justify-center gap-3 p-6 text-center"
      >
        <p className="text-sm font-medium">This panel hit a problem</p>
        <p className="max-w-md truncate font-mono text-xs text-muted-foreground">
          {entry.error instanceof Error ? entry.error.message : String(entry.error)}
        </p>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={this.reset}>
            Retry
          </Button>
          {report && <CopyReportButton report={report} />}
        </div>
      </div>
    );
  }
}

export function CopyReportButton({
  report,
  className,
}: {
  report: string;
  className?: string;
}) {
  return (
    <Button
      size="sm"
      variant="outline"
      className={cn(className)}
      onClick={(event) => {
        const button = event.currentTarget;
        void copyWithFallback(report).then((ok) => {
          if (!button || !button.isConnected) return;
          const previous = button.textContent;
          button.textContent = ok ? "Copied" : "Select manually";
          window.setTimeout(() => {
            button.textContent = previous;
          }, 1500);
          if (!ok) {
            const area = document.createElement("textarea");
            area.value = report;
            document.body.appendChild(area);
            area.select();
            window.setTimeout(() => area.remove(), 30000);
          }
        });
      }}
    >
      Copy report
    </Button>
  );
}
