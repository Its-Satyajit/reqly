import { useEffect, useState } from "react";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { getGitBridge } from "#lib/git";

export function GitView() {
  const [status, setStatus] = useState<string[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [message, setMessage] = useState("");
  const [diff, setDiff] = useState("");
  const [log, setLog] = useState<string[]>([]);
  const [tab, setTab] = useState<"status" | "diff" | "log">("status");

  const load = async () => {
    try {
      const adapter = getGitBridge();
      const s = await adapter.status();
      if (s) setStatus(s);
      const d = await adapter.diff(false);
      if (d) setDiff(d);
      const l = await adapter.log(50, 0);
      if (l) setLog(l);
    } catch {}
  };
  useEffect(() => { void load(); }, []);

  const toggle = (line: string) => {
    const next = new Set(selected);
    if (next.has(line)) next.delete(line);
    else next.add(line);
    setSelected(next);
  };
  const commit = async () => {
    const adapter = getGitBridge();
    const files = Array.from(selected).map((l) => l.slice(3).trim());
    await adapter.commit(message, files);
    setMessage("");
    setSelected(new Set());
    void load();
  };

  return (
    <div className="flex h-full flex-col gap-3 p-3">
      <div className="flex gap-2">
        <Button variant={tab === "status" ? "default" : "outline"} size="sm" onClick={() => setTab("status")}>Status</Button>
        <Button variant={tab === "diff" ? "default" : "outline"} size="sm" onClick={() => setTab("diff")}>Diff</Button>
        <Button variant={tab === "log" ? "default" : "outline"} size="sm" onClick={() => setTab("log")}>Log</Button>
      </div>
      {tab === "status" && (
        <div className="flex flex-col gap-2">
          {status.length === 0 ? <p className="text-xs text-muted-foreground">Clean</p> : status.map((line) => (
            <label key={line} className="flex items-center gap-2 text-xs">
              <input type="checkbox" checked={selected.has(line)} onChange={() => toggle(line)} /> <code>{line}</code>
            </label>
          ))}
          <div className="flex gap-2">
            <Input value={message} onChange={(e) => setMessage(e.target.value)} placeholder="commit message" className="flex-1" />
            <Button onClick={commit} disabled={!message.trim()}>Commit</Button>
          </div>
        </div>
      )}
      {tab === "diff" && <pre className="overflow-auto rounded bg-muted p-2 text-xs">{diff || "No diff"}</pre>}
      {tab === "log" && <pre className="overflow-auto rounded bg-muted p-2 text-xs">{log.join("\n") || "No log"}</pre>}
    </div>
  );
}
