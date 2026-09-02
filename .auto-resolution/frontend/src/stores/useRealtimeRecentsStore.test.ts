import { describe, it, expect, beforeEach } from "vitest";
import { useRealtimeRecentsStore } from "./useRealtimeRecentsStore";

describe("useRealtimeRecentsStore", () => {
  beforeEach(() => useRealtimeRecentsStore.setState({ recents: [] }));

  it("dedupes and caps at 12", () => {
    for (let i = 0; i < 15; i++) useRealtimeRecentsStore.getState().addRecent(`wss://a/${i}`, "ws");
    expect(useRealtimeRecentsStore.getState().recents.length).toBe(12);
    useRealtimeRecentsStore.getState().addRecent("wss://a/14", "ws");
    expect(useRealtimeRecentsStore.getState().recents[0].url).toBe("wss://a/14");
    expect(useRealtimeRecentsStore.getState().recents.length).toBe(12);
  });

  it("filters by kind", () => {
    useRealtimeRecentsStore.getState().addRecent("wss://ws", "ws");
    useRealtimeRecentsStore.getState().addRecent("https://sse", "sse");
    expect(useRealtimeRecentsStore.getState().forKind("ws").length).toBe(1);
    expect(useRealtimeRecentsStore.getState().forKind("sse").length).toBe(1);
  });
});
