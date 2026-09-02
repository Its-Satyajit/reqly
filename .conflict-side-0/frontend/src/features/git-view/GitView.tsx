import { useCallback, useEffect, useState } from "react";
import { GitBranch, GitCommit, FileDiff, ScrollText, RefreshCw, Copy } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Alert, AlertDescription } from "#components/ui/alert";
import { getGitBridge } from "#lib/git";
import { copyText } from "#lib/response";
import { cn } from "#lib/utils";

type Tab = "status" | "diff" | "log";

function StatusBadge({ status }: { status: string[] }) {
  const clean = status.length === 0 || (status.length === 1 && status[0] === "");
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full border px-2 py-0.5 font-mono text-[11px]",
        clean ? "border-status-ok/20 bg-status-ok/10 text-status-ok" : "border-status-warn/20 bg-status-warn/10 text-status-warn",
      )}
    >
      <span className={cn("size-1.5 rounded-full", clean ? "bg-status-ok" : "bg-status-warn")} aria-hidden />
      {clean ? "clean" : `${status.length} changed`}
    </span>
  );
}

function parseStatusLine(line: string) {
  // porcelain: XY<space>path — e.g. " M collections/a.yaml", "A  collections/b.yaml", "?? untracked"
  if (line.startsWith("??") || line.startsWith("!!")) {
    return { staged: line[0] ?? " ", unstaged: line[1] ?? " ", path: line.slice(3).trim(), raw: line };
  }
  return { staged: line[0] ?? " ", unstaged: line[1] ?? " ", path: line.slice(3).trim(), raw: line };
}

