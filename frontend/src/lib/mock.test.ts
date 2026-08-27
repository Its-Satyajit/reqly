import { describe, it, expect } from "vitest";
import {
  matchRoute,
  pruneExpiredState,
  createMockScenario,
} from "./mock";
import type { RequestMatcher, MockStateVariable } from "./mock";

describe("mock lib — extended", () => {
  describe("matchRoute", () => {
    const matchers: RequestMatcher[] = [
      { id: "m1", method: "GET", pathPattern: "/users", priority: 10 },
      { id: "m2", pathPattern: "/admin/**", priority: 1 },
      { id: "m3", method: "POST", pathPattern: "/users", priority: 5 },
    ];

    it("matches method + path", () => {
      const m = matchRoute("GET", "/users", matchers);
      expect(m?.id).toBe("m1");
    });
    it("matches method-less matcher", () => {
      const m = matchRoute("DELETE", "/admin/settings", matchers);
      expect(m?.id).toBe("m2");
    });
    it("returns null on no match", () => {
      expect(matchRoute("GET", "/unknown", matchers)).toBeNull();
    });
    it("prefers lower priority number", () => {
      const m = matchRoute("POST", "/users", matchers);
      expect(m?.id).toBe("m3"); // priority 5 < 10
    });
  });

  describe("pruneExpiredState", () => {
    it("removes expired variables", () => {
      const vars: MockStateVariable[] = [
        { key: "a", value: "1", ttl: 1000, updatedAt: Date.now() - 2000 },
        { key: "b", value: "2", updatedAt: Date.now() },
      ];
      const result = pruneExpiredState(vars);
      expect(result).toHaveLength(1);
      expect(result[0].key).toBe("b");
    });
    it("keeps permanent variables (ttl=0)", () => {
      const vars: MockStateVariable[] = [
        { key: "a", value: "1", ttl: 0, updatedAt: Date.now() - 999999 },
      ];
      expect(pruneExpiredState(vars)).toHaveLength(1);
    });
    it("keeps variables without ttl", () => {
      const vars: MockStateVariable[] = [
        { key: "a", value: "1", updatedAt: Date.now() - 999999 },
      ];
      expect(pruneExpiredState(vars)).toHaveLength(1);
    });
  });

  describe("createMockScenario", () => {
    it("creates scenario with id and name", () => {
      const s = createMockScenario("Success");
      expect(s.id).toBeTruthy();
      expect(s.name).toBe("Success");
      expect(s.routes).toEqual([]);
    });
  });
});
