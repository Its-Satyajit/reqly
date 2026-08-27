import { describe, it, expect, beforeEach, vi } from "vitest";
import { STORAGE_KEY } from "#lib/themes";

function mockMatchMedia(dark: boolean) {
  const mql = {
    matches: dark,
    media: "(prefers-color-scheme: dark)",
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
  } as unknown as MediaQueryList;
  Object.defineProperty(window, "matchMedia", {
    writable: true,
    value: vi.fn().mockReturnValue(mql),
  });
  return mql as MediaQueryList & { addEventListener: ReturnType<typeof vi.fn> };
}

describe("useThemeStore", () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute("data-theme");
    document.documentElement.classList.remove("dark");
    vi.resetModules();
  });

  it("defaults to atlas-dark when no stored value and OS is dark", async () => {
    mockMatchMedia(true);
    const { useThemeStore } = await import("./useThemeStore");
    expect(useThemeStore.getState().theme).toBe("atlas-dark");
    expect(useThemeStore.getState().resolvedTheme).toBe("atlas-dark");
  });

  it("resolves system to atlas-light when OS is light", async () => {
    mockMatchMedia(false);
    localStorage.setItem(STORAGE_KEY, "system");
    const { useThemeStore } = await import("./useThemeStore");
    expect(useThemeStore.getState().theme).toBe("system");
    expect(useThemeStore.getState().resolvedTheme).toBe("atlas-light");
  });

  it("migrates legacy light/dark values", async () => {
    mockMatchMedia(true);
    localStorage.setItem(STORAGE_KEY, "light");
    const { useThemeStore } = await import("./useThemeStore");
    expect(useThemeStore.getState().theme).toBe("atlas-light");
  });

  it("persists setTheme and updates resolvedTheme", async () => {
    mockMatchMedia(true);
    const { useThemeStore } = await import("./useThemeStore");
    useThemeStore.getState().setTheme("atlas-light");
    expect(localStorage.getItem(STORAGE_KEY)).toBe("atlas-light");
    expect(useThemeStore.getState().resolvedTheme).toBe("atlas-light");
    expect(document.documentElement.getAttribute("data-theme")).toBe("atlas-light");
  });

  it("cycles light -> dark -> system -> light", async () => {
    mockMatchMedia(false);
    const { useThemeStore } = await import("./useThemeStore");
    useThemeStore.getState().setTheme("atlas-light");
    useThemeStore.getState().cycleTheme();
    expect(useThemeStore.getState().theme).toBe("atlas-dark");
    useThemeStore.getState().cycleTheme();
    expect(useThemeStore.getState().theme).toBe("system");
    useThemeStore.getState().cycleTheme();
    expect(useThemeStore.getState().theme).toBe("atlas-light");
  });
});
