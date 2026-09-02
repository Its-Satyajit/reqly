import { describe, it, expect, beforeEach } from "vitest";
import { useSpecEditorStore } from "./useSpecEditorStore";

describe("useSpecEditorStore", () => {
  beforeEach(() => useSpecEditorStore.setState({ content: "a", selectedId: "info", dirty: false }));
  it("marks dirty on content change", () => {
    useSpecEditorStore.getState().setContent("b");
    expect(useSpecEditorStore.getState().dirty).toBe(true);
  });
  it("clears dirty on save", () => {
    useSpecEditorStore.getState().setContent("b");
    useSpecEditorStore.getState().markSaved();
    expect(useSpecEditorStore.getState().dirty).toBe(false);
  });
  it("selects node", () => {
    useSpecEditorStore.getState().setSelected("paths:/users");
    expect(useSpecEditorStore.getState().selectedId).toBe("paths:/users");
  });
});
