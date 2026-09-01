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
    it("has 14 built-in themes", () => {
      expect(THEMES).toHaveLength(14);
      expect(THEMES.map((t) => t.id)).toContain("atlas-light");
      expect(THEMES.map((t) => t.id)).toContain("atlas-dark");
      expect(THEMES.map((t) => t.id)).toContain("windows-11-light");
      expect(THEMES.map((t) => t.id)).toContain("windows-11-dark");
      expect(THEMES.map((t) => t.id)).toContain("windows-11");
      expect(THEMES.map((t) => t.id)).toContain("macos-tahoe-light");
      expect(THEMES.map((t) => t.id)).toContain("macos-tahoe-dark");
      expect(THEMES.map((t) => t.id)).toContain("macos-tahoe");
      expect(THEMES.map((t) => t.id)).toContain("linux-kde-light");
      expect(THEMES.map((t) => t.id)).toContain("linux-kde-dark");
      expect(THEMES.map((t) => t.id)).toContain("linux-kde");
      expect(THEMES.map((t) => t.id)).toContain("linux-gnome-light");
      expect(THEMES.map((t) => t.id)).toContain("linux-gnome-dark");
      expect(THEMES.map((t) => t.id)).toContain("linux-gnome");
    });
    it("validates theme ids", () => {
      expect(isThemeId("atlas-light")).toBe(true);
      expect(isThemeId("windows-11-light")).toBe(true);
      expect(isThemeId("windows-11-dark")).toBe(true);
      expect(isThemeId("windows-11")).toBe(true);
      expect(isThemeId("macos-tahoe-light")).toBe(true);
      expect(isThemeId("macos-tahoe-dark")).toBe(true);
      expect(isThemeId("macos-tahoe")).toBe(true);
      expect(isThemeId("linux-kde-light")).toBe(true);
      expect(isThemeId("linux-kde-dark")).toBe(true);
      expect(isThemeId("linux-kde")).toBe(true);
      expect(isThemeId("linux-gnome-light")).toBe(true);
      expect(isThemeId("linux-gnome-dark")).toBe(true);
      expect(isThemeId("linux-gnome")).toBe(true);
      expect(isThemeId("custom")).toBe(false);
    });
    it("resolves system preference", () => {
      expect(resolveTheme("system", true)).toBe("atlas-dark");
      expect(resolveTheme("system", false)).toBe("atlas-light");
      expect(resolveTheme("atlas-light", true)).toBe("atlas-light");
      expect(resolveTheme("windows-11-light", false)).toBe("windows-11-light");
      expect(resolveTheme("macos-tahoe-light", false)).toBe("macos-tahoe-light");
      expect(resolveTheme("linux-kde-light", false)).toBe("linux-kde-light");
      expect(resolveTheme("linux-gnome-light", false)).toBe("linux-gnome-light");
      expect(resolveTheme("linux-gnome-dark", false)).toBe("linux-gnome-dark");
    });
    it("returns appearance", () => {
      expect(appearanceFor("atlas-light")).toBe("light");
      expect(appearanceFor("atlas-dark")).toBe("dark");
      expect(appearanceFor("windows-11-light")).toBe("light");
      expect(appearanceFor("windows-11-dark")).toBe("dark");
      expect(appearanceFor("windows-11")).toBe("dark");
      expect(appearanceFor("macos-tahoe-light")).toBe("light");
      expect(appearanceFor("macos-tahoe-dark")).toBe("dark");
      expect(appearanceFor("macos-tahoe")).toBe("dark");
      expect(appearanceFor("linux-kde-light")).toBe("light");
      expect(appearanceFor("linux-kde-dark")).toBe("dark");
      expect(appearanceFor("linux-kde")).toBe("dark");
      expect(appearanceFor("linux-gnome-light")).toBe("light");
      expect(appearanceFor("linux-gnome-dark")).toBe("dark");
      expect(appearanceFor("linux-gnome")).toBe("dark");
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
