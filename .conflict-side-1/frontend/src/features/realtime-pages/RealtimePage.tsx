import { useEffect } from "react";
import { RealtimeTab } from "../realtime-view/RealtimeTab";
import { useRealtimeStore } from "#stores/useRealtimeStore";
import { useRealtimeRecentsStore } from "#stores/useRealtimeRecentsStore";

const WS_PAGE_ID = "realtime-websocket-page";
const SSE_PAGE_ID = "realtime-sse-page";

export function WebSocketPage() {
  const tabs = useRealtimeStore((s) => s.tabs);
  const newTab = useRealtimeStore((s) => s.newTab);
  const addRecent = useRealtimeRecentsStore((s) => s.addRecent);
  const tab = tabs[WS_PAGE_ID];

  useEffect(() => {
    if (!tabs[WS_PAGE_ID]) {
      newTab(WS_PAGE_ID, "ws");
    }
  }, [tabs, newTab]);

  useEffect(() => {
    if (tab?.status === "connected" && tab?.url?.trim()) {
      addRecent(tab.url.trim(), "ws");
    }
  }, [tab?.status, tab?.url, addRecent]);

  if (!tabs[WS_PAGE_ID]) return null;
  return <RealtimeTab tabId={WS_PAGE_ID} />;
}

export function SSEPage() {
  const tabs = useRealtimeStore((s) => s.tabs);
  const newTab = useRealtimeStore((s) => s.newTab);
  const addRecent = useRealtimeRecentsStore((s) => s.addRecent);
  const tab = tabs[SSE_PAGE_ID];

  useEffect(() => {
    if (!tabs[SSE_PAGE_ID]) {
      newTab(SSE_PAGE_ID, "sse");
    }
  }, [tabs, newTab]);

  useEffect(() => {
    if (tab?.status === "connected" && tab?.url?.trim()) {
      addRecent(tab.url.trim(), "sse");
    }
  }, [tab?.status, tab?.url, addRecent]);

  if (!tabs[SSE_PAGE_ID]) return null;
  return <RealtimeTab tabId={SSE_PAGE_ID} />;
}
