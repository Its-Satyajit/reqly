import { describe, it, expect, beforeEach } from "vitest";
import { useCicdStore } from "./useCicdStore";

describe("useCicdStore", () => {
  beforeEach(() =>
    useCicdStore.setState({
      pipeline: {
        name: "CI Tests",
        environment: "production",
        secrets: [],
      },
    }),
  );

  it("starts with defaults", () => {
    const s = useCicdStore.getState();
    expect(s.pipeline.name).toBe("CI Tests");
    expect(s.pipeline.environment).toBe("production");
  });

  it("setPipeline patches pipeline config", () => {
    useCicdStore.getState().setPipeline({ name: "Nightly" });
    expect(useCicdStore.getState().pipeline.name).toBe("Nightly");
  });

  it("addSecret adds a secret", () => {
    useCicdStore.getState().addSecret("API_KEY");
    expect(useCicdStore.getState().pipeline.secrets).toContain("API_KEY");
  });

  it("removeSecret removes a secret", () => {
    useCicdStore.getState().addSecret("API_KEY");
    useCicdStore.getState().removeSecret("API_KEY");
    expect(useCicdStore.getState().pipeline.secrets).not.toContain("API_KEY");
  });

  it("getCommand generates CLI command", () => {
    useCicdStore.getState().setPipeline({ collectionPath: "./coll.yaml" });
    const cmd = useCicdStore.getState().getCommand();
    expect(cmd).toContain("reqly collection test");
  });
});
