import { describe, it, expect } from "vitest";
import {
  CATEGORIES,
  getCategoryList,
  getTemplatesByCategory,
  getTemplateById,
  instantiateTemplate,
  searchTemplates,
  type RequestTemplate,
} from "./templates";

describe("templates lib", () => {
  describe("getCategoryList", () => {
    it("returns all four categories", () => {
      const cats = getCategoryList();
      expect(cats).toHaveLength(4);
      expect(cats.map((c) => c.id)).toEqual(["rest", "graphql", "grpc", "realtime"]);
    });
  });

  describe("getTemplatesByCategory", () => {
    it("returns templates for rest category", () => {
      const templates = getTemplatesByCategory("rest");
      expect(templates.length).toBeGreaterThan(0);
      expect(templates.every((t) => t.category === "rest")).toBe(true);
    });
    it("returns templates for graphql category", () => {
      const templates = getTemplatesByCategory("graphql");
      expect(templates.length).toBeGreaterThan(0);
    });
    it("returns templates for grpc category", () => {
      const templates = getTemplatesByCategory("grpc");
      expect(templates.length).toBeGreaterThan(0);
    });
    it("returns templates for realtime category", () => {
      const templates = getTemplatesByCategory("realtime");
      expect(templates.length).toBeGreaterThan(0);
    });
    it("returns empty array for unknown category", () => {
      // SAFETY: deliberate invalid input to test guard clause
      expect(getTemplatesByCategory("unknown" as never)).toEqual([]);
    });
  });

  describe("getTemplateById", () => {
    it("finds a template by id", () => {
      const first = CATEGORIES[0].templates[0];
      expect(getTemplateById(first.id)).toBe(first);
    });
    it("returns undefined for unknown id", () => {
      expect(getTemplateById("nonexistent")).toBeUndefined();
    });
  });

  describe("instantiateTemplate", () => {
    it("creates a request object from a REST template", () => {
      const template: RequestTemplate = {
        id: "test-get-list",
        category: "rest",
        subcategory: "CRUD",
        name: "GET /items",
        description: "List all items",
        method: "GET",
        path: "/items",
        headers: { Accept: "application/json" },
      };
      const req = instantiateTemplate(template);
      expect(req.method).toBe("GET");
      expect(req.path).toBe("/items");
      expect(req.headers).toEqual({ Accept: "application/json" });
      expect(req.body).toBeUndefined();
    });
    it("creates a request object from a template with body", () => {
      const template: RequestTemplate = {
        id: "test-post",
        category: "rest",
        name: "POST /items",
        description: "Create item",
        method: "POST",
        path: "/items",
        body: '{"name":"test"}',
        contentType: "application/json",
      };
      const req = instantiateTemplate(template);
      expect(req.method).toBe("POST");
      expect(req.body).toBe('{"name":"test"}');
      expect(req.headers?.["Content-Type"]).toBe("application/json");
    });
    it("creates a request object from a GraphQL template", () => {
      const template: RequestTemplate = {
        id: "test-gql-query",
        category: "graphql",
        name: "Query",
        description: "GraphQL query",
        method: "POST",
        path: "/graphql",
        body: "{ __typename }",
        contentType: "application/json",
      };
      const req = instantiateTemplate(template);
      expect(req.method).toBe("POST");
      expect(req.body).toBe("{ __typename }");
    });
  });

  describe("searchTemplates", () => {
    it("finds templates by name substring", () => {
      const results = searchTemplates("CRUD");
      expect(results.length).toBeGreaterThan(0);
      expect(results.some((t) => t.subcategory === "CRUD")).toBe(true);
    });
    it("finds templates by description", () => {
      const results = searchTemplates("WebSocket");
      expect(results.length).toBeGreaterThan(0);
    });
    it("returns empty for no match", () => {
      expect(searchTemplates("zzz-nonexistent")).toEqual([]);
    });
    it("is case-insensitive", () => {
      const results = searchTemplates("crud");
      expect(results.length).toBeGreaterThan(0);
    });
  });
});
