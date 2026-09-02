import { describe, expect, it } from "vitest";

import { diagnosticsForSpec, nodesForContent, flattenSpecTree, validateEndpoint, patchEndpointInContent } from "./specTree";

describe("nodesForContent", () => {
  it("returns static sections for empty/invalid yaml", () => {
    const nodes = nodesForContent("not: yaml: :");
    expect(nodes.some((n) => n.id === "info")).toBe(true);
    expect(nodes.some((n) => n.id === "paths")).toBe(true);
  });

  it("derives paths children from yaml", () => {
    const yaml = "openapi: 3.1.0\ninfo:\n  title: T\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      summary: List\n  /orders:\n    post:\n      summary: Create\n";
    const nodes = nodesForContent(yaml);
    const paths = nodes.find((n) => n.id === "paths")!;
    expect(paths.children?.map((c) => c.label)).toEqual(["/users", "/orders"]);
  });

  it("handles missing paths", () => {
    const yaml = "openapi: 3.1.0\ninfo:\n  title: T\n  version: 1.0.0\n";
    const nodes = nodesForContent(yaml);
    const paths = nodes.find((n) => n.id === "paths")!;
    expect(paths.children).toEqual([]);
  });
});

describe("diagnosticsForSpec", () => {
  it("reports invalid yaml", () => {
    const d = diagnosticsForSpec("openapi: [unclosed");
    expect(d.length).toBeGreaterThan(0);
    expect(d[0].message).toMatch(/yaml/i);
  });

  it("reports missing openapi field", () => {
    const d = diagnosticsForSpec("info:\n  title: T\n  version: 1.0.0\npaths: {}\n");
    expect(d.some((x) => x.message.includes("openapi"))).toBe(true);
  });

  it("reports missing info.title", () => {
    const d = diagnosticsForSpec("openapi: 3.1.0\ninfo:\n  version: 1.0.0\npaths: {}\n");
    expect(d.some((x) => x.message.includes("title"))).toBe(true);
  });

  it("no diagnostics for valid spec", () => {
    const d = diagnosticsForSpec("openapi: 3.1.0\ninfo:\n  title: T\n  version: 1.0.0\npaths: {}\n");
    expect(d).toEqual([]);
  });

  it("reports missing paths when required fields present", () => {
    // paths is allowed to be empty object, but missing entirely is warning
    const d = diagnosticsForSpec("openapi: 3.1.0\ninfo:\n  title: T\n  version: 1.0.0\n");
    expect(d.some((x) => x.message.includes("paths"))).toBe(true);
  });
});

describe("flattenSpecTree", () => {
  it("flattens nested nodes", () => {
    const nodes = nodesForContent("openapi: 3.1.0\ninfo:\n  title: T\n  version: 1.0.0\npaths:\n  /a:\n    get: {}\n");
    const flat = flattenSpecTree(nodes);
    expect(flat.some((n) => n.id === "paths:/a")).toBe(true);
  });
});

describe("validateEndpoint", () => {
  it("accepts valid GET endpoint", () => {
    expect(validateEndpoint({ path: "/users", method: "GET" })).toEqual([]);
  });

  it("accepts valid POST with summary", () => {
    expect(validateEndpoint({ path: "/users", method: "POST", summary: "Create user" })).toEqual([]);
  });

  it("rejects path missing leading slash", () => {
    const errs = validateEndpoint({ path: "users", method: "GET" });
    expect(errs.some((e) => /path/i.test(e))).toBe(true);
  });

  it("rejects empty path", () => {
    const errs = validateEndpoint({ path: "", method: "GET" });
    expect(errs.length).toBeGreaterThan(0);
  });

  it("rejects path with spaces", () => {
    const errs = validateEndpoint({ path: "/users list", method: "GET" });
    expect(errs.some((e) => /path/i.test(e))).toBe(true);
  });

  it("rejects invalid method", () => {
    const errs = validateEndpoint({ path: "/users", method: "FETCH" });
    expect(errs.some((e) => /method/i.test(e))).toBe(true);
  });

  it("accepts all standard HTTP methods", () => {
    for (const m of ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE"]) {
      expect(validateEndpoint({ path: "/test", method: m })).toEqual([]);
    }
  });

  it("is case-insensitive for method", () => {
    expect(validateEndpoint({ path: "/users", method: "get" })).toEqual([]);
  });

  it("rejects operationId with spaces", () => {
    const errs = validateEndpoint({ path: "/users", method: "GET", operationId: "bad id" });
    expect(errs.some((e) => /operationId/i.test(e))).toBe(true);
  });
});

describe("patchEndpointInContent", () => {
  const base = "openapi: 3.1.0\ninfo:\n  title: T\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      summary: List users\n";

  it("renames path key when path changes", () => {
    const next = patchEndpointInContent(base, "/users", { path: "/customers", method: "GET" });
    expect(next).toContain("/customers:");
    expect(next).not.toContain("/users:");
  });

  it("keeps content unchanged when path and method same", () => {
    const next = patchEndpointInContent(base, "/users", { path: "/users", method: "GET", summary: "List users" });
    // Should still contain original path and not duplicate
    expect(next).toContain("/users:");
    expect((next.match(/\/users:/g) ?? []).length).toBe(1);
  });

  it("inserts method block when method missing", () => {
    const withoutMethod = "openapi: 3.1.0\ninfo:\n  title: T\n  version: 1.0.0\npaths:\n  /users:\n";
    const next = patchEndpointInContent(withoutMethod, "/users", { path: "/users", method: "POST", summary: "Create" });
    expect(next).toContain("post:");
    expect(next).toContain("summary: Create");
  });

  it("updates summary when provided", () => {
    const next = patchEndpointInContent(base, "/users", { path: "/users", method: "GET", summary: "New summary" });
    expect(next).toContain("New summary");
  });
});
