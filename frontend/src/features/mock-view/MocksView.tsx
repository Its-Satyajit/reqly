import { useEffect } from "react";
import { Antenna, Play, Plus, Square, Trash2 } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { Textarea } from "#components/ui/textarea";
import { cn } from "#lib/utils";
import { MOCK_METHOD_OPTIONS } from "#lib/mock";
import { useMockStore } from "#stores/useMockStore";

/** Parses a numeric input value, falling back when the text isn't a number. */
function inputInt(value: string, fallback: number): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

const numberInput =
  "rounded-md border border-border bg-transparent px-2 py-1 text-xs font-mono w-24";

export function MocksView() {
  const specPath = useMockStore((s) => s.specPath);
  const port = useMockStore((s) => s.port);
  const delayMs = useMockStore((s) => s.delayMs);
  const failEvery = useMockStore((s) => s.failEvery);
  const routes = useMockStore((s) => s.routes);
  const status = useMockStore((s) => s.status);
  const busy = useMockStore((s) => s.busy);
  const error = useMockStore((s) => s.error);
  const setSpecPath = useMockStore((s) => s.setSpecPath);
  const setPort = useMockStore((s) => s.setPort);
  const setDelayMs = useMockStore((s) => s.setDelayMs);
  const setFailEvery = useMockStore((s) => s.setFailEvery);
  const updateRoute = useMockStore((s) => s.updateRoute);
  const addRoute = useMockStore((s) => s.addRoute);
  const removeRoute = useMockStore((s) => s.removeRoute);
  const start = useMockStore((s) => s.start);
  const stop = useMockStore((s) => s.stop);

  // Re-sync the panel when the backend restarts outside this view.
  useEffect(() => {
    void useMockStore.getState().refreshStatus();
  }, []);

  return (
    <section className="flex h-full min-h-0 flex-col overflow-y-auto" aria-label="Mock server">
      <PageHeader
        icon={Antenna}
        title="Mock Server"
        description="Serve schema- and example-driven mock responses from an OpenAPI spec or manual routes"
        actions={
          <div className="flex items-center gap-2">
            <span
              title={status.running ? "Running" : status.error ? "Crashed" : "Stopped"}
              className={cn("size-2 rounded-full", status.running ? "bg-status-ok" : status.error ? "bg-destructive" : "bg-muted-foreground")}
              aria-label={status.running ? "Running" : "Stopped"}
            />
            <span className="text-xs text-muted-foreground">{status.running ? "Running" : status.error ? "Crashed" : "Stopped"}</span>
            {status.url && (
              <Badge variant="secondary" className="font-mono text-status-ok">
                {status.url}
              </Badge>
            )}
            {status.running ? (
              <Button variant="outline" size="sm" disabled={busy} onClick={() => void stop()}>
                {busy ? <Spinner data-icon="inline-start" /> : <Square data-icon="inline-start" />}
                Stop
              </Button>
            ) : status.error ? (
              <Button size="sm" disabled={busy} onClick={() => void start()}>
                {busy ? <Spinner data-icon="inline-start" /> : <Play data-icon="inline-start" />}
                Restart
              </Button>
            ) : (
              <Button size="sm" disabled={busy} onClick={() => void start()}>
                {busy ? <Spinner data-icon="inline-start" /> : <Play data-icon="inline-start" />}
                Start
              </Button>
            )}
          </div>
        }
      />
      <div className="flex flex-col gap-3 p-4">

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="flex flex-wrap items-end gap-3 rounded-md border border-border p-3">
        <div className="flex min-w-64 flex-1 flex-col gap-1">
          <label htmlFor="mock-spec" className="text-xs font-medium">
            OpenAPI spec (optional — manual routes below always apply)
          </label>
          <Input
            id="mock-spec"
            value={specPath}
            onChange={(e) => setSpecPath(e.target.value)}
            placeholder="specs/pets.yaml"
            spellCheck={false}
            className="font-mono text-xs"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="mock-port" className="text-xs font-medium">Port</label>
          <input
            id="mock-port"
            type="number"
            value={port}
            onChange={(e) => setPort(inputInt(e.target.value, port))}
            className={numberInput}
            min={1}
            max={65535}
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="mock-delay" className="text-xs font-medium">Delay (ms)</label>
          <input
            id="mock-delay"
            type="number"
            value={delayMs}
            onChange={(e) => setDelayMs(inputInt(e.target.value, delayMs))}
            className={numberInput}
            min={0}
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="mock-fail" className="text-xs font-medium">Fail every Nth</label>
          <input
            id="mock-fail"
            type="number"
            value={failEvery}
            onChange={(e) => setFailEvery(inputInt(e.target.value, failEvery))}
            className={numberInput}
            min={0}
          />
        </div>
      </div>

      <div className="flex items-center justify-between">
        <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
          Manual routes ({routes.filter((r) => r.enabled).length}/{routes.length})
        </p>
        <Button variant="outline" size="sm" onClick={addRoute}>
          <Plus data-icon="inline-start" />
          Add route
        </Button>
      </div>

      <ul className="flex flex-col gap-2 pb-4">
        {routes.map((route, i) => (
          <li key={route.id} className="flex flex-col gap-2 rounded-md border border-border p-3">
            <div className="flex items-center gap-1.5">
              <input
                type="checkbox"
                checked={route.enabled}
                onChange={(e) => updateRoute(i, { enabled: e.target.checked })}
                title={route.enabled ? "Enabled" : "Disabled"}
                aria-label={`Route ${route.path} enabled`}
                className="size-3.5 shrink-0 accent-(--primary)"
              />
              <select
                value={route.method}
                onChange={(e) => updateRoute(i, { method: e.target.value })}
                aria-label={`Route ${route.path} method`}
                className="w-20 rounded-md border border-border bg-transparent px-1 py-1 text-xs"
              >
                {MOCK_METHOD_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>{o.label}</option>
                ))}
              </select>
              <Input
                value={route.path}
                onChange={(e) => updateRoute(i, { path: e.target.value })}
                placeholder="/path"
                spellCheck={false}
                aria-label={`Route ${route.path} path`}
                className="min-w-0 flex-1 font-mono text-xs"
              />
              <input
                type="number"
                value={route.status || 200}
                onChange={(e) => updateRoute(i, { status: inputInt(e.target.value, route.status || 200) })}
                aria-label={`Route ${route.path} status`}
                className={cn(numberInput)}
                min={100}
                max={599}
              />
              <Button
                variant="ghost"
                size="sm"
                aria-label={`Remove route ${route.path}`}
                onClick={() => removeRoute(i)}
              >
                <Trash2 />
              </Button>
            </div>
            <Textarea
              value={route.body}
              onChange={(e) => updateRoute(i, { body: e.target.value })}
              rows={3}
              spellCheck={false}
              aria-label={`Route ${route.path} response body`}
              placeholder='Response body, e.g. {"ok":true}'
              className="resize-y font-mono text-xs"
            />
            <Textarea
              value={route.headerLines.join("\n")}
              onChange={(e) =>
                updateRoute(i, {
                  headerLines: e.target.value.split("\n"),
                })
              }
              rows={2}
              spellCheck={false}
              aria-label={`Route ${route.path} response headers`}
              placeholder={"Content-Type: application/json\nX-Mock: true"}
              className="resize-y font-mono text-[11px]"
            />
            <div className="flex gap-2">
              <label className="flex flex-col gap-1 text-xs">
                Latency (ms)
                <input
                  type="number"
                  value={route.latencyMs ?? 0}
                  onChange={(e) => updateRoute(i, { latencyMs: inputInt(e.target.value, 0) })}
                  aria-label={`Route ${route.path} latency`}
                  className={numberInput}
                  min={0}
                />
              </label>
              <label className="flex flex-col gap-1 text-xs">
                Scenario
                <Input
                  value={route.scenario ?? ""}
                  onChange={(e) => updateRoute(i, { scenario: e.target.value })}
                  placeholder="default"
                  aria-label={`Route ${route.path} scenario`}
                  className="w-32 font-mono text-xs"
                />
              </label>
              <span className="self-end text-[11px] text-muted-foreground">Body file via body field (JSON/text)</span>
            </div>
          </li>
        ))}
      </ul>
      </div>
    </section>
  );
}
