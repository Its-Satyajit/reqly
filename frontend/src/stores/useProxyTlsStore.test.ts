import { describe, it, expect, beforeEach } from "vitest";
import { useProxyTlsStore } from "./useProxyTlsStore";
import { DEFAULT_PROXY, DEFAULT_TLS } from "#lib/proxyTls";

describe("useProxyTlsStore", () => {
  beforeEach(() =>
    useProxyTlsStore.setState({
      proxy: { ...DEFAULT_PROXY },
      tls: { ...DEFAULT_TLS },
    }),
  );

  it("starts with defaults", () => {
    const s = useProxyTlsStore.getState();
    expect(s.proxy.enabled).toBe(false);
    expect(s.tls.verifyPeer).toBe(true);
  });

  it("setProxy patches proxy config", () => {
    useProxyTlsStore.getState().setProxy({ enabled: true, host: "proxy.local" });
    const s = useProxyTlsStore.getState();
    expect(s.proxy.enabled).toBe(true);
    expect(s.proxy.host).toBe("proxy.local");
    expect(s.proxy.port).toBe(8080); // unchanged default
  });

  it("setTls patches tls config", () => {
    useProxyTlsStore.getState().setTls({ verifyPeer: false });
    expect(useProxyTlsStore.getState().tls.verifyPeer).toBe(false);
  });

  it("resetProxy restores defaults", () => {
    useProxyTlsStore.getState().setProxy({ enabled: true, host: "x" });
    useProxyTlsStore.getState().resetProxy();
    expect(useProxyTlsStore.getState().proxy).toEqual(DEFAULT_PROXY);
  });

  it("resetTls restores defaults", () => {
    useProxyTlsStore.getState().setTls({ verifyPeer: false });
    useProxyTlsStore.getState().resetTls();
    expect(useProxyTlsStore.getState().tls).toEqual(DEFAULT_TLS);
  });

  it("validateProxy returns errors for invalid config", () => {
    useProxyTlsStore.getState().setProxy({ enabled: true, host: "" });
    const errors = useProxyTlsStore.getState().validateProxy();
    expect(errors.length).toBeGreaterThan(0);
  });

  it("validateTls returns errors for insecure config", () => {
    useProxyTlsStore.getState().setTls({ verifyPeer: false, verifyHostnames: false });
    const errors = useProxyTlsStore.getState().validateTls();
    expect(errors.length).toBeGreaterThan(0);
  });
});
