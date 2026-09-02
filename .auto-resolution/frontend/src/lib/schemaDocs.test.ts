import { describe, it, expect } from "vitest";
import {
  parseGraphQlSchema,
  searchSchemaDocs,
  renderDocMarkdown,
  type SchemaDocType,
} from "./schemaDocs";

describe("schemaDocs lib", () => {
  const sampleSdl = `
    type User {
      id: ID!
      name: String!
      email: String
      posts: [Post!]!
    }

    type Post {
      id: ID!
      title: String!
      author: User!
    }

    type Query {
      user(id: ID!): User
      posts: [Post!]!
    }

    type Mutation {
      createUser(name: String!, email: String): User!
    }

    enum Role {
      ADMIN
      USER
      GUEST
    }
  `;

  describe("parseGraphQlSchema", () => {
    it("parses types from SDL", () => {
      const tree = parseGraphQlSchema(sampleSdl);
      expect(tree.types.length).toBeGreaterThan(0);
      const user = tree.types.find((t) => t.name === "User");
      expect(user).toBeDefined();
      expect(user!.kind).toBe("type");
    });
    it("parses fields with types", () => {
      const tree = parseGraphQlSchema(sampleSdl);
      const user = tree.types.find((t) => t.name === "User");
      expect(user!.fields).toBeDefined();
      expect(user!.fields!.length).toBe(4);
      expect(user!.fields![0].name).toBe("id");
      expect(user!.fields![0].isRequired).toBe(true);
    });
    it("parses enums", () => {
      const tree = parseGraphQlSchema(sampleSdl);
      const role = tree.types.find((t) => t.name === "Role");
      expect(role).toBeDefined();
      expect(role!.kind).toBe("enum");
      expect(role!.values).toEqual(["ADMIN", "USER", "GUEST"]);
    });
    it("identifies query and mutation types", () => {
      const tree = parseGraphQlSchema(sampleSdl);
      expect(tree.queryType).toBe("Query");
      expect(tree.mutationType).toBe("Mutation");
    });
    it("handles empty SDL", () => {
      const tree = parseGraphQlSchema("");
      expect(tree.types).toHaveLength(0);
    });
  });

  describe("searchSchemaDocs", () => {
    it("finds types by name", () => {
      const tree = parseGraphQlSchema(sampleSdl);
      const results = searchSchemaDocs(tree, "User");
      expect(results.length).toBeGreaterThan(0);
      expect(results.some((t) => t.name === "User")).toBe(true);
    });
    it("finds types by field name", () => {
      const tree = parseGraphQlSchema(sampleSdl);
      const results = searchSchemaDocs(tree, "email");
      expect(results.length).toBeGreaterThan(0);
    });
    it("returns empty for no match", () => {
      const tree = parseGraphQlSchema(sampleSdl);
      expect(searchSchemaDocs(tree, "zzz-nonexistent")).toEqual([]);
    });
    it("is case-insensitive", () => {
      const tree = parseGraphQlSchema(sampleSdl);
      expect(searchSchemaDocs(tree, "user").length).toBeGreaterThan(0);
    });
  });

  describe("renderDocMarkdown", () => {
    it("renders a type with fields", () => {
      const type: SchemaDocType = {
        name: "User",
        kind: "type",
        fields: [
          { name: "id", type: "ID!", isRequired: true },
          { name: "name", type: "String!", isRequired: true },
        ],
      };
      const md = renderDocMarkdown(type);
      expect(md).toContain("## User");
      expect(md).toContain("id");
      expect(md).toContain("name");
    });
    it("renders an enum", () => {
      const type: SchemaDocType = {
        name: "Role",
        kind: "enum",
        values: ["ADMIN", "USER"],
      };
      const md = renderDocMarkdown(type);
      expect(md).toContain("ADMIN");
      expect(md).toContain("USER");
    });
  });
});