export function GitView() {
  const [status, setStatus] = useState<string[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [message, setMessage] = useState("");
  const [diff, setDiff] = useState("");
  const [stagedDiff, setStagedDiff] = useState("");
  const [log, setLog] = useState<string[]>([]);
  const [tab, setTab] = useState<Tab>("status");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [branch, setBranch] = useState<string>("");

  const load = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      const adapter = getGitBridge();
      const [s, d, sd, l] = await Promise.all([adapter.status(), adapter.diff(false), adapter.diff(true), adapter.log(50, 0)]);
      setStatus(s ?? []);
      setDiff(d ?? "");
      setStagedDiff(sd ?? "");
      setLog(l ?? []);
      // Try to infer branch from log's first line or status? For now, placeholder.
      setBranch(s && s.length > 0 ? "worktree" : "main");
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const toggle = (line: string) => {
    const next = new Set(selected);
    if (next.has(line)) next.delete(line);
    else next.add(line);
    setSelected(next);
  };

  const commit = async () => {
    if (!message.trim()) return;
    setBusy(true);
    setError(null);
    try {
      const adapter = getGitBridge();
      const files = Array.from(selected).map((l) => parseStatusLine(l).path);
      await adapter.commit(message.trim(), files);
      setMessage("");
      setSelected(new Set());
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <section className="flex h-full min-h-0 flex-col" aria-label="Git status">
      <PageHeader
        icon={GitBranch}
        title="Git"
        description="Local-only Git status — branch, porcelain status, diff, and log. 0600, no cloud."
        actions={
          <div className="flex items-center gap-1.5">
            <Badge variant="outline" className="gap-1 font-mono text-[11px]">
              <GitBranch className="size-3" /> {branch || "—"}
            </Badge>
            <StatusBadge status={status} />
            <Button variant="ghost" size="xs" onClick={() => void load()} disabled={busy} className="gap-1">
              <RefreshCw className={cn("size-3.5", busy && "animate-spin")} /> Refresh
            </Button>
          </div>
        }
      />

      <div className="flex gap-1 border-b border-border bg-muted/20 px-3 py-1.5">
        {
          // SAFETY: literal array is exactly Tab union
          (["status", "diff", "log"] as Tab[]).map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => setTab(t)}
            className={cn(
              "rounded px-3 py-1 text-xs font-medium capitalize",
              tab === t ? "bg-primary text-primary-foreground" : "border border-border hover:bg-muted",
            )}
            aria-selected={tab === t}
          >
            {t === "status" ? "Status" : t === "diff" ? "Diff" : "Log"}
          </button>
        ))}
        <span className="ml-auto hidden items-center gap-1 font-mono text-[10px] text-muted-foreground sm:inline-flex">
          <FileDiff className="size-3" /> {tab === "status" ? `${status.length} files` : tab === "diff" ? `${diff.length} chars` : `${log.length} commits`}
        </span>
      </div>

      {error && (
        <Alert variant="destructive" className="mx-3 mt-3 py-2">
          <AlertDescription className="font-mono text-xs">{error}</AlertDescription>
        </Alert>
      )}

      <div className="min-h-0 flex-1 overflow-auto p-3">
        {tab === "status" && (
          <div className="flex flex-col gap-3">
            {status.length === 0 ? (
              <div className="flex min-h-[180px] flex-col items-center justify-center gap-2 rounded border border-dashed border-border bg-card/60 p-6 text-center">
                <GitCommit className="size-5 text-muted-foreground/50" aria-hidden />
                <p className="text-sm font-medium">Working tree clean</p>
                <p className="max-w-[48ch] text-balance font-mono text-xs text-muted-foreground">No porcelain status — nothing to commit. Edit collections, environments, or request files and `git status` will appear here.</p>
              </div>
            ) : (
              <div className="overflow-auto rounded border border-border bg-card">
                <table className="w-full border-collapse text-xs">
                  <thead className="sticky top-0 border-b border-border bg-muted/30">
                    <tr className="text-left font-mono text-[11px] text-muted-foreground">
                      <th className="w-6 px-2 py-1.5">
                        <input
                          type="checkbox"
                          checked={selected.size === status.length}
                          onChange={(e) => setSelected(e.target.checked ? new Set(status) : new Set())}
                          aria-label="Select all"
                        />
                      </th>
                      <th className="px-2 py-1.5 font-semibold">XY</th>
                      <th className="px-2 py-1.5 font-semibold">Path</th>
                    </tr>
                  </thead>
                  <tbody>
                    {status.map((line) => {
                      const { staged, unstaged, path } = parseStatusLine(line);
                      const stagedSig = staged.trim() !== "" && staged !== " " && staged !== "?";
                      const unstagedSig = unstaged.trim() !== "" && unstaged !== " " && unstaged !== "?";
                      return (
                        <tr key={line} className="border-b border-border/40 last:border-0 hover:bg-muted/20">
                          <td className="px-2 py-1.5">
                            <input type="checkbox" checked={selected.has(line)} onChange={() => toggle(line)} aria-label={`Select ${path}`} />
                          </td>
                          <td className="px-2 py-1.5 font-mono text-[11px]">
                            <span className={cn("rounded px-1", stagedSig ? "bg-status-redirect/10 text-status-redirect" : "text-muted-foreground")}>{staged || " "}</span>
                            <span className={cn("rounded px-1", unstagedSig ? "bg-status-warn/10 text-status-warn" : "text-muted-foreground")}>{unstaged || " "}</span>
                          </td>
                          <td className="max-w-[42ch] truncate px-2 py-1.5 font-mono text-[11px]">{path}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}

            <div className="flex flex-col gap-2 rounded border border-border bg-card p-3">
              <label className="flex flex-col gap-1">
                <span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Commit message</span>
                <Input value={message} onChange={(e) => setMessage(e.target.value)} placeholder="feat(api): add users endpoint" className="font-mono text-xs" />
              </label>
              <div className="flex flex-wrap items-center gap-2">
                <Button onClick={() => void commit()} disabled={!message.trim() || busy} className="gap-1.5">
                  <GitCommit className="size-3.5" /> Commit {selected.size > 0 ? `(${selected.size} files)` : "(all)"}
                </Button>
                <span className="font-mono text-[11px] text-muted-foreground">{selected.size > 0 ? "staged selection" : "git commit -a will be simulated via explicit add"}</span>
                <Button variant="ghost" size="xs" onClick={() => void load()} className="ml-auto gap-1">
                  <RefreshCw className="size-3.5" /> Reload status
                </Button>
              </div>
            </div>
          </div>
        )}

        {tab === "diff" && (
          <div className="flex flex-col gap-3">
            <div className="grid gap-3 lg:grid-cols-2">
              <div className="rounded border border-border bg-card">
                <div className="flex items-center justify-between border-b border-border px-2 py-1.5">
                  <p className="font-mono text-xs font-semibold">Unstaged diff</p>
                  <Button variant="ghost" size="xs" onClick={() => void copyText(diff)} className="gap-1">
                    <Copy className="size-3" /> Copy
                  </Button>
                </div>
                <pre className="max-h-[420px] overflow-auto whitespace-pre-wrap break-all p-2 font-mono text-[11px] leading-relaxed">{diff || "No unstaged diff — working tree matches index."}</pre>
              </div>
              <div className="rounded border border-border bg-card">
                <div className="flex items-center justify-between border-b border-border px-2 py-1.5">
                  <p className="font-mono text-xs font-semibold">Staged diff</p>
                  <Button variant="ghost" size="xs" onClick={() => void copyText(stagedDiff)} className="gap-1">
                    <Copy className="size-3" /> Copy
                  </Button>
                </div>
                <pre className="max-h-[420px] overflow-auto whitespace-pre-wrap break-all p-2 font-mono text-[11px] leading-relaxed">{stagedDiff || "No staged diff — index matches HEAD."}</pre>
              </div>
            </div>
          </div>
        )}

        {tab === "log" && (
          <div className="flex flex-col gap-2">
            {log.length === 0 ? (
              <div className="flex min-h-[180px] flex-col items-center justify-center gap-2 rounded border border-dashed border-border bg-card/60 p-6 text-center">
                <ScrollText className="size-5 text-muted-foreground/50" aria-hidden />
                <p className="text-sm font-medium">No commits yet</p>
                <p className="font-mono text-xs text-muted-foreground">`git log --oneline` will appear here once you commit.</p>
              </div>
            ) : (
              <div className="overflow-auto rounded border border-border bg-card">
                <table className="w-full border-collapse text-xs">
                  <thead className="sticky top-0 border-b border-border bg-muted/30">
                    <tr className="text-left font-mono text-[11px] text-muted-foreground">
                      <th className="px-2 py-1.5 font-semibold">Graph</th>
                      <th className="px-2 py-1.5 font-semibold">Message</th>
                    </tr>
                  </thead>
                  <tbody>
                    {log.map((line, idx) => {
                      const hash = line.split(" ")[0] ?? "";
                      const msg = line.slice(hash.length).trim();
                      return (
                        <tr key={`${hash}-${idx}`} className="border-b border-border/40 last:border-0 hover:bg-muted/20">
                          <td className="whitespace-nowrap px-2 py-1.5 font-mono text-[11px] text-muted-foreground">
                            <span className="rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">{hash.slice(0, 7)}</span>
                          </td>
                          <td className="px-2 py-1.5 font-mono text-[11px]">{msg}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </div>
    </section>
  );
}
