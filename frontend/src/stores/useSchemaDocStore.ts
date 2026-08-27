import { create } from "zustand";
import { searchSchemaDocs, type SchemaDocTree, type SchemaDocType } from "#lib/schemaDocs";

interface SchemaDocState {
  schemaTree: SchemaDocTree;
  searchQuery: string;
  selectedTypeName: string | null;
  setSchemaTree(tree: SchemaDocTree): void;
  setSearchQuery(query: string): void;
  setSelectedType(name: string | null): void;
  getSearchResults(): SchemaDocType[];
}

export const useSchemaDocStore = create<SchemaDocState>((set, get) => ({
  schemaTree: { types: [] },
  searchQuery: "",
  selectedTypeName: null,

  setSchemaTree(tree) {
    set({ schemaTree: tree });
  },

  setSearchQuery(query) {
    set({ searchQuery: query });
  },

  setSelectedType(name) {
    set({ selectedTypeName: name });
  },

  getSearchResults() {
    const { schemaTree, searchQuery } = get();
    if (!searchQuery.trim()) return schemaTree.types;
    return searchSchemaDocs(schemaTree, searchQuery);
  },
}));
