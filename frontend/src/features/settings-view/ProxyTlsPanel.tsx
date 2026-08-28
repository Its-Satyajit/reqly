import { useProxyTlsStore } from "#stores/useProxyTlsStore";
import { isProxyType, isTlsVersion } from "#lib/proxyTls";
import { Input } from "#components/ui/input";
import { CompactSelect } from "#components/CompactSelect";
import { Alert, AlertDescription } from "#components/ui/alert";

const PROXY_TYPE_OPTIONS = [
  { value: "http", label: "HTTP" },
  { value: "https", label: "HTTPS" },
  { value: "socks5", label: "SOCKS5" },
];

const TLS_VERSION_OPTIONS = [
  { value: "1.0", label: "TLS 1.0" },
  { value: "1.1", label: "TLS 1.1" },
  { value: "1.2", label: "TLS 1.2 (Default)" },
  { value: "1.3", label: "TLS 1.3" },
];

export function ProxyTlsPanel() {
  const proxy = useProxyTlsStore((s) => s.proxy);
  const tls = useProxyTlsStore((s) => s.tls);
  const setProxy = useProxyTlsStore((s) => s.setProxy);
  const setTls = useProxyTlsStore((s) => s.setTls);
  const resetProxy = useProxyTlsStore((s) => s.resetProxy);
  const resetTls = useProxyTlsStore((s) => s.resetTls);

  const proxyErrors = useProxyTlsStore((s) => s.validateProxy)();
  const tlsErrors = useProxyTlsStore((s) => s.validateTls)();

  return (
    <section className="rounded-lg border border-border bg-card p-4 space-y-4">
      <div>
        <h2 className="text-sm font-semibold">Network & Security (Proxy / TLS)</h2>
        <p className="mt-1 text-xs text-muted-foreground">
          Configure global proxy routes and TLS / certificate validation behaviors.
        </p>
      </div>

      {/* Proxy Config */}
      <div className="rounded-md border border-border/80 bg-background/50 p-3 space-y-3">
        <div className="flex items-center justify-between">
          <label className="flex items-center gap-2 text-xs font-medium cursor-pointer">
            <input
              type="checkbox"
              checked={proxy.enabled}
              onChange={(e) => setProxy({ enabled: e.target.checked })}
              className="size-3.5 rounded border-border text-primary focus:ring-primary"
            />
            Enable Proxy
          </label>
          {proxy.enabled && (
            <button
              type="button"
              onClick={resetProxy}
              className="text-[11px] text-muted-foreground hover:text-foreground underline"
            >
              Reset Proxy
            </button>
          )}
        </div>

        {proxy.enabled && (
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 pt-1">
            <div className="flex flex-col gap-1">
              <span className="text-[11px] text-muted-foreground">Type</span>
              <CompactSelect
                value={proxy.type}
                onChange={(v) => {
                  if (isProxyType(v)) {
                    setProxy({ type: v });
                  }
                }}
                options={PROXY_TYPE_OPTIONS}
                ariaLabel="Proxy type"
              />
            </div>
            <div className="flex flex-col gap-1 sm:col-span-2">
              <span className="text-[11px] text-muted-foreground">Host</span>
              <Input
                value={proxy.host}
                onChange={(e) => setProxy({ host: e.target.value })}
                placeholder="127.0.0.1 or proxy.example.com"
                className="h-8 font-mono text-xs"
              />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-[11px] text-muted-foreground">Port</span>
              <Input
                type="number"
                value={proxy.port === 0 ? "" : proxy.port}
                onChange={(e) => {
                  const val = e.target.value === "" ? 0 : Number(e.target.value);
                  setProxy({ port: Number.isNaN(val) ? 8080 : val });
                }}
                placeholder="8080"
                className="h-8 font-mono text-xs"
              />
            </div>
            <div className="flex flex-col gap-1 sm:col-span-2">
              <span className="text-[11px] text-muted-foreground">Bypass List (comma separated)</span>
              <Input
                value={proxy.bypassList.join(", ")}
                onChange={(e) =>
                  setProxy({
                    bypassList: e.target.value
                      .split(",")
                      .map((s) => s.trim())
                      .filter(Boolean),
                  })
                }
                placeholder="localhost, 127.0.0.1, *.internal"
                className="h-8 font-mono text-xs"
              />
            </div>
            <div className="flex flex-col gap-1 sm:col-span-3 pt-1 border-t border-border/40">
              <span className="text-[11px] font-medium text-muted-foreground">Proxy Authentication (optional)</span>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mt-1">
                <Input
                  value={proxy.auth?.username ?? ""}
                  onChange={(e) =>
                    setProxy({
                      auth: e.target.value || proxy.auth?.password
                        ? { username: e.target.value, password: proxy.auth?.password ?? "" }
                        : undefined,
                    })
                  }
                  placeholder="Username"
                  className="h-8 font-mono text-xs"
                />
                <Input
                  type="password"
                  value={proxy.auth?.password ?? ""}
                  onChange={(e) =>
                    setProxy({
                      auth: e.target.value || proxy.auth?.username
                        ? { username: proxy.auth?.username ?? "", password: e.target.value }
                        : undefined,
                    })
                  }
                  placeholder="Password"
                  className="h-8 font-mono text-xs"
                />
              </div>
            </div>
          </div>
        )}

        {proxyErrors.length > 0 && (
          <Alert variant="destructive" className="py-2">
            <AlertDescription className="text-xs">
              {proxyErrors.join(", ")}
            </AlertDescription>
          </Alert>
        )}
      </div>

      {/* TLS Config */}
      <div className="rounded-md border border-border/80 bg-background/50 p-3 space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium">TLS & SSL Verification</span>
          <button
            type="button"
            onClick={resetTls}
            className="text-[11px] text-muted-foreground hover:text-foreground underline"
          >
            Reset TLS
          </button>
        </div>

        <div className="space-y-2">
          <label className="flex items-center gap-2 text-xs cursor-pointer">
            <input
              type="checkbox"
              checked={tls.verifyPeer}
              onChange={(e) => setTls({ verifyPeer: e.target.checked })}
              className="size-3.5 rounded border-border text-primary focus:ring-primary"
            />
            Verify Server Certificate (Peer Verification)
          </label>

          <label className="flex items-center gap-2 text-xs cursor-pointer">
            <input
              type="checkbox"
              checked={tls.verifyHostnames}
              onChange={(e) => setTls({ verifyHostnames: e.target.checked })}
              className="size-3.5 rounded border-border text-primary focus:ring-primary"
            />
            Verify Hostnames
          </label>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-1">
          <div className="flex flex-col gap-1">
            <span className="text-[11px] text-muted-foreground">Min TLS Version</span>
            <CompactSelect
              value={tls.minVersion}
              onChange={(v) => {
                if (isTlsVersion(v)) {
                  setTls({ minVersion: v });
                }
              }}
              options={TLS_VERSION_OPTIONS}
              ariaLabel="Minimum TLS Version"
            />
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-[11px] text-muted-foreground">Custom CA Path (optional)</span>
            <Input
              value={tls.customCaPath ?? ""}
              onChange={(e) => setTls({ customCaPath: e.target.value.trim() || undefined })}
              placeholder="/path/to/ca.pem"
              className="h-8 font-mono text-xs"
            />
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-[11px] text-muted-foreground">Client Certificate Path (mTLS)</span>
            <Input
              value={tls.clientCertPath ?? ""}
              onChange={(e) => setTls({ clientCertPath: e.target.value.trim() || undefined })}
              placeholder="/path/to/client-cert.pem"
              className="h-8 font-mono text-xs"
            />
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-[11px] text-muted-foreground">Client Key Path (mTLS)</span>
            <Input
              value={tls.clientKeyPath ?? ""}
              onChange={(e) => setTls({ clientKeyPath: e.target.value.trim() || undefined })}
              placeholder="/path/to/client-key.pem"
              className="h-8 font-mono text-xs"
            />
          </div>
        </div>

        {tlsErrors.length > 0 && (
          <Alert variant="destructive" className="py-2">
            <AlertDescription className="text-xs">
              {tlsErrors.join(", ")}
            </AlertDescription>
          </Alert>
        )}
      </div>
    </section>
  );
}
