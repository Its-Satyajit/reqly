import { useEffect, useRef, useState } from "react";
import { ChevronDown, ChevronRight, PlugZap, Send, Square } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import type { KeyValueRow } from "#lib/request";
import { cn } from "#lib/utils";
import { KeyValueEditor } from "#components/KeyValueEditor";
import {
  formatFrameTime,
  type RealtimeKind,
} from "#lib/realtime";
import { useRealtimeStore } from "#stores/useRealtimeStore";

const KIND_OPTIONS = [
  { value: "ws", label: "WebSocket" },
  { value: "sse", label: "SSE" },
];

function statusBadge(status: string) {
  switch (status) {
    case "connected":
      return <Badge variant="secondary" className="text-status-ok">connected</Badge>;
    case "connecting":
      return (
        <Badge variant="secondary" className="flex items-center gap-1">
          <Spinner className="size-2.5" /> connecting
        </Badge>
      );
    case "error":
      return <Badge variant="destructive">error</Badge>;
    case "closed":
      return <Badge variant="ghost">closed</Badge>;
    default:
      return <Badge variant="ghost">idle</Badge>;
  }
}

export function RealtimeTab({ tabId }: { tabId: string }) {
  const tab = useRealtimeStore((s) => s.tabs[tabId]);
  const update = useRealtimeStore((s) => s.update);
  const connect = useRealtimeStore((s) => s.connect);
  const send = useRealtimeStore((s) => s.send);
  const disconnect = useRealtimeStore((s) => s.disconnect);
  const [draft, setDraft] = useState("");
  const [headersOpen, setHeadersOpen] = useState(false);
  const logRef = useRef<HTMLDivElement>(null);

  // Keep the inspector pinned to the newest frame while streaming.
  useEffect(() => {
    if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight;
  }, [tab?.frames.length]);

  if (!tab) return null;
  const isWS = tab.kind === "ws";

  return (
    <div className="flex h-full min-h-0 flex-col gap-2 p-3">
      <div className="flex items-center gap-1.5">
        <select
          value={tab.kind}
          disabled={tab.status === "connected" || tab.status === "connecting"}
          onChange={(e) => {
            // SAFETY: options are exactly the two supported realtime kinds.
            update(tabId, { kind: e.target.value as RealtimeKind });
          }}
          aria-label="Protocol"
          className="rounded-md border border-border bg-transparent px-2 py-1 text-xs"
        >
          {KIND_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
        <Input
          value={tab.url}
          onChange={(e) => update(tabId, { url: e.target.value })}
          placeholder={isWS ? "wss://echo.websocket.org" : "https://example.com/events"}
          spellCheck={false}
          className="flex-1 font-mono text-xs"
          aria-label="Endpoint URL"
        />
        {tab.status === "connected" || tab.status === "connecting" ? (
          <Button variant="outline" size="sm" onClick={() => void disconnect(tabId)}>
            {tab.status === "connecting" ? <Spinner data-icon="inline-start" /> : <Square data-icon="inline-start" />}
            Disconnect
          </Button>
        ) : (
          <Button size="sm" onClick={() => void connect(tabId)} disabled={tab.url.trim() === ""}>
            <PlugZap data-icon="inline-start" />
            Connect
          </Button>
        )}
        {statusBadge(tab.status)}
      </div>

      <div className="flex flex-col">
        <button
          type="button"
          className="flex w-fit items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
          aria-expanded={headersOpen}
          onClick={() => setHeadersOpen(!headersOpen)}
        >
          {headersOpen ? <ChevronDown className="size-3.5" /> : <ChevronRight className="size-3.5" />}
          Headers {tab.headers.length > 0 && `(${tab.headers.filter((h) => h.enabled).length})`}
        </button>
        {headersOpen && (
          <div
            className={cn(
              "pt-1",
              (tab.status === "connected" || tab.status === "connecting") && "pointer-events-none opacity-50",
            )}
            title={tab.status === "connected" ? "Disconnect before editing headers" : undefined}
          >
            <KeyValueEditor
              rows={tab.headers}
              onChange={(rows: KeyValueRow[]) => update(tabId, { headers: rows })}
              keyPlaceholder="header name"
              valuePlaceholder="value"
            />
          </div>
        )}
      </div>

      {tab.error && (
        <Alert variant="destructive">
          <AlertDescription>{tab.error}</AlertDescription>
        </Alert>
      )}

      <div
        ref={logRef}
        className="min-h-0 flex-1 overflow-y-auto rounded-md border border-border p-2"
        aria-label="Message log"
      >
        {tab.frames.length === 0 ? (
          <p className="text-xs text-muted-foreground">
            No messages yet. Frames appear here with timestamps.
          </p>
        ) : (
          <ul className="flex flex-col gap-1">
            {tab.frames.map((f) => (
              <li
                key={f.seq}
                className={cn(
                  "rounded border border-border/60 px-2 py-1 font-mono text-[11px]",
                  f.direction === "out" && "bg-muted/40",
                )}
              >
                <span className="mr-2 text-muted-foreground">{formatFrameTime(f.timestamp)}</span>
                <span
                  className={cn(
                    "mr-2 font-sans font-semibold",
                    f.direction === "out" ? "text-status-info" : "text-status-ok",
                  )}
                >
                  {f.direction === "out" ? "↑ out" : isWS ? "↓ in" : "event"}
                </span>
                {(f.name || f.id) && (
                  <span className="mr-2 font-sans text-muted-foreground">
                    {f.name || "message"}
                    {f.id ? ` #${f.id}` : ""}
                  </span>
                )}
                <span className="break-all whitespace-pre-wrap">{f.data}</span>
              </li>
            ))}
          </ul>
        )}
      </div>

      {isWS && (
        <form
          className="flex items-center gap-1.5"
          onSubmit={(e) => {
            e.preventDefault();
            if (draft.trim() === "" || tab.status !== "connected") return;
            void send(tabId, draft);
            setDraft("");
          }}
        >
          <Input
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            placeholder='Message to send, e.g. {"action":"ping"}'
            spellCheck={false}
            className="flex-1 font-mono text-xs"
            aria-label="Message to send"
            disabled={tab.status !== "connected"}
          />
          <Button type="submit" size="sm" disabled={draft.trim() === "" || tab.status !== "connected"}>
            <Send data-icon="inline-start" />
            Send
          </Button>
        </form>
      )}
    </div>
  );
}
