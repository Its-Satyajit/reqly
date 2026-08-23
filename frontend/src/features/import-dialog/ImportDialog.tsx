import { useMemo, useRef, useState } from "react";
import { ArrowLeft, FileUp, Upload } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "#components/ui/dialog";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Label } from "#components/ui/field";
import { Spinner } from "#components/ui/spinner";
import { Textarea } from "#components/ui/textarea";
import { CompactSelect } from "#components/CompactSelect";
import { cn } from "#lib/utils";
import {
  formatLabel,
  IMPORT_FORMAT_OPTIONS,
  type ImportedOperation,
} from "#lib/import";
import { useImportStore } from "#stores/useImportStore";
import { ImportReportView } from "./ImportReportView";

const OPERATION_CAP = 50;

function formatBadge(format: string, ok: boolean, hint: string) {
  if (hint !== "") return null;
  if (!ok) {
    return <Badge variant="ghost">Unknown format</Badge>;
  }
  return (
    <Badge variant="secondary" className="text-status-ok">
      {formatLabel(format)}
    </Badge>
  );
}

function OperationGroups({ operations }: { operations: ImportedOperation[] }) {
  const [expanded, setExpanded] = useState(false);
  const groups = useMemo(() => {
    const byTag = new Map<string, ImportedOperation[]>();
    for (const op of operations) {
      const tag = op.tags?.[0] ?? "untagged";
      const list = byTag.get(tag) ?? [];
      list.push(op);
      byTag.set(tag, list);
    }
    return [...byTag.entries()].sort(([a], [b]) => a.localeCompare(b));
  }, [operations]);

  const total = operations.length;
  const shown = expanded ? total : Math.min(total, OPERATION_CAP);

  let rendered = 0;
  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-col gap-2">
        {groups.map(([tag, ops]) => {
          const visible = expanded ? ops : ops.slice(0, Math.max(0, shown - rendered));
          rendered += visible.length;
          if (visible.length === 0) return null;
          return (
            <div key={tag}>
              <p className="pb-0.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                {tag} · {ops.length}
              </p>
              <ul className="flex flex-col divide-y divide-border/60 rounded-md border border-border">
                {visible.map((op, i) => (
                  <li key={i} className="flex items-baseline gap-2 px-2 py-1 font-mono text-[11px]">
                    <span className="w-12 shrink-0 font-sans text-[10px] font-semibold text-status-info">
                      {op.method}
                    </span>
                    <span className="truncate">{op.path}</span>
                    {op.summary && (
                      <span className="truncate font-sans text-[10px] text-muted-foreground">
                        {op.summary}
                      </span>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          );
        })}
      </div>
      {!expanded && total > OPERATION_CAP && (
        <Button variant="ghost" size="sm" className="self-start" onClick={() => setExpanded(true)}>
          Show all {total} operations
        </Button>
      )}
    </div>
  );
}

export function ImportDialog({ onImported }: { onImported?: () => void }) {
  const open = useImportStore((s) => s.open);
  const setOpen = useImportStore((s) => s.setOpen);
  const stage = useImportStore((s) => s.stage);
  const content = useImportStore((s) => s.content);
  const filename = useImportStore((s) => s.filename);
  const formatHint = useImportStore((s) => s.formatHint);
  const detected = useImportStore((s) => s.detected);
  const outcome = useImportStore((s) => s.outcome);
  const targetDir = useImportStore((s) => s.targetDir);
  const busy = useImportStore((s) => s.busy);
  const error = useImportStore((s) => s.error);
  const setContent = useImportStore((s) => s.setContent);
  const setFormatHint = useImportStore((s) => s.setFormatHint);
  const setTargetDir = useImportStore((s) => s.setTargetDir);
  const runPreview = useImportStore((s) => s.runPreview);
  const commit = useImportStore((s) => s.commit);
  const back = useImportStore((s) => s.back);

  const fileInputRef = useRef<HTMLInputElement>(null);
  const [dragging, setDragging] = useState(false);

  const readFile = async (file: File) => {
    setContent(await file.text(), file.name);
  };

  const report = outcome?.report;
  const isWorkspaceKind = outcome != null && outcome.kind === "workspace";

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="flex max-h-[85vh] w-full flex-col gap-3 overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Import</DialogTitle>
          <DialogDescription>
            Drop a collection export or paste a cURL command. Nothing is written
            until you confirm the preview.
          </DialogDescription>
        </DialogHeader>

        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {stage === "input" && (
          <div className="flex flex-col gap-3">
            <div
              onDragOver={(e) => {
                e.preventDefault();
                setDragging(true);
              }}
              onDragLeave={() => setDragging(false)}
              onDrop={(e) => {
                e.preventDefault();
                setDragging(false);
                const file = e.dataTransfer.files[0];
                if (file) void readFile(file);
              }}
              className={cn(
                "flex flex-col items-center gap-1 rounded-md border border-dashed px-4 py-6 text-center transition-colors",
                dragging ? "border-primary bg-primary/5" : "border-border",
              )}
            >
              <Upload className="size-4 text-muted-foreground" aria-hidden />
              <p className="text-xs text-muted-foreground">
                Drag a file here, or{" "}
                <button
                  type="button"
                  className="underline underline-offset-2 hover:text-foreground"
                  onClick={() => fileInputRef.current?.click()}
                >
                  browse
                </button>
              </p>
              <input
                ref={fileInputRef}
                type="file"
                accept=".json,.yaml,.yml,.har,.txt"
                className="sr-only"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) void readFile(file);
                  e.target.value = "";
                }}
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="import-content" className="text-xs font-medium">
                Or paste content
              </Label>
              <Textarea
                id="import-content"
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="curl https://api.example.com/users"
                rows={7}
                spellCheck={false}
                className="resize-y font-mono text-xs"
              />
            </div>

            <div className="flex items-center justify-between gap-2">
              <div className="flex items-center gap-2">
                <span className="text-xs text-muted-foreground">Format</span>
                <CompactSelect
                  value={formatHint}
                  onChange={setFormatHint}
                  options={IMPORT_FORMAT_OPTIONS}
                  ariaLabel="Import format"
                />
              </div>
              {formatBadge(detected?.format ?? "", detected?.ok ?? false, formatHint)}
            </div>

            <DialogFooter>
              <Button onClick={() => void runPreview()} disabled={content.trim() === "" || busy}>
                {busy ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <FileUp data-icon="inline-start" />
                )}
                {busy ? "Reading…" : "Preview import"}
                {filename !== "" && ` · ${filename}`}
              </Button>
            </DialogFooter>
          </div>
        )}

        {stage === "preview" && outcome && (
          <div className="flex min-h-0 flex-col gap-3">
            <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs">
              <span className="font-medium">
                {outcome.title || formatLabel(outcome.format)}
              </span>
              <span className="text-muted-foreground">
                {outcome.requestCount} request{outcome.requestCount === 1 ? "" : "s"}
                {outcome.environmentCount
                  ? `, ${outcome.environmentCount} environment${outcome.environmentCount === 1 ? "" : "s"}`
                  : ""}
              </span>
            </div>

            {isWorkspaceKind && (
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="import-target-dir" className="text-xs font-medium">
                  Folder name in workspace
                </Label>
                <Input
                  id="import-target-dir"
                  value={targetDir}
                  onChange={(e) => setTargetDir(e.target.value)}
                  spellCheck={false}
                  className="font-mono text-xs"
                />
              </div>
            )}

            {outcome.operations && outcome.operations.length > 0 && (
              <OperationGroups operations={outcome.operations} />
            )}

            <div className="min-h-0">
              <p className="pb-1 text-xs font-medium">What changed vs the source</p>
              <ImportReportView report={report} />
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={back} disabled={busy}>
                <ArrowLeft data-icon="inline-start" />
                Back
              </Button>
              {isWorkspaceKind && (
                <Button
                  onClick={() =>
                    void commit().then((res) => {
                      if (res) onImported?.();
                    })
                  }
                  disabled={busy}
                >
                  {busy && <Spinner data-icon="inline-start" />}
                  Import into workspace
                </Button>
              )}
            </DialogFooter>
          </div>
        )}

        {stage === "results" && outcome && (
          <div className="flex min-h-0 flex-col gap-3">
            <div className="rounded-md border border-border px-3 py-2 text-xs">
              {outcome.kind === "request" ? (
                <>Parsed as a cURL command — it will open as an unsaved request tab.</>
              ) : (
                <>
                  Imported{" "}
                  <span className="font-medium">
                    {outcome.requestCount} request{outcome.requestCount === 1 ? "" : "s"}
                  </span>{" "}
                  into <span className="font-mono">{targetDir || outcome.targetDir}</span>.
                </>
              )}
            </div>

            <div className="min-h-0">
              <p className="pb-1 text-xs font-medium">What changed vs the source</p>
              <ImportReportView report={report} />
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
