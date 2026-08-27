import { describe, it, expect } from "vitest";
import { nodesForSpec, edgesForNodes, SCHEMA_NODES } from "./schemaGraph";

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
