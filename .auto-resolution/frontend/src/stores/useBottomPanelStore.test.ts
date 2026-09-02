import { describe, it, expect, beforeEach } from "vitest";
import { useBottomPanelStore } from "./useBottomPanelStore";

describe("useBottomPanelStore", () => {
  beforeEach(() => useBottomPanelStore.setState({ active: null, collapsed: true }));
  it("opens panel on toggle", () => {
    useBottomPanelStore.getState().toggle("console");
    expect(useBottomPanelStore.getState().active).toBe("console");
    expect(useBottomPanelStore.getState().collapsed).toBe(false);
  });
  it("collapses when toggling active panel again", () => {
    useBottomPanelStore.getState().toggle("network");
    useBottomPanelStore.getState().toggle("network");
    expect(useBottomPanelStore.getState().collapsed).toBe(true);
  });
  it("switches panel without collapsing", () => {
    useBottomPanelStore.getState().toggle("console");
    useBottomPanelStore.getState().toggle("variables");
    expect(useBottomPanelStore.getState().active).toBe("variables");
    expect(useBottomPanelStore.getState().collapsed).toBe(false);
  });
  it("setActive null collapses", () => {
    useBottomPanelStore.getState().setActive("tests");
    useBottomPanelStore.getState().setActive(null);
    expect(useBottomPanelStore.getState().collapsed).toBe(true);
  });
});
