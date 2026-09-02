import { describe, it, expect } from "vitest";
import { nodesForSpec, edgesForNodes, SCHEMA_NODES, clampZoom, getNodeAt, getTransform } from "./schemaGraph";

describe("schemaGraph", () => {
  it("returns single node for non-components selection", () => {
    expect(nodesForSpec("info")).toHaveLength(1);
  });
  it("returns full graph for components", () => {
    expect(nodesForSpec("components:schemas")).toHaveLength(5);
  });
  it("filters edges to visible nodes", () => {
    expect(edgesForNodes(SCHEMA_NODES.slice(0, 1))).toHaveLength(0);
    expect(edgesForNodes(SCHEMA_NODES)).toHaveLength(4);
  });
});

describe("interactive helpers", () => {
  it("clamps zoom between 0.5 and 2", () => {
    expect(clampZoom(0.1)).toBe(0.5);
    expect(clampZoom(3)).toBe(2);
    expect(clampZoom(1)).toBe(1);
    expect(clampZoom(1.5)).toBe(1.5);
  });

  it("returns node at position within bounds", () => {
    // Node User at 200,40 with 88x36 size (centered at x-44)
    const node = getNodeAt(SCHEMA_NODES, 200, 40);
    expect(node?.id).toBe("User");
  });

  it("returns undefined for empty area", () => {
    expect(getNodeAt(SCHEMA_NODES, 0, 0)).toBeUndefined();
  });

  it("computes transform string for zoom/pan", () => {
    expect(getTransform(1, 0, 0)).toBe("translate(0 0) scale(1)");
    expect(getTransform(1.5, 10, 20)).toBe("translate(10 20) scale(1.5)");
  });

  it("clamps zoom in transform", () => {
    expect(getTransform(5, 0, 0)).toBe("translate(0 0) scale(2)");
  });
});
