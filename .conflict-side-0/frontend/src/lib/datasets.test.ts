import { describe, it, expect } from "vitest";
import {
  parseCsv,
  parseJsonDataset,
  getRowVariables,
  validateDataset,
  type Dataset,
} from "./datasets";

describe("datasets lib", () => {
  describe("parseCsv", () => {
    it("parses simple CSV", () => {
      const ds = parseCsv("id,name\n1,Alice\n2,Bob");
      expect(ds.columns).toHaveLength(2);
      expect(ds.rows).toHaveLength(2);
      expect(ds.rows[0].values).toEqual({ id: "1", name: "Alice" });
      expect(ds.rows[1].values).toEqual({ id: "2", name: "Bob" });
    });
    it("handles quoted fields with commas", () => {
      const ds = parseCsv('id,name\n1,"Smith, John"');
      expect(ds.rows[0].values.name).toBe("Smith, John");
    });
    it("handles quoted fields with newlines", () => {
      const ds = parseCsv('id,name\n1,"Line1\nLine2"');
      expect(ds.rows[0].values.name).toBe("Line1\nLine2");
    });
    it("handles empty CSV", () => {
      const ds = parseCsv("");
      expect(ds.columns).toHaveLength(0);
      expect(ds.rows).toHaveLength(0);
    });
    it("handles header-only CSV", () => {
      const ds = parseCsv("id,name");
      expect(ds.columns).toHaveLength(2);
      expect(ds.rows).toHaveLength(0);
    });
    it("trims whitespace in headers", () => {
      const ds = parseCsv(" id , name \n1,Alice");
      expect(ds.columns[0].name).toBe("id");
      expect(ds.columns[1].name).toBe("name");
    });
    it("sets source to csv", () => {
      const ds = parseCsv("id\n1");
      expect(ds.source).toBe("csv");
    });
    it("handles unterminated quotes gracefully (lenient parsing)", () => {
      const ds = parseCsv('id,name\n1,"Smith');
      expect(ds.rows[0].values.name).toBe("Smith");
    });
  });

  describe("parseJsonDataset", () => {
    it("parses array of objects", () => {
      const ds = parseJsonDataset('[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]');
      expect(ds.columns).toHaveLength(2);
      expect(ds.rows).toHaveLength(2);
      expect(ds.rows[0].values).toEqual({ id: "1", name: "Alice" });
    });
    it("handles empty array", () => {
      const ds = parseJsonDataset("[]");
      expect(ds.columns).toHaveLength(0);
      expect(ds.rows).toHaveLength(0);
    });
    it("sets source to json", () => {
      const ds = parseJsonDataset('[{"id":1}]');
      expect(ds.source).toBe("json");
    });
    it("throws on invalid JSON", () => {
      expect(() => parseJsonDataset("not json")).toThrow();
    });
    it("throws on non-array JSON", () => {
      expect(() => parseJsonDataset('{"id":1}')).toThrow();
    });
  });

  describe("getRowVariables", () => {
    it("returns all row values as variables", () => {
      const ds: Dataset = {
        name: "test",
        columns: [
          { name: "id", index: 0 },
          { name: "name", index: 1 },
        ],
        rows: [{ index: 0, values: { id: "1", name: "Alice" } }],
        source: "csv",
        rawContent: "",
      };
      const vars = getRowVariables(ds, 0);
      expect(vars).toEqual({ id: "1", name: "Alice" });
    });
    it("returns empty for missing row", () => {
      const ds: Dataset = {
        name: "test",
        columns: [],
        rows: [],
        source: "csv",
        rawContent: "",
      };
      expect(getRowVariables(ds, 0)).toEqual({});
    });
  });

  describe("validateDataset", () => {
    it("returns empty for valid dataset", () => {
      const ds = parseCsv("id,name\n1,Alice");
      expect(validateDataset(ds)).toEqual([]);
    });
    it("warns on empty rows", () => {
      const ds = parseCsv("id,name");
      expect(validateDataset(ds).length).toBeGreaterThan(0);
    });
    it("warns on duplicate column names", () => {
      const ds = parseCsv("id,id\n1,2");
      expect(validateDataset(ds).some((e) => /duplicate/i.test(e))).toBe(true);
    });
  });
});
