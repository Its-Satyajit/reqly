import { describe, it, expect, beforeEach } from "vitest";
import { useDatasetStore } from "./useDatasetStore";

describe("useDatasetStore", () => {
  beforeEach(() =>
    useDatasetStore.setState({
      dataset: null,
      config: { iterations: 1, concurrency: 1, bindTo: "request" },
      currentIteration: 0,
      isRunning: false,
      results: [],
    }),
  );

  it("starts empty", () => {
    const s = useDatasetStore.getState();
    expect(s.dataset).toBeNull();
    expect(s.config.iterations).toBe(1);
  });

  it("loadCsv parses and sets dataset", () => {
    useDatasetStore.getState().loadCsv("id,name\n1,Alice");
    const s = useDatasetStore.getState();
    expect(s.dataset).not.toBeNull();
    expect(s.dataset!.rows).toHaveLength(1);
    expect(s.dataset!.source).toBe("csv");
  });

  it("loadJson parses and sets dataset", () => {
    useDatasetStore.getState().loadJson('[{"id":1}]');
    expect(useDatasetStore.getState().dataset!.rows).toHaveLength(1);
  });

  it("setConfig patches config", () => {
    useDatasetStore.getState().setConfig({ iterations: 5, concurrency: 3 });
    const s = useDatasetStore.getState();
    expect(s.config.iterations).toBe(5);
    expect(s.config.concurrency).toBe(3);
  });

  it("getRowVariables returns values for row", () => {
    useDatasetStore.getState().loadCsv("id,name\n1,Alice");
    const vars = useDatasetStore.getState().getRowVariables(0);
    expect(vars).toEqual({ id: "1", name: "Alice" });
  });

  it("clearDataset resets state", () => {
    useDatasetStore.getState().loadCsv("id\n1");
    useDatasetStore.getState().clearDataset();
    expect(useDatasetStore.getState().dataset).toBeNull();
  });

  describe("loadDataset auto-detect", () => {
    it("auto-detects CSV and sets dataset", () => {
      useDatasetStore.getState().loadDataset("id,name\n1,Alice", "test.csv");
      const s = useDatasetStore.getState();
      expect(s.dataset).not.toBeNull();
      expect(s.dataset!.source).toBe("csv");
      expect(s.dataset!.rows).toHaveLength(1);
    });

    it("auto-detects JSON array and sets dataset", () => {
      useDatasetStore.getState().loadDataset('[{"id":1,"name":"Bob"}]', "data.json");
      const s = useDatasetStore.getState();
      expect(s.dataset).not.toBeNull();
      expect(s.dataset!.source).toBe("json");
      expect(s.dataset!.rows[0].values).toEqual({ id: "1", name: "Bob" });
    });

    it("prefers JSON when content is JSON array even with .csv name", () => {
      useDatasetStore.getState().loadDataset('[{"id":1}]', "weird.csv");
      expect(useDatasetStore.getState().dataset!.source).toBe("json");
    });

    it("falls back to CSV when JSON parse fails", () => {
      useDatasetStore.getState().loadDataset("id,name\n1,Alice", "data.json");
      // content is CSV even though name suggests JSON — should still parse as CSV
      expect(useDatasetStore.getState().dataset!.source).toBe("csv");
    });

    it("throws or validates on empty dataset (no rows)", () => {
      // empty header only should result in validation error via getValidationErrors
      useDatasetStore.getState().loadDataset("id,name", "empty.csv");
      const s = useDatasetStore.getState();
      expect(s.dataset).not.toBeNull();
      expect(s.getValidationErrors().length).toBeGreaterThan(0);
    });
  });

  describe("getValidationErrors", () => {
    it("returns empty for valid dataset", () => {
      useDatasetStore.getState().loadCsv("id,name\n1,Alice");
      expect(useDatasetStore.getState().getValidationErrors()).toEqual([]);
    });

    it("returns errors for empty dataset", () => {
      useDatasetStore.getState().loadCsv("id,name");
      expect(useDatasetStore.getState().getValidationErrors().length).toBeGreaterThan(0);
    });
  });
});
