import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "#components/ui/dialog";
import { Button } from "#components/ui/button";
import { CopyReportButton } from "#components/ErrorBoundary";

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
      <DialogContent
        overlayClassName="z-(--z-crash)"
        className="z-(--z-crash) flex max-h-[85vh] flex-col gap-4 sm:max-w-2xl"
      >
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
