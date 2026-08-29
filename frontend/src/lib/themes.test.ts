import { describe, it, expect } from "vitest";
import {
  THEMES,
  isThemeId,
  resolveTheme,
  appearanceFor,
  nextPreference,
  normalizeStored,
  applyThemeToDocument,
  parseCustomTheme,
  validateCustomTheme,
  customThemeToCSS,
  type CustomTheme,
} from "./themes";

describe("themes lib", () => {
  describe("built-ins", () => {
    it("has two atlas themes", () => {
      expect(THEMES).toHaveLength(2);
      expect(THEMES.map((t) => t.id)).toContain("atlas-light");
      expect(THEMES.map((t) => t.id)).toContain("atlas-dark");
    });
    it("validates theme ids", () => {
      expect(isThemeId("atlas-light")).toBe(true);
      expect(isThemeId("custom")).toBe(false);
    });
    it("resolves system preference", () => {
      expect(resolveTheme("system", true)).toBe("atlas-dark");
      expect(resolveTheme("system", false)).toBe("atlas-light");
      expect(resolveTheme("atlas-light", true)).toBe("atlas-light");
    });
    it("returns appearance", () => {
      expect(appearanceFor("atlas-light")).toBe("light");
      expect(appearanceFor("atlas-dark")).toBe("dark");
    });
    it("cycles preferences", () => {
      expect(nextPreference("atlas-light")).toBe("atlas-dark");
      expect(nextPreference("atlas-dark")).toBe("system");
      expect(nextPreference("system")).toBe("atlas-light");
    });
    it("normalizes stored values", () => {
      expect(normalizeStored(null)).toBe(null);
      expect(normalizeStored("atlas-light")).toBe("atlas-light");
      expect(normalizeStored("light")).toBe("atlas-light");
      expect(normalizeStored("dark")).toBe("atlas-dark");
      expect(normalizeStored("invalid")).toBe(null);
    });
    it("applies theme to document", () => {
      applyThemeToDocument("atlas-light");
      expect(document.documentElement.getAttribute("data-theme")).toBe("atlas-light");
      applyThemeToDocument("atlas-dark");
      expect(document.documentElement.getAttribute("data-theme")).toBe("atlas-dark");
    });
  });

  describe("custom theme sharing (M67)", () => {
    const valid: CustomTheme = {
      id: "my-theme",
      label: "My Theme",
      appearance: "light",
      tokens: { primary: "#ff0000", background: "#ffffff" },
    };

    it("validates a correct theme", () => {
      expect(validateCustomTheme(valid)).toBeNull();
    });
    it("rejects missing id", () => {
      expect(validateCustomTheme({ ...valid, id: "" })).toMatch(/id/);
    });
    it("rejects bad id (kebab-case)", () => {
      expect(validateCustomTheme({ ...valid, id: "Bad_ID" })).toMatch(/kebab/);
    });
    it("rejects missing label", () => {
      expect(validateCustomTheme({ ...valid, label: "" })).toMatch(/label/);
    });
    it("rejects bad appearance", () => {
      // SAFETY: testing invalid appearance string by casting literal
      const bad = "blue" as CustomTheme["appearance"];
      expect(validateCustomTheme({ ...valid, appearance: bad })).toMatch(/appearance/);
    });
    it("parses YAML and JSON", () => {
      const yaml = `id: yaml-theme\nlabel: YAML\nappearance: dark\ntokens:\n  primary: "#123456"\n`;
      const parsedYaml = parseCustomTheme(yaml);
      expect(parsedYaml.id).toBe("yaml-theme");
      const json = `{"id":"json-theme","label":"JSON","appearance":"light","tokens":{"primary":"#000000"}}`;
      const parsedJson = parseCustomTheme(json);
      expect(parsedJson.id).toBe("json-theme");
    });
    it("throws on invalid parse", () => {
      expect(() => parseCustomTheme("not: [valid")).toThrow();
      expect(() => parseCustomTheme(`id: bad\nlabel: ""\nappearance: light`)).toThrow();
    });
    it("generates CSS", () => {
      const css = customThemeToCSS(valid);
      expect(css).toContain(`[data-theme="my-theme"]`);
      expect(css).toContain("--primary: #ff0000");
      expect(css).toContain("--background: #ffffff");
    });
  });
});
