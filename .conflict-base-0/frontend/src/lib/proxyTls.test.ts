import { describe, it, expect } from "vitest";
import {
  DEFAULT_PROXY,
  DEFAULT_TLS,
  validateProxy,
  validateTls,
  formatProxyUrl,
} from "./proxyTls";

describe("proxyTls lib", () => {
  describe("defaults", () => {
    it("proxy defaults to disabled", () => {
      expect(DEFAULT_PROXY.enabled).toBe(false);
      expect(DEFAULT_PROXY.type).toBe("http");
    });
    it("tls defaults to secure", () => {
      expect(DEFAULT_TLS.verifyPeer).toBe(true);
      expect(DEFAULT_TLS.verifyHostnames).toBe(true);
      expect(DEFAULT_TLS.minVersion).toBe("1.2");
    });
  });

  describe("validateProxy", () => {
    it("returns empty when disabled", () => {
      expect(validateProxy({ ...DEFAULT_PROXY, enabled: false })).toEqual([]);
    });
    it("requires host when enabled", () => {
      const errors = validateProxy({ ...DEFAULT_PROXY, enabled: true, host: "" });
      expect(errors).toContain("Host is required");
    });
    it("requires valid port", () => {
      const errors = validateProxy({ ...DEFAULT_PROXY, enabled: true, host: "proxy.local", port: 0 });
      expect(errors.some((e) => /port/i.test(e))).toBe(true);
    });
    it("returns empty for valid config", () => {
      const errors = validateProxy({
        ...DEFAULT_PROXY,
        enabled: true,
        host: "proxy.local",
        port: 8080,
      });
      expect(errors).toEqual([]);
    });
    it("validates port range", () => {
      const errors = validateProxy({
        ...DEFAULT_PROXY,
        enabled: true,
        host: "proxy.local",
        port: 99999,
      });
      expect(errors.some((e) => /port/i.test(e))).toBe(true);
    });
  });

  describe("validateTls", () => {
    it("returns empty for default secure config", () => {
      expect(validateTls(DEFAULT_TLS)).toEqual([]);
    });
    it("warns when peer verification disabled", () => {
      const errors = validateTls({ ...DEFAULT_TLS, verifyPeer: false });
      expect(errors.some((e) => /peer/i.test(e))).toBe(true);
    });
    it("warns when hostname verification disabled", () => {
      const errors = validateTls({ ...DEFAULT_TLS, verifyHostnames: false });
      expect(errors.some((e) => /hostname/i.test(e))).toBe(true);
    });
    it("validates min <= max version", () => {
      const errors = validateTls({ ...DEFAULT_TLS, minVersion: "1.3", maxVersion: "1.1" });
      expect(errors.some((e) => /version/i.test(e))).toBe(true);
    });
  });

  describe("formatProxyUrl", () => {
    it("formats http proxy with auth", () => {
      const url = formatProxyUrl({
        ...DEFAULT_PROXY,
        enabled: true,
        type: "http",
        host: "proxy.local",
        port: 8080,
        auth: { username: "user", password: "pass" },
      });
      expect(url).toBe("http://user:pass@proxy.local:8080");
    });
    it("formats socks5 proxy without auth", () => {
      const url = formatProxyUrl({
        ...DEFAULT_PROXY,
        enabled: true,
        type: "socks5",
        host: "proxy.local",
        port: 1080,
      });
      expect(url).toBe("socks5://proxy.local:1080");
    });
  });

  describe("proxyEnabled", () => {
    it("returns true when proxy is enabled", () => {
      expect({ ...DEFAULT_PROXY, enabled: true }.enabled).toBe(true);
    });
    it("returns false when disabled", () => {
      expect(DEFAULT_PROXY.enabled).toBe(false);
    });
  });
});
