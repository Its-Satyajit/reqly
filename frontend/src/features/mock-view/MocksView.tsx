import { useEffect, useState } from "react";
import { Antenna, Play, Plus, Square, Trash2, Sliders, ShieldAlert, ListFilter, Activity } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { Textarea } from "#components/ui/textarea";
import { CompactSelect } from "#components/CompactSelect";
import { cn } from "#lib/utils";
import { MOCK_METHOD_OPTIONS, type FaultType } from "#lib/mock";
import { useMockStore } from "#stores/useMockStore";

function isFaultType(v: string): v is FaultType {
  return v === "delay" || v === "drop" || v === "error" || v === "corrupt";
}

/** Parses a numeric input value, falling back when the text isn't a number. */
function inputInt(value: string, fallback: number): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

const numberInput =
  "rounded-md border border-border bg-transparent px-2 py-1 text-xs font-mono w-24";

type MockTab = "routes" | "scenarios" | "faults" | "logs";

export function MocksView() {
  const [activeTab, setActiveTab] = useState<MockTab>("routes");
  const specPath = useMockStore((s) => s.specPath);
  const port = useMockStore((s) => s.port);
  const delayMs = useMockStore((s) => s.delayMs);
  const failEvery = useMockStore((s) => s.failEvery);
  const routes = useMockStore((s) => s.routes);
  const scenarios = useMockStore((s) => s.scenarios);
  const activeScenarioId = useMockStore((s) => s.activeScenarioId);
  const faultInjection = useMockStore((s) => s.faultInjection);
  const logs = useMockStore((s) => s.logs);
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
  const createScenario = useMockStore((s) => s.createScenario);
  const deleteScenario = useMockStore((s) => s.deleteScenario);
  const setActiveScenario = useMockStore((s) => s.setActiveScenario);
  const setFaultInjection = useMockStore((s) => s.setFaultInjection);
  const clearMockLogs = useMockStore((s) => s.clearMockLogs);
  const start = useMockStore((s) => s.start);
  const stop = useMockStore((s) => s.stop);

  const [newScenarioName, setNewScenarioName] = useState("");

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

      {/* Subtabs: Routes, Scenarios, Fault Injection, Logs */}
      <div className="flex items-center gap-1 border-b border-border pb-2">
        <button
          type="button"
          onClick={() => setActiveTab("routes")}
          className={`flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
            activeTab === "routes"
              ? "bg-primary/10 text-primary"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <Sliders className="size-3.5" />
          Routes ({routes.length})
        </button>
        <button
          type="button"
          onClick={() => setActiveTab("scenarios")}
          className={`flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
            activeTab === "scenarios"
              ? "bg-primary/10 text-primary"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <ListFilter className="size-3.5" />
          Scenarios ({scenarios.length})
        </button>
        <button
          type="button"
          onClick={() => setActiveTab("faults")}
          className={`flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
            activeTab === "faults"
              ? "bg-primary/10 text-primary"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <ShieldAlert className="size-3.5" />
          Fault Injection {faultInjection.enabled && "•"}
        </button>
        <button
          type="button"
          onClick={() => setActiveTab("logs")}
          className={`flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
            activeTab === "logs"
              ? "bg-primary/10 text-primary"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <Activity className="size-3.5" />
          Logs ({logs.length})
        </button>
      </div>

      {activeTab === "routes" && (
        <>
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
        </>
      )}

      {activeTab === "scenarios" && (
        <div className="space-y-4">
          <div className="flex gap-2">
            <Input
              value={newScenarioName}
              onChange={(e) => setNewScenarioName(e.target.value)}
              placeholder="New Scenario name (e.g. RateLimited, UserNotFound)"
              className="h-8 font-mono text-xs"
            />
            <Button
              size="sm"
              onClick={() => {
                if (newScenarioName.trim()) {
                  createScenario(newScenarioName.trim());
                  setNewScenarioName("");
                }
              }}
              disabled={!newScenarioName.trim()}
            >
              <Plus data-icon="inline-start" />
              Add Scenario
            </Button>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            {scenarios.map((sc) => (
              <div
                key={sc.id}
                className={`rounded-md border p-3 flex flex-col gap-2 ${
                  activeScenarioId === sc.id
                    ? "border-primary bg-primary/5"
                    : "border-border bg-card"
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="font-semibold text-xs">{sc.name}</span>
                  <div className="flex items-center gap-1">
                    <Button
                      size="sm"
                      variant={activeScenarioId === sc.id ? "default" : "outline"}
                      className="h-6 text-[10px] px-2"
                      onClick={() =>
                        setActiveScenario(
                          activeScenarioId === sc.id ? null : sc.id,
                        )
                      }
                    >
                      {activeScenarioId === sc.id ? "Active" : "Activate"}
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      className="h-6 w-6 p-0"
                      onClick={() => deleteScenario(sc.id)}
                    >
                      <Trash2 className="size-3 text-destructive" />
                    </Button>
                  </div>
                </div>
                {sc.description && (
                  <p className="text-[11px] text-muted-foreground">
                    {sc.description}
                  </p>
                )}
                <span className="text-[10px] text-muted-foreground font-mono">
                  {sc.routes.length} route overrides
                </span>
              </div>
            ))}
            {scenarios.length === 0 && (
              <p className="text-xs text-muted-foreground sm:col-span-2 py-4 text-center">
                No scenarios configured yet. Create one to test alternative backend behaviors.
              </p>
            )}
          </div>
        </div>
      )}

      {activeTab === "faults" && (
        <div className="rounded-md border border-border bg-card p-4 space-y-3">
          <label className="flex items-center gap-2 text-xs font-medium cursor-pointer">
            <input
              type="checkbox"
              checked={faultInjection.enabled}
              onChange={(e) => setFaultInjection({ enabled: e.target.checked })}
              className="size-3.5 rounded border-border text-primary focus:ring-primary"
            />
            Enable Fault Injection
          </label>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
            <div className="flex flex-col gap-1">
              <span className="text-[11px] text-muted-foreground">Fault Type</span>
              <CompactSelect
                value={faultInjection.type}
                onChange={(v) => {
                  if (isFaultType(v)) {
                    setFaultInjection({ type: v });
                  }
                }}
                options={[
                  { value: "delay", label: "Artificial Latency" },
                  { value: "error", label: "HTTP Error Code" },
                  { value: "drop", label: "Drop Connection" },
                  { value: "corrupt", label: "Corrupt Payload" },
                ]}
                ariaLabel="Fault Type"
              />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-[11px] text-muted-foreground">
                Probability: {Math.round(faultInjection.probability * 100)}%
              </span>
              <input
                type="range"
                min={0}
                max={1}
                step={0.05}
                value={faultInjection.probability}
                aria-label="Fault probability"
                onChange={(e) =>
                  setFaultInjection({ probability: Number(e.target.value) })
                }
                className="accent-primary"
              />
            </div>
            {faultInjection.type === "delay" && (
              <div className="flex flex-col gap-1">
                <span className="text-[11px] text-muted-foreground">Delay (ms)</span>
                <Input
                  type="number"
                  min={0}
                  step={50}
                  value={faultInjection.delayMs}
                  aria-label="Fault delay milliseconds"
                  onChange={(e) =>
                    setFaultInjection({ delayMs: Number(e.target.value) })
                  }
                  className="font-mono text-xs"
                />
              </div>
            )}
            {faultInjection.type === "error" && (
              <div className="flex flex-col gap-1">
                <span className="text-[11px] text-muted-foreground">Status Code</span>
                <Input
                  type="number"
                  value={faultInjection.errorCode ?? 500}
                  onChange={(e) =>
                    setFaultInjection({ errorCode: Number(e.target.value) || 500 })
                  }
                  className="h-8 font-mono text-xs"
                />
              </div>
            )}
          </div>
        </div>
      )}

      {activeTab === "logs" && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium">Recent Mock Server Invocations</span>
            <Button
              size="sm"
              variant="ghost"
              className="h-6 text-[11px]"
              onClick={clearMockLogs}
            >
              Clear Logs
            </Button>
          </div>
          {logs.length === 0 ? (
            <p className="py-8 text-center text-xs text-muted-foreground">
              No requests received by the mock server yet.
            </p>
          ) : (
            <div className="rounded border border-border bg-background overflow-x-auto font-mono text-xs">
              <table className="w-full">
                <thead className="text-left text-muted-foreground border-b border-border bg-muted/20">
                  <tr>
                    <th className="px-2.5 py-1.5 font-medium">Time</th>
                    <th className="px-2.5 py-1.5 font-medium">Method</th>
                    <th className="px-2.5 py-1.5 font-medium">Path</th>
                    <th className="px-2.5 py-1.5 font-medium">Status</th>
                    <th className="px-2.5 py-1.5 font-medium">Duration</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((l, i) => (
                    <tr key={`${l.timestamp}-${l.method}-${l.path}-${i}`} className="border-b border-border/50 last:border-0">
                      <td className="px-2.5 py-1 text-muted-foreground">
                        {new Date(l.timestamp).toLocaleTimeString()}
                      </td>
                      <td className="px-2.5 py-1 font-semibold text-primary">{l.method}</td>
                      <td className="px-2.5 py-1 truncate max-w-[200px]">{l.path}</td>
                      <td className="px-2.5 py-1">
                        <span
                          className={
                            l.status >= 400 ? "text-destructive" : "text-status-ok"
                          }
                        >
                          {l.status}
                        </span>
                      </td>
                      <td className="px-2.5 py-1 text-muted-foreground">{l.duration}ms</td>
                    </tr>
                  ))}
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
