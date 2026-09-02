import { describe, it, expect, beforeEach } from "vitest";
import { useMockStore } from "./useMockStore";

describe("useMockStore — §56.7 extensions", () => {
  beforeEach(() =>
    useMockStore.setState({
      scenarios: [],
      activeScenarioId: null,
      stateVariables: [],
      faultInjection: { enabled: false, type: "delay", probability: 0 },
      requestMatchers: [],
      logs: [],
    }),
  );

  describe("scenarios", () => {
    it("creates a scenario", () => {
      useMockStore.getState().createScenario("Success");
      expect(useMockStore.getState().scenarios).toHaveLength(1);
      expect(useMockStore.getState().scenarios[0].name).toBe("Success");
    });
    it("deletes a scenario", () => {
      useMockStore.getState().createScenario("Test");
      const id = useMockStore.getState().scenarios[0].id;
      useMockStore.getState().deleteScenario(id);
      expect(useMockStore.getState().scenarios).toHaveLength(0);
    });
    it("sets active scenario", () => {
      useMockStore.getState().createScenario("Test");
      const id = useMockStore.getState().scenarios[0].id;
      useMockStore.getState().setActiveScenario(id);
      expect(useMockStore.getState().activeScenarioId).toBe(id);
    });
    it("updates scenario", () => {
      useMockStore.getState().createScenario("Old");
      const id = useMockStore.getState().scenarios[0].id;
      useMockStore.getState().updateScenario(id, { name: "New" });
      expect(useMockStore.getState().scenarios[0].name).toBe("New");
    });
  });

  describe("state variables", () => {
    it("adds a state variable", () => {
      useMockStore.getState().setMockStateVariable("count", "0");
      expect(useMockStore.getState().stateVariables).toHaveLength(1);
      expect(useMockStore.getState().stateVariables[0].key).toBe("count");
    });
    it("updates existing variable", () => {
      useMockStore.getState().setMockStateVariable("count", "0");
      useMockStore.getState().setMockStateVariable("count", "1");
      expect(useMockStore.getState().stateVariables).toHaveLength(1);
      expect(useMockStore.getState().stateVariables[0].value).toBe("1");
    });
    it("clears a variable", () => {
      useMockStore.getState().setMockStateVariable("x", "1");
      useMockStore.getState().clearMockStateVariable("x");
      expect(useMockStore.getState().stateVariables).toHaveLength(0);
    });
  });

  describe("fault injection", () => {
    it("sets fault injection", () => {
      useMockStore.getState().setFaultInjection({ enabled: true, type: "drop", probability: 0.5 });
      const fi = useMockStore.getState().faultInjection;
      expect(fi.enabled).toBe(true);
      expect(fi.type).toBe("drop");
      expect(fi.probability).toBe(0.5);
    });
  });

  describe("request matchers", () => {
    it("adds a matcher", () => {
      useMockStore.getState().addRequestMatcher({ pathPattern: "/api/**", priority: 1 });
      expect(useMockStore.getState().requestMatchers).toHaveLength(1);
    });
    it("removes a matcher", () => {
      useMockStore.getState().addRequestMatcher({ pathPattern: "/a", priority: 1 });
      const id = useMockStore.getState().requestMatchers[0].id;
      useMockStore.getState().removeRequestMatcher(id);
      expect(useMockStore.getState().requestMatchers).toHaveLength(0);
    });
  });

  describe("logs", () => {
    it("adds a log entry", () => {
      useMockStore.getState().addMockLog({
        timestamp: Date.now(),
        method: "GET",
        path: "/test",
        status: 200,
        duration: 42,
      });
      expect(useMockStore.getState().logs).toHaveLength(1);
    });
    it("clears logs", () => {
      useMockStore.getState().addMockLog({
        timestamp: Date.now(),
        method: "GET",
        path: "/test",
        status: 200,
        duration: 42,
      });
      useMockStore.getState().clearMockLogs();
      expect(useMockStore.getState().logs).toHaveLength(0);
    });
  });
});
