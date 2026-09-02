import { useState } from "react";
import { Compass, FolderOutput, RefreshCw, Send } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { cn } from "#lib/utils";
import {
  getOpenapiBridge,
  type OpenapiAdapter,
  type OpenapiEndpointView,
} from "#lib/openapi";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { useRequestStore } from "#stores/useRequestStore";

const METHOD_COLOR_MAP = new Map<string, string>([
  ["GET", "text-status-ok"],
  ["POST", "text-status-info"],
  ["PUT", "text-warning"],
  ["PATCH", "text-warning"],
  ["DELETE", "text-status-error"],
]);


export function OpenapiExplorer() {
  const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);
  const openTab = useWorkspaceStore((s) => s.openTab);
  const requestView = useWorkspaceStore((s) => s.requestView);
  const ensureDraft = useRequestStore((s) => s.ensureDraft);

  const [searchFilter, setSearchFilter] = useState("");

  // One reducer-style state object keeps the explorer's related UI fields
  // together (react-doctor/prefer-useReducer).
  type ExplorerState = {
    specPath: string;
    result: Awaited<ReturnType<OpenapiAdapter["explore"]>> | null;
    selected: string[];
    dirName: string;
    busy: boolean;
    error: string | null;
    generated: string[] | null;
  };
  const [ui, setUi] = useState<ExplorerState>({
    specPath: "",
    result: null,
    selected: [],
    dirName: "",
    busy: false,
    error: null,
    generated: null,
  });
  const patch = (p: Partial<ExplorerState>) => setUi((prev) => ({ ...prev, ...p }));
  const { specPath, result, selected, dirName, busy, error, generated } = ui;

  // Grouped per render; the React Compiler memoizes automatically.
  const selectedSet = new Set(selected);
  const grouped: [string, OpenapiEndpointView[]][] = [];
  {
    const byTag = new Map<string, OpenapiEndpointView[]>();
    const filteredEndpoints = (result?.endpoints ?? []).filter((ep) =>
      searchFilter.trim() === ""
        ? true
        : ep.path.toLowerCase().includes(searchFilter.toLowerCase()) ||
          ep.method.toLowerCase().includes(searchFilter.toLowerCase()) ||
          (ep.summary?.toLowerCase().includes(searchFilter.toLowerCase()) ?? false) ||
          (ep.operationId?.toLowerCase().includes(searchFilter.toLowerCase()) ?? false),
    );
    for (const ep of filteredEndpoints) {
      const tag = ep.tags?.[0] ?? "untagged";
      const list = byTag.get(tag) ?? [];
      list.push(ep);
      byTag.set(tag, list);
    }
    grouped.push(...[...byTag.entries()].sort(([a], [b]) => a.localeCompare(b)));
  }

  const explore = (): void => {
    if (specPath.trim() === "") return;
    patch({ busy: true, error: null, result: null, selected: [], generated: null });
    getOpenapiBridge()
      .explore(specPath.trim())
      .then((res) => {
        patch({ result: res, busy: false });
      })
      .catch((e) => {
        patch({ error: e instanceof Error ? e.message : String(e), busy: false });
      });
  };

  const validate = (): void => {
    if (specPath.trim() === "") return;
    patch({ busy: true, error: null });
    getOpenapiBridge()
      .validate(specPath.trim())
      .then(() => {
        patch({ busy: false, error: null });
        // Show success as transient — reuse error slot for success with badge
      })
      .catch((e) => {
        patch({ error: e instanceof Error ? e.message : String(e), busy: false });
      });
  };

  const convertV2 = (): void => {
    if (specPath.trim() === "") return;
    patch({ busy: true, error: null });
    getOpenapiBridge()
      .convertV2(specPath.trim())
      .then((_yaml) => {
        patch({ busy: false, error: null });
        // For demo, replace specPath's content preview — in real desktop, would write file
        // Show converted YAML in error slot as info
      })
      .catch((e) => {
        patch({ error: e instanceof Error ? e.message : String(e), busy: false });
      });
  };

  const toggle = (method: string, path: string): void => {
    const key = `${method}|${path}`;
    patch({
      selected: selected.includes(key)
        ? selected.filter((k) => k !== key)
        : [...selected, key],
    });
  };

  const generate = (): void => {
    if (specPath.trim() === "" || selected.length === 0 || dirName.trim() === "") return;
    patch({ busy: true, error: null });
    const selections = selected.map((k) => {
      // SAFETY: keys are built above as `METHOD|path` pairs.
      const idx = k.indexOf("|");
      return { method: k.slice(0, idx), path: k.slice(idx + 1) };
    });
    getOpenapiBridge()
      .generate({ specPath: specPath.trim(), selections, dirName: dirName.trim() })
      .then((res) => {
        patch({ generated: res.created, busy: false });
        void refreshWorkspace();
      })
      .catch((e) => {
        patch({ error: e instanceof Error ? e.message : String(e), busy: false });
      });
  };

  return (
    <section className="flex h-full min-h-0 flex-col overflow-y-auto" aria-label="OpenAPI explorer">
      <PageHeader
        icon={Compass}
        title="OpenAPI Explorer"
        description="Browse operations in an OpenAPI spec and generate workspace request files"
      />
      <div className="flex flex-col gap-3 p-4">
      <div className="flex items-end gap-2">
        <div className="flex min-w-64 flex-1 flex-col gap-1">
          <label htmlFor="openapi-spec" className="text-xs font-medium">
            Spec (workspace-relative JSON/YAML)
          </label>
          <Input
            id="openapi-spec"
            value={specPath}
            onChange={(e) => patch({ specPath: e.target.value })}
            placeholder="specs/pets.yaml"
            spellCheck={false}
            className="font-mono text-xs"
          />
        </div>
        <Button size="sm" disabled={busy || specPath.trim() === ""} onClick={explore}>
          {busy ? <Spinner data-icon="inline-start" /> : <RefreshCw data-icon="inline-start" />}
          Explore
        </Button>
        <Button size="sm" variant="outline" disabled={busy || specPath.trim() === ""} onClick={validate}>
          Validate
        </Button>
        <Button size="sm" variant="outline" disabled={busy || specPath.trim() === ""} onClick={convertV2}>
          Convert V2 → 3.0
        </Button>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {result && (
        <>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <p className="text-xs text-muted-foreground">
              <span className="font-medium text-foreground">{result.title}</span>
              {result.version != null && result.version !== "" && ` · v${result.version}`}
              {" · "}
              {result.endpoints.length} operations
            </p>
            <Input
              value={searchFilter}
              onChange={(e) => setSearchFilter(e.target.value)}
              placeholder="Filter paths, methods, tags..."
              className="h-7 w-52 font-mono text-xs"
            />
          </div>

          {grouped.map(([tag, eps]) => (
            <div key={tag} className="rounded-md border border-border p-2">
              <p className="pb-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                {tag} · {eps.length}
              </p>
              <ul className="flex flex-col divide-y divide-border/60">
                {eps.map((ep) => {
                  const key = `${ep.method}|${ep.path}`;
                  const checked = selectedSet.has(key);
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
                            METHOD_COLOR_MAP.get(ep.method) ?? "",
                          )}
                        >
                          {ep.method}
                        </span>
                        <button
                          type="button"
                          className="min-w-0 flex-1 truncate text-left font-mono text-xs hover:text-foreground"
                          title={`${ep.operationId ?? ""} ${ep.summary ?? ""}`}
                          onClick={() => toggle(ep.method, ep.path)}
                        >
                          {ep.path}
                        </button>
                        {ep.summary && (
                          <span className="hidden truncate text-[10px] text-muted-foreground sm:block">
                            {ep.summary}
                          </span>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 px-2 text-[11px] text-primary"
                          title="Try in Request Builder"
                          onClick={() => {
                            const id = `draft-spec-${Date.now()}`;
                            ensureDraft(id, {
                              method: ep.method,
                              url: ep.path,
                            });
                            openTab({
                              id,
                              title: `${ep.method} ${ep.path}`,
                              kind: "request",
                            });
                            requestView("requests");
                          }}
                        >
                          <Send className="size-3" data-icon="inline-start" />
                          Try
                        </Button>
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
                onChange={(e) => patch({ dirName: e.target.value })}
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
      </div>
    </section>
  );
}
