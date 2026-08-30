// Reqly - Interactive Endpoint Editing for Spec Editor (§56.1)
// SPDX-License-Identifier: Apache-2.0
import { useState, useMemo } from "react";
import { AlertCircle, Check, Pencil } from "lucide-react";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { validateEndpoint, type EndpointInput } from "#lib/specTree";

const METHODS = ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE"] as const;

interface EndpointEditorProps {
  initial: EndpointInput;
  onSave: (updated: EndpointInput) => void;
  onCancel?: () => void;
}

export function EndpointEditor({ initial, onSave, onCancel }: EndpointEditorProps) {
  const [path, setPath] = useState(initial.path);
  const [method, setMethod] = useState(initial.method.toUpperCase());
  const [summary, setSummary] = useState(initial.summary ?? "");
  const [operationId, setOperationId] = useState(initial.operationId ?? "");
  const [touched, setTouched] = useState(false);

  const draft: EndpointInput = useMemo(
    () => ({
      path,
      method,
      summary: summary.trim() === "" ? undefined : summary.trim(),
      operationId: operationId.trim() === "" ? undefined : operationId.trim(),
    }),
    [path, method, summary, operationId],
  );

  const errors = useMemo(() => validateEndpoint(draft), [draft]);
  const hasErrors = errors.length > 0;

  const handleSave = () => {
    setTouched(true);
    if (hasErrors) return;
    onSave(draft);
  };

  return (
    <div className="space-y-3 rounded border border-border bg-card p-3" aria-label="Endpoint editor">
      <div className="flex items-center gap-1.5 text-xs font-semibold">
        <Pencil className="size-3" />
        Edit Endpoint
        <span className="ml-auto text-[11px] font-normal text-muted-foreground">Validation on save</span>
      </div>

      <div className="grid grid-cols-[110px_1fr] items-center gap-2">
        <label htmlFor="ep-method" className="text-xs font-medium">
          Method
        </label>
        <select
          id="ep-method"
          value={method}
          onChange={(e) => setMethod(e.target.value)}
          className="h-7 rounded border border-border bg-background px-2 font-mono text-xs"
          aria-label="Endpoint method"
        >
          {METHODS.map((m) => (
            <option key={m} value={m}>
              {m}
            </option>
          ))}
        </select>

        <label htmlFor="ep-path" className="text-xs font-medium">
          Path
        </label>
        <Input
          id="ep-path"
          value={path}
          onChange={(e) => setPath(e.target.value)}
          placeholder="/users/{id}"
          className="font-mono text-xs"
          aria-label="Endpoint path"
        />

        <label htmlFor="ep-summary" className="text-xs font-medium">
          Summary
        </label>
        <Input
          id="ep-summary"
          value={summary}
          onChange={(e) => setSummary(e.target.value)}
          placeholder="List users"
          className="text-xs"
          aria-label="Endpoint summary"
        />

        <label htmlFor="ep-opid" className="text-xs font-medium">
          operationId
        </label>
        <Input
          id="ep-opid"
          value={operationId}
          onChange={(e) => setOperationId(e.target.value)}
          placeholder="listUsers"
          className="font-mono text-xs"
          aria-label="Endpoint operationId"
        />
      </div>

      {touched && hasErrors && (
        <div className="rounded border border-destructive/30 bg-destructive/10 p-2" role="alert">
          <p className="flex items-center gap-1 text-xs font-medium text-destructive">
            <AlertCircle className="size-3" />
            Validation
          </p>
          <ul className="mt-1 list-disc pl-4 text-xs text-destructive">
            {errors.map((e, i) => (
              <li key={i}>{e}</li>
            ))}
          </ul>
        </div>
      )}

      {!touched && hasErrors && errors.length > 0 && (
        <p className="text-xs text-muted-foreground">{errors.length} validation issue(s) — will block save.</p>
      )}

      <div className="flex items-center gap-1.5">
        <Button size="sm" className="h-7 px-3 text-xs" onClick={handleSave} aria-label="Save endpoint">
          <Check className="mr-1 size-3" />
          Save
        </Button>
        {onCancel && (
          <Button variant="ghost" size="sm" className="h-7 px-2 text-xs" onClick={onCancel} aria-label="Cancel edit">
            Cancel
          </Button>
        )}
        {!hasErrors && <span className="ml-1 text-xs text-status-ok">Valid</span>}
      </div>
    </div>
  );
}
