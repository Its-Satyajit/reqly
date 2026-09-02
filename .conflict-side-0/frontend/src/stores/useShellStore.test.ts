import { describe, it, expect, beforeEach, vi } from "vitest";

describe("useShellStore", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.resetModules();
  });

  it("initializes with default inspector closed and responseMode horizontal", async () => {
    const { useShellStore } = await import("./useShellStore");
    expect(useShellStore.getState().inspectorOpen).toBe(false);
    expect(useShellStore.getState().inspectorTab).toBeNull();
    expect(useShellStore.getState().responseMode).toBe("horizontal");
  });

  it("opens and closes inspector with persistence", async () => {
    const { useShellStore } = await import("./useShellStore");
    useShellStore.getState().openInspector("details");
    expect(useShellStore.getState().inspectorOpen).toBe(true);
    expect(useShellStore.getState().inspectorTab).toBe("details");
    expect(localStorage.getItem("reqly-shell-inspector-open")).toBe("1");
    expect(localStorage.getItem("reqly-shell-inspector-tab")).toBe("details");

    useShellStore.getState().closeInspector();
    expect(useShellStore.getState().inspectorOpen).toBe(false);
    expect(localStorage.getItem("reqly-shell-inspector-open")).toBeNull();
  });

  it("toggles inspector and persists open state", async () => {
    const { useShellStore } = await import("./useShellStore");
    useShellStore.getState().toggleInspector();
    expect(useShellStore.getState().inspectorOpen).toBe(true);
    expect(localStorage.getItem("reqly-shell-inspector-open")).toBe("1");

    useShellStore.getState().toggleInspector();
    expect(useShellStore.getState().inspectorOpen).toBe(false);
    expect(localStorage.getItem("reqly-shell-inspector-open")).toBeNull();
  });

  it("persists responseMode changes", async () => {
    const { useShellStore } = await import("./useShellStore");
    useShellStore.getState().setResponseMode("vertical");
    expect(useShellStore.getState().responseMode).toBe("vertical");
    expect(localStorage.getItem("reqly-shell-response-mode")).toBe("vertical");

    useShellStore.getState().toggleResponseMode();
    expect(useShellStore.getState().responseMode).toBe("horizontal");
    expect(localStorage.getItem("reqly-shell-response-mode")).toBe("horizontal");
  });
});
