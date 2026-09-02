import { describe, it, expect, beforeEach } from "vitest";
import { useSchemaDocStore } from "./useSchemaDocStore";
import type { SchemaDocTree } from "#lib/schemaDocs";

const emptyTree: SchemaDocTree = { types: [] };

describe("useSchemaDocStore", () => {
  beforeEach(() =>
    useSchemaDocStore.setState({
      schemaTree: emptyTree,
      searchQuery: "",
      selectedTypeName: null,
    }),
  );

  it("starts empty", () => {
    const s = useSchemaDocStore.getState();
    expect(s.schemaTree.types).toHaveLength(0);
    expect(s.searchQuery).toBe("");
  });

  it("setSchemaTree sets the tree", () => {
    const tree: SchemaDocTree = {
      types: [{ name: "User", kind: "type", fields: [{ name: "id", type: "ID!", isRequired: true }] }],
    };
    useSchemaDocStore.getState().setSchemaTree(tree);
    expect(useSchemaDocStore.getState().schemaTree.types).toHaveLength(1);
  });

  it("setSearchQuery sets query", () => {
    useSchemaDocStore.getState().setSearchQuery("User");
    expect(useSchemaDocStore.getState().searchQuery).toBe("User");
  });

  it("setSelectedType sets selected type", () => {
    useSchemaDocStore.getState().setSelectedType("User");
    expect(useSchemaDocStore.getState().selectedTypeName).toBe("User");
  });

  it("getSearchResults returns filtered types", () => {
    useSchemaDocStore.getState().setSchemaTree({
      types: [
        { name: "User", kind: "type" },
        { name: "Post", kind: "type" },
      ],
    });
    useSchemaDocStore.getState().setSearchQuery("User");
    const results = useSchemaDocStore.getState().getSearchResults();
    expect(results).toHaveLength(1);
    expect(results[0].name).toBe("User");
  });
});
