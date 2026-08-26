import { describe, it, expect, beforeEach, vi } from "vitest";
import { useCommandPaletteStore } from "./useCommandPaletteStore";

describe("useCommandPaletteStore", () => {
  beforeEach(() => useCommandPaletteStore.setState({ commands: [], providers: [], query: "", open: false }));

  it("registers and filters via fuse", () => {
    useCommandPaletteStore.getState().registerCommand({ id: "a", title: "Open Settings", keywords: "settings", run: vi.fn() });
    useCommandPaletteStore.getState().registerCommand({ id: "b", title: "New Request", keywords: "request", run: vi.fn() });
    useCommandPaletteStore.getState().setQuery("sett");
    const res = useCommandPaletteStore.getState().filtered();
    expect(res[0].id).toBe("a");
  });

  it("runs command", () => {
    const fn = vi.fn();
    useCommandPaletteStore.getState().registerCommand({ id: "x", title: "Theme", run: fn });
    useCommandPaletteStore.getState().filtered()[0].run();
    expect(fn).toHaveBeenCalled();
  });

  it("empty query returns top commands capped", () => {
    for (let i = 0; i < 25; i++) useCommandPaletteStore.getState().registerCommand({ id: `${i}`, title: `Cmd ${i}`, run: vi.fn() });
    expect(useCommandPaletteStore.getState().filtered().length).toBe(20);
  });
});
