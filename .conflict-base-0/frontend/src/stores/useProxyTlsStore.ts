import { create } from "zustand";
import {
  DEFAULT_PROXY,
  DEFAULT_TLS,
  validateProxy,
  validateTls,
  type ProxyConfig,
  type TlsConfig,
} from "#lib/proxyTls";

interface ProxyTlsState {
  proxy: ProxyConfig;
  tls: TlsConfig;
  setProxy(patch: Partial<ProxyConfig>): void;
  setTls(patch: Partial<TlsConfig>): void;
  resetProxy(): void;
  resetTls(): void;
  validateProxy(): string[];
  validateTls(): string[];
}

export const useProxyTlsStore = create<ProxyTlsState>((set, get) => ({
  proxy: { ...DEFAULT_PROXY },
  tls: { ...DEFAULT_TLS },

  setProxy(patch) {
    set((s) => ({ proxy: { ...s.proxy, ...patch } }));
  },

  setTls(patch) {
    set((s) => ({ tls: { ...s.tls, ...patch } }));
  },

  resetProxy() {
    set({ proxy: { ...DEFAULT_PROXY } });
  },

  resetTls() {
    set({ tls: { ...DEFAULT_TLS } });
  },

  validateProxy() {
    return validateProxy(get().proxy);
  },

  validateTls() {
    return validateTls(get().tls);
  },
}));
