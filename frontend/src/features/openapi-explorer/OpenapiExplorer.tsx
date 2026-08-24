import { useState } from "react";
import { FolderOutput, RefreshCw } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { cn } from "#lib/utils";
import {
  getOpenapiBridge,
  type OpenapiEndpointView,
} from "#lib/openapi";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";

const METHOD_COLORS = {
  GET: "text-status-ok",
  POST: "text-status-info",
  PUT: "text-warning",
  PATCH: "text-warning",
  DELETE: "text-status-error",
} satisfies Record<string, string>;

export function OpenapiExplorer() {
  const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);
  const [specPath, setSpecPath] = useState("");
  const [result, setResult] = useState<Awaited<
    ReturnType<import("#lib/openapi").OpenapiAdapter["explore"]>
  > | null>(null);
  const [selected, setSelected] = useState<string[]>([]);
  const [dirName, setDirName] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [generated, setGenerated] = useState<string[] | null>(null);

  // Grouped per render; the React Compiler memoizes automatically.
  const grouped: [string, OpenapiEndpointView[]][] = [];
  {
    const byTag = new Map<string, OpenapiEndpointView[]>();
    for (const ep of result?.endpoints ?? []) {
      const tag = ep.tags?.[0] ?? "untagged";
      const list = byTag.get(tag) ?? [];
      list.push(ep);
      byTag.set(tag, list);
    }
    grouped.push(...[...byTag.entries()].sort(([a], [b]) => a.localeCompare(b)));
  }

  const explore = (): void => {
    if (specPath.trim() === "") return;
    setBusy(true);
    setError(null);
    setResult(null);
    setSelected([]);
    setGenerated(null);
    getOpenapiBridge()
      .explore(specPath.trim())
      .then((res) => {
        setResult(res);
        setBusy(false);
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : String(e));
        setBusy(false);
      });
  };

  const toggle = (method: string, path: string): void => {
    const key = `${method}|${path}`;
    setSelected((prev) =>
      prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key],
    );
  };

  const generate = (): void => {
    if (specPath.trim() === "" || selected.length === 0 || dirName.trim() === "") return;
    setBusy(true);
    setError(null);
    const selections = selected.map((k) => {
      // SAFETY: keys are built above as `METHOD|path` pairs.
      const idx = k.indexOf("|");
      return { method: k.slice(0, idx), path: k.slice(idx + 1) };
    });
    getOpenapiBridge()
      .generate({ specPath: specPath.trim(), selections, dirName: dirName.trim() })
      .then((res) => {
        setGenerated(res.created);
        setBusy(false);
        void refreshWorkspace();
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : String(e));
        setBusy(false);
      });
  };

  return (
    <section className="flex h-full min-h-0 flex-col gap-3 overflow-y-auto p-4" aria-label="OpenAPI explorer">
      <h2 className="text-sm font-semibold">OpenAPI Explorer</h2>

      <div className="flex items-end gap-2">
        <div className="flex min-w-64 flex-1 flex-col gap-1">
          <label htmlFor="openapi-spec" className="text-xs font-medium">
            Spec (workspace-relative JSON/YAML)
          </label>
          <Input
            id="openapi-spec"
            value={specPath}
            onChange={(e) => setSpecPath(e.target.value)}
            placeholder="specs/pets.yaml"
            spellCheck={false}
            className="font-mono text-xs"
          />
        </div>
        <Button size="sm" disabled={busy || specPath.trim() === ""} onClick={explore}>
          {busy ? <Spinner data-icon="inline-start" /> : <RefreshCw data-icon="inline-start" />}
          Explore
        </Button>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {result && (
        <>
          <p className="text-xs text-muted-foreground">
            <span className="font-medium text-foreground">{result.title}</span>
            {result.version != null && result.version !== "" && ` · v${result.version}`}
            {" · "}
            {result.endpoints.length} operations
          </p>

          {grouped.map(([tag, eps]) => (
            <div key={tag} className="rounded-md border border-border p-2">
              <p className="pb-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                {tag} · {eps.length}
              </p>
              <ul className="flex flex-col divide-y divide-border/60">
                {eps.map((ep) => {
                  const key = `${ep.method}|${ep.path}`;
                  const checked = selected.includes(key);
                  const schemaKeys = Object.keys(ep.responseSchemas ?? {});
                  return (
                    <li key={key} className="flex flex-col gap-1 py-1">
                      <div className="flex items-center gap-2">
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={() => toggle(ep.method, ep.path)}
                          aria-label={`Select ${ep.method} ${ep.path}`}
                          className="size-3.5 shrink-0 accent-(--primary)"
                        />
                        <span
                          className={cn(
                            "w-14 shrink-0 font-sans text-[10px] font-semibold",
                            METHOD_COLORS[ep.method as keyof typeof METHOD_COLORS] ?? "",
                          )}
                        >
                          {ep.method}
                        </span>
                        <button
                          type="button"
                          className="min-w-0 flex-1 truncate text-left font-mono text-xs hover:text-foreground"
                          title={`${ep.operationId ?? ""} ${ep.summary ?? ""}`}
                          onClick={() =>
                            setSelected(
                              checked ? selected.filter((k) => k !== key) : [...selected, key],
                            )
                          }
                        >
                          {ep.path}
                        </button>
                        {ep.summary && (
                          <span className="hidden truncate text-[10px] text-muted-foreground sm:block">
                            {ep.summary}
                          </span>
                        )}
                      </div>
                      {(ep.requestSchema !== "" ||
                        schemaKeys.length > 0) && (
                        <details className="pl-8">
                          <summary className="cursor-pointer text-[11px] text-muted-foreground">
                            Schemas
                          </summary>
                          {ep.requestSchema !== "" && (
                            <pre className="mt-1 max-h-40 overflow-auto whitespace-pre-wrap break-all rounded bg-muted/50 p-1 font-mono text-[10px]">
                              request: {ep.requestSchema}
                            </pre>
                          )}
                          {schemaKeys.map((status) => (
                            <pre
                              key={status}
                              className="mt-1 max-h-40 overflow-auto whitespace-pre-wrap break-all rounded bg-muted/50 p-1 font-mono text-[10px]"
                            >
                              {status}: {ep.responseSchemas?.[status]}
                            </pre>
                          ))}
                        </details>
                      )}
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}

          <div className="flex items-end gap-2 border-t border-border pt-3">
            <div className="flex flex-col gap-1">
              <label htmlFor="openapi-dir" className="text-xs font-medium">
                Generate into collections/
              </label>
              <Input
                id="openapi-dir"
                value={dirName}
                onChange={(e) => setDirName(e.target.value)}
                placeholder="from-pets-api"
                spellCheck={false}
                className="w-56 font-mono text-xs"
              />
            </div>
            <Badge variant="secondary">{selected.length} selected</Badge>
            <Button
              size="sm"
              disabled={busy || selected.length === 0 || dirName.trim() === ""}
              onClick={generate}
            >
              {busy ? <Spinner data-icon="inline-start" /> : <FolderOutput data-icon="inline-start" />}
              Generate requests
            </Button>
          </div>

          {generated != null && (
            <Alert>
              <AlertDescription>
                Created {generated.length} request file
                {generated.length === 1 ? "" : "s"} under{" "}
                <span className="font-mono">collections/{dirName}</span>. The workspace tree has been
                refreshed.
              </AlertDescription>
            </Alert>
          )}
        </>
      )}
    </section>
  );
}
