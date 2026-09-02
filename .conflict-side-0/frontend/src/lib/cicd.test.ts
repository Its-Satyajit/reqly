import { describe, it, expect } from "vitest";
import {
  generateCliCommand,
  generateGitHubAction,
  parseTestReport,
  type CicdPipeline,
} from "./cicd";

describe("cicd lib", () => {
  const basePipeline: CicdPipeline = {
    name: "API Tests",
    environment: "production",
    secrets: ["API_KEY"],
    collectionPath: "./collections/api.yaml",
  };

  describe("generateCliCommand", () => {
    it("generates basic command", () => {
      const cmd = generateCliCommand(basePipeline);
      expect(cmd).toContain("reqly collection test");
      expect(cmd).toContain("--env production");
    });
    it("includes collection path", () => {
      const cmd = generateCliCommand(basePipeline);
      expect(cmd).toContain("./collections/api.yaml");
    });
    it("omits collection path when undefined", () => {
      const cmd = generateCliCommand({ ...basePipeline, collectionPath: undefined });
      expect(cmd).not.toContain("--collection");
    });
  });

  describe("generateGitHubAction", () => {
    it("generates valid YAML", () => {
      const yaml = generateGitHubAction(basePipeline);
      expect(yaml).toContain("name:");
      expect(yaml).toContain("API Tests");
      expect(yaml).toContain("reqly collection test");
    });
    it("includes secret env vars", () => {
      const yaml = generateGitHubAction(basePipeline);
      expect(yaml).toContain("API_KEY");
    });
  });

  describe("parseTestReport", () => {
    it("parses JUnit-style report", () => {
      const report = "4 passed, 2 failed, 1 skipped";
      const result = parseTestReport(report);
      expect(result.passed).toBe(4);
      expect(result.failed).toBe(2);
      expect(result.skipped).toBe(1);
    });
    it("handles zero failures", () => {
      const result = parseTestReport("10 passed, 0 failed");
      expect(result.passed).toBe(10);
      expect(result.failed).toBe(0);
    });
    it("returns zeros for empty/unparseable", () => {
      const result = parseTestReport("no data here");
      expect(result.passed).toBe(0);
      expect(result.failed).toBe(0);
    });
  });
});
