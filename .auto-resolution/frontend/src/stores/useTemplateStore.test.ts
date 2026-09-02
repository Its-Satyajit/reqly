import { describe, it, expect, beforeEach } from "vitest";
import { useTemplateStore } from "./useTemplateStore";

describe("useTemplateStore", () => {
  beforeEach(() =>
    useTemplateStore.setState({
      selectedCategory: null,
      selectedTemplateId: null,
      customTemplates: [],
    }),
  );

  it("starts with no selection", () => {
    const s = useTemplateStore.getState();
    expect(s.selectedCategory).toBeNull();
    expect(s.selectedTemplateId).toBeNull();
  });

  it("selectCategory sets category and clears template", () => {
    useTemplateStore.getState().selectCategory("rest");
    const s = useTemplateStore.getState();
    expect(s.selectedCategory).toBe("rest");
    expect(s.selectedTemplateId).toBeNull();
  });

  it("selectTemplate sets template id", () => {
    useTemplateStore.getState().selectTemplate("rest-crud-list");
    expect(useTemplateStore.getState().selectedTemplateId).toBe("rest-crud-list");
  });

  it("addCustomTemplate appends to custom list", () => {
    useTemplateStore.getState().addCustomTemplate({
      id: "custom-1",
      category: "rest",
      name: "Custom GET",
      description: "My custom template",
      method: "GET",
      path: "/custom",
    });
    expect(useTemplateStore.getState().customTemplates).toHaveLength(1);
    expect(useTemplateStore.getState().customTemplates[0].id).toBe("custom-1");
  });

  it("removeCustomTemplate removes by id", () => {
    useTemplateStore.getState().addCustomTemplate({
      id: "custom-1",
      category: "rest",
      name: "Custom GET",
      description: "My custom template",
    });
    useTemplateStore.getState().removeCustomTemplate("custom-1");
    expect(useTemplateStore.getState().customTemplates).toHaveLength(0);
  });

  it("getAllTemplates includes built-in and custom", () => {
    useTemplateStore.getState().addCustomTemplate({
      id: "custom-1",
      category: "rest",
      name: "Custom",
      description: "Custom template",
    });
    const all = useTemplateStore.getState().getAllTemplates();
    expect(all.length).toBeGreaterThan(1);
    expect(all.some((t) => t.id === "custom-1")).toBe(true);
  });
});
