import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "#components/ui/dialog";
import { Button } from "#components/ui/button";
import { cn } from "#lib/utils";

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

interface CrashReportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  report: string;
  onReload?: () => void;
}

export function CrashReportDialog({
  open,
  onOpenChange,
  report,
  onReload,
}: CrashReportDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[85vh] flex-col gap-4 sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Reqly hit an unexpected error</DialogTitle>
          <DialogDescription>
            The app state may be unreliable. The report below contains no request
            data, headers, URLs, or credentials — copy it to share what happened.
          </DialogDescription>
        </DialogHeader>
        <pre className="min-h-0 flex-1 overflow-auto rounded-md border border-border bg-muted p-3 font-mono text-xs leading-relaxed">
          {report}
        </pre>
        <DialogFooter className="gap-2">
          <Button variant="ghost" size="sm" onClick={() => onOpenChange(false)}>
            Dismiss
          </Button>
          {onReload && (
            <Button variant="outline" size="sm" onClick={onReload}>
              Reload
            </Button>
          )}
          <CopyReportButton report={report} />
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
