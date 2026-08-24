import { useState } from "react";
import { CheckCircle2, GitCompareArrows, ShieldCheck } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Spinner } from "#components/ui/spinner";
import { cn } from "#lib/utils";
import {
  getBridge,
  type CrossEnvGap,
  type EnvIssue,
  type EnvKeyDiff,
} from "#lib/envtools";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";

type Panel = "" | "diff" | "validate" | "cross";

function EnvPicker({
  id,
  value,
  names,
  onChange,
}: {
  id: string;
  value: string;
  names: string[];
  onChange: (v: string) => void;
}) {
  return (
    <select
      id={id}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="rounded-md border border-border bg-transparent px-2 py-1 text-xs"
    >
      <option value="">Pick environment…</option>
      {names.map((n) => (
        <option key={n} value={n}>{n}</option>
      ))}
    </select>
  );
}

function DiffTable({ diffs }: { diffs: EnvKeyDiff[] }) {
  if (diffs.length === 0) {
    return <p className="text-xs text-status-ok">Environments are identical.</p>;
  }
  return (
    <table className="w-full table-fixed text-left text-xs">
      <thead>
        <tr className="text-muted-foreground">
          <th className="w-24 py-1">Status</th>
          <th className="py-1">Key</th>
          <th className="w-36 py-1">From</th>
          <th className="w-36 py-1">To</th>
        </tr>
      </thead>
      <tbody>
        {diffs.map((d) => (
          <tr key={`${d.kind}-${d.name}`} className="border-t border-border/60">
            <td className="py-0.5">
              <Badge
                variant="outline"
                className={cn(
                  d.status === "removed" && "border-status-error/40 text-status-error",
                  d.status === "added" && "border-status-ok/40 text-status-ok",
                )}
              >
                {d.status}
              </Badge>
            </td>
            <td className="truncate py-0.5 font-mono" title={d.name}>
              {d.name}
              {d.kind === "secret" && (
                <span className="ml-1 text-[10px] text-muted-foreground">(secret)</span>
              )}
            </td>
            <td className="truncate py-0.5 font-mono text-muted-foreground">{d.from}</td>
            <td className="truncate py-0.5 font-mono">{d.to}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function EnvToolsPanel() {
  const environments = useWorkspaceStore((s) => s.environments);
  const names = environments.map((e) => e.name);
  const [panel, setPanel] = useState<Panel>("diff");
  const [envA, setEnvA] = useState("");
  const [envB, setEnvB] = useState("");
  const [validateTarget, setValidateTarget] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [diffs, setDiffs] = useState<EnvKeyDiff[] | null>(null);
  const [issues, setIssues] = useState<EnvIssue[] | null>(null);
  const [gaps, setGaps] = useState<CrossEnvGap[] | null>(null);

  const runDiff = (): void => {
    if (envA === "" || envB === "") return;
    setBusy(true);
    setError(null);
    getBridge()
      .diff(envA, envB)
      .then((res) => {
        setDiffs(res.diffs);
        setBusy(false);
      })
      .catch((error) => {
        setError(error instanceof Error ? error.message : String(error));
        setBusy(false);
      });
  };

  const runValidate = (): void => {
    if (validateTarget === "") return;
    setBusy(true);
    setError(null);
    getBridge()
      .validate(validateTarget)
      .then((res) => {
        setIssues(res.issues);
        setBusy(false);
      })
      .catch((error) => {
        setError(error instanceof Error ? error.message : String(error));
        setBusy(false);
      });
  };

  const runCrossValidate = (): void => {
    setBusy(true);
    setError(null);
    getBridge()
      .crossValidate()
      .then((res) => {
        setGaps(res);
        setBusy(false);
      })
      .catch((error) => {
        setError(error instanceof Error ? error.message : String(error));
        setBusy(false);
      });
  };

  return (
    <div className="flex flex-col gap-3 border-t border-border pt-3" aria-label="Environment diff and validation">
      <div className="flex items-center gap-2">
        <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
          Diff &amp; Validation
        </span>
        <div className="flex gap-1">
          <Button
            variant={panel === "diff" ? "secondary" : "ghost"}
            size="sm"
            onClick={() => setPanel("diff")}
          >
            <GitCompareArrows data-icon="inline-start" />
            Diff
          </Button>
          <Button
            variant={panel === "validate" ? "secondary" : "ghost"}
            size="sm"
            onClick={() => setPanel("validate")}
          >
            <ShieldCheck data-icon="inline-start" />
            Validate
          </Button>
          <Button
            variant={panel === "cross" ? "secondary" : "ghost"}
            size="sm"
            onClick={() => setPanel("cross")}
          >
            <CheckCircle2 data-icon="inline-start" />
            Cross-env
          </Button>
        </div>
      </div>

      {panel === "diff" && (
        <div className="flex items-end gap-2">
          <EnvPicker id="envtools-a" value={envA} names={names} onChange={setEnvA} />
          <EnvPicker id="envtools-b" value={envB} names={names} onChange={setEnvB} />
          <Button
            size="sm"
            disabled={busy || envA === "" || envB === "" || envA === envB}
            onClick={runDiff}
          >
            {busy ? <Spinner data-icon="inline-start" /> : <GitCompareArrows data-icon="inline-start" />}
            Diff
          </Button>
        </div>
      )}

      {panel === "validate" && (
        <div className="flex items-center gap-2">
          <EnvPicker id="envtools-validate" value={validateTarget} names={names} onChange={setValidateTarget} />
          <Button size="sm" disabled={busy || validateTarget === ""} onClick={runValidate}>
            {busy ? <Spinner data-icon="inline-start" /> : <ShieldCheck data-icon="inline-start" />}
            Validate
          </Button>
        </div>
      )}

      {panel === "cross" && (
        <div>
          <Button size="sm" disabled={busy} onClick={runCrossValidate}>
            {busy ? <Spinner data-icon="inline-start" /> : <CheckCircle2 data-icon="inline-start" />}
            Check all environments
          </Button>
        </div>
      )}

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {diffs != null && panel === "diff" && <DiffTable diffs={diffs} />}

      {issues != null && panel === "validate" && (
        <ul className="flex flex-col gap-1 text-xs">
          {issues.length === 0 ? (
            <li className="text-status-ok">No issues found.</li>
          ) : (
            issues.map((issue, i) => (
              <li key={`${issue.severity}-${issue.message}-${i}`} className="flex items-baseline gap-1.5">
                <Badge
                  variant="outline"
                  className={cn(
                    issue.severity === "error" && "border-status-error/40 text-status-error",
                  )}
                >
                  {issue.severity}
                </Badge>
                {issue.message}
              </li>
            ))
          )}
        </ul>
      )}

      {gaps != null && panel === "cross" && (
        <ul className="flex flex-col gap-1 text-xs">
          {gaps.length === 0 ? (
            <li className="text-status-ok">All variable keys are consistent across environments.</li>
          ) : (
            gaps.map((gap) => (
              <li key={gap.key} className="flex flex-wrap items-baseline gap-1.5">
                <span className="font-mono">{gap.key}</span>
                <span className="text-muted-foreground">
                  missing in {gap.missingIn.join(", ")}
                </span>
                <span className="text-[11px] text-muted-foreground">
                  (present in {gap.presentIn.join(", ")})
                </span>
              </li>
            ))
          )}
        </ul>
      )}
    </div>
  );
}
