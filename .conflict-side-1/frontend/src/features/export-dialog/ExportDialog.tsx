import { Download } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "#components/ui/dialog";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { CompactSelect } from "#components/CompactSelect";
import {
  EXPORT_FORMAT_OPTIONS,
  exportFormatLabel,
  type ExportFormat,
} from "#lib/export";
import { useExportStore } from "#stores/useExportStore";

export function ExportDialog() {
  const open = useExportStore((s) => s.open);
  const setOpen = useExportStore((s) => s.setOpen);
  const format = useExportStore((s) => s.format);
  const collection = useExportStore((s) => s.collection);
  const outName = useExportStore((s) => s.outName);
  const outcome = useExportStore((s) => s.outcome);
  const busy = useExportStore((s) => s.busy);
  const error = useExportStore((s) => s.error);
  const setFormat = useExportStore((s) => s.setFormat);
  const setOutName = useExportStore((s) => s.setOutName);
  const run = useExportStore((s) => s.run);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="flex max-h-[85vh] w-full flex-col gap-3 overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Export</DialogTitle>
          <DialogDescription>
            Write a shareable copy of this workspace to{" "}
            <span className="font-mono">.reqly/exports</span>. Files stay local.
          </DialogDescription>
        </DialogHeader>

        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {outcome == null && (
          <div className="flex flex-col gap-3">
            <div className="flex items-center gap-2">
              <span className="text-xs font-medium">Format</span>
              <CompactSelect
                value={format}
                onChange={(v) => {
                  // SAFETY: options come from EXPORT_FORMAT_OPTIONS, all valid ExportFormats.
                  setFormat(v as ExportFormat);
                }}
                options={EXPORT_FORMAT_OPTIONS}
                ariaLabel="Export format"
              />
            </div>

            {format === "openapi" && (
              <div className="flex flex-col gap-1.5">
                <label htmlFor="export-collection" className="text-xs font-medium">
                  Collection (empty exports the whole workspace)
                </label>
                <Input
                  id="export-collection"
                  value={collection}
                  onChange={(e) => useExportStore.getState().setCollection(e.target.value)}
                  placeholder="users"
                  spellCheck={false}
                  className="font-mono text-xs"
                />
              </div>
            )}

            <div className="flex flex-col gap-1.5">
              <label htmlFor="export-out-name" className="text-xs font-medium">
                File name (optional)
              </label>
              <Input
                id="export-out-name"
                value={outName}
                onChange={(e) => setOutName(e.target.value)}
                spellCheck={false}
                className="font-mono text-xs"
              />
            </div>

            <DialogFooter>
              <Button onClick={() => void run()} disabled={busy}>
                {busy ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <Download data-icon="inline-start" />
                )}
                {busy ? "Exporting…" : "Export"}
              </Button>
            </DialogFooter>
          </div>
        )}

        {outcome != null && (
          <div className="flex min-h-0 flex-col gap-3">
            <div className="rounded-md border border-border px-3 py-2 text-xs">
              Exported{" "}
              <span className="font-medium">{exportFormatLabel(outcome.format)}</span>
              {" · "}
              {outcome.requestCount != null && outcome.requestCount > 0 && (
                <>
                  {outcome.requestCount} request
                  {outcome.requestCount === 1 ? "" : "s"} ·{" "}
                </>
              )}
              {outcome.entryCount != null && <>last {outcome.entryCount} history entries · </>}
              written to <span className="font-mono break-all">{outcome.path}</span>.
            </div>
            <DialogFooter>
              <Button onClick={() => setOpen(false)}>Done</Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
