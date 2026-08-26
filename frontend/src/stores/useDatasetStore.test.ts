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
});
