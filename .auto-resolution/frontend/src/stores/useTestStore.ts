import { create } from "zustand";
import {
  fallbackTestAdapter,
  type TestAdapter,
  type TestFileContent,
  type TestFileRef,
  type TestRunOutcome,
} from "../lib/test";

interface TestTabState {
  path: string;
  title: string;
  content: string;
  format: string;
  /** Baseline version from the last read; save() refreshes it. */
  version: string;
  dirty: boolean;
  outcome: TestRunOutcome | null
  busyAction: "none" | "run" | "save" | "load";
  error: string | null;
}

interface TestState {
  adapter: TestAdapter
  tabs: Record<string, TestTabState>
  tests: TestFileRef[]
  setAdapter(adapter: TestAdapter): void
  openPath(tabId: string, path: string): Promise<void>
  newTab(tabId: string): void
  setContent(tabId: string, content: string): void
  save(tabId: string): Promise<boolean>
  run(tabId: string): Promise<TestRunOutcome | null>
  closeTab(tabId: string): void
  refreshList(): Promise<void>
}

function emptyTab(path = "tests/untitled.reqly-test.yaml", title = "untitled.reqly-test"): TestTabState {
  return {
    path,
    title,
    content: `name: untitled\nrequest:\n  method: GET\n  url: https://api.example.com/\ntests:\n  - name: ok\n    assertions:\n      - kind: status\n        expected: 200\n`,
    format: "yaml",
    version: "",
    dirty: false,
    outcome: null,
    busyAction: "none",
    error: null,
  };
}

export const useTestStore = create<TestState>((set, get) => ({
  adapter: fallbackTestAdapter,
  tabs: {},
  tests: [],

  setAdapter(adapter) {
    set({ adapter });
    void get().refreshList();
  },

  async openPath(tabId, path) {
    const { adapter } = get();
    set((s) => ({
      tabs: {
        ...s.tabs,
        [tabId]: { ...(s.tabs[tabId] ?? emptyTab()), path, busyAction: "load", error: null },
      },
    }));
    try {
      const file: TestFileContent = await adapter.read(path);
      set((s) => ({
        tabs: {
          ...s.tabs,
          [tabId]: {
            ...(s.tabs[tabId] ?? emptyTab()),
            path,
            title: path.split("/").pop() ?? path,
            content: file.content,
            format: file.format,
            version: file.version,
            dirty: false,
            busyAction: "none",
            error: null,
          },
        },
      }));
    } catch (err) {
      set((s) => ({
        tabs: {
          ...s.tabs,
          [tabId]: {
            ...(s.tabs[tabId] ?? emptyTab(path)),
            busyAction: "none",
            error: err instanceof Error ? err.message : String(err),
          },
        },
      }));
    }
  },

  newTab(tabId) {
    set((s) => ({ tabs: { ...s.tabs, [tabId]: emptyTab() } }));
  },

  setContent(tabId, content) {
    set((s) => {
      const tab = s.tabs[tabId];
      if (!tab) return s;
      return { tabs: { ...s.tabs, [tabId]: { ...tab, content, dirty: true } } };
    });
  },

  async save(tabId) {
    const { adapter, tabs } = get();
    const tab = tabs[tabId];
    if (!tab || tab.path === "") return false;
    set((s) => ({ tabs: { ...s.tabs, [tabId]: { ...tab, busyAction: "save", error: null } } }));
    try {
      await adapter.write(tab.path, tab.content);
      set((s) => ({
        tabs: {
          ...s.tabs,
          [tabId]: { ...s.tabs[tabId], version: `${Date.now()}`, dirty: false, busyAction: "none" },
        },
        tests: s.tests.some((t) => t.path === tab.path)
          ? s.tests
          : [
              ...s.tests,
              {
                name: (tab.path.split("/").pop() ?? "").replace(/\.reqly-test\.\w+$/, ""),
                path: tab.path,
              },
            ],
      }));
      void get().refreshList();
      return true;
    } catch (err) {
      set((s) => ({
        tabs: {
          ...s.tabs,
          [tabId]: {
            ...s.tabs[tabId],
            busyAction: "none",
            error: err instanceof Error ? err.message : String(err),
          },
        },
      }));
      return false;
    }
  },

  async run(tabId) {
    const { adapter, tabs } = get();
    const tab = tabs[tabId];
    if (!tab) return null;
    set((s) => ({ tabs: { ...s.tabs, [tabId]: { ...tab, busyAction: "run", error: null } } }));
    try {
      const outcome = await adapter.run({
        path: tab.path,
        content: tab.dirty || tab.path === "" ? tab.content : undefined,
      });
      set((s) => ({
        tabs: { ...s.tabs, [tabId]: { ...s.tabs[tabId], outcome, busyAction: "none" } },
      }));
      return outcome;
    } catch (err) {
      set((s) => ({
        tabs: {
          ...s.tabs,
          [tabId]: {
            ...s.tabs[tabId],
            busyAction: "none",
            error: err instanceof Error ? err.message : String(err),
          },
        },
      }));
      return null;
    }
  },

  closeTab(tabId) {
    set((s) => {
      const tabs = { ...s.tabs };
      delete tabs[tabId];
      return { tabs };
    });
  },

  async refreshList() {
    try {
      const tests = await get().adapter.list();
      set({ tests });
    } catch {
      set({ tests: [] });
    }
  },
}));
