import { describe, expect, it } from "vitest";

import { diagnosticsForSpec, nodesForContent, flattenSpecTree } from "./specTree";

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
