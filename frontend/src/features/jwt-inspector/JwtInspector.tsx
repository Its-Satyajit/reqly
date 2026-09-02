import { useState } from "react";
import { ClipboardPaste, ScanSearch, ShieldCheck, PenLine } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { Textarea } from "#components/ui/textarea";
import { cn } from "#lib/utils";
import {
  extractBearer,
  formatRemaining,
  getJwtBridge,
  type JwtClaim,
  type JwtTokenView,
} from "#lib/jwt";
import { useRequestStore } from "#stores/useRequestStore";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";

function expiryBadge(expiry: JwtTokenView["expiry"]) {
  switch (expiry.status) {
    case "valid":
      return (
        <Badge variant="outline" className="border-status-ok/40 text-status-ok">
          valid · {formatRemaining(expiry.remaining)} left
        </Badge>
      );
    case "expired":
      return <Badge variant="destructive">expired</Badge>;
    case "not_yet_valid":
      return (
        <Badge variant="outline" className="text-warning">
          not yet valid · in {formatRemaining(-expiry.remaining)}
        </Badge>
      );
    default:
      return <Badge variant="ghost">no expiry</Badge>;
  }
}

function ClaimsTable({
  claims,
  epochs,
}: {
  claims: JwtClaim[];
  epochs?: { exp?: number; iat?: number };
}) {
  if (claims.length === 0) {
    return <p className="text-xs text-muted-foreground">(empty)</p>;
  }
  return (
    <table className="w-full table-fixed text-left text-xs">
      <tbody>
        {claims.map((claim) => (
          <tr key={claim.key} className="border-b border-border/50 last:border-0">
            <th scope="row" className="w-28 py-0.5 pr-2 align-top font-medium">
              {claim.key}
              {(claim.key === "exp" || claim.key === "iat") &&
                (claim.key === "exp" ? epochs?.exp : epochs?.iat) != null && (
                  <span className="block font-normal text-muted-foreground">
                    {new Date(
                      (claim.key === "exp" ? epochs?.exp : epochs?.iat)! * 1000,
                    ).toLocaleString()}
                  </span>
                )}
            </th>
            <td className="break-all py-0.5 font-mono">{claim.value}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

type Tab = "decode" | "verify" | "sign";

export function JwtInspector() {
  const [active, setActive] = useState<Tab>("decode");
  const [token, setToken] = useState("");
  const [decoded, setDecoded] = useState<JwtTokenView | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [verifySecret, setVerifySecret] = useState("");
  const [verifyResult, setVerifyResult] = useState<boolean | null>(null);
  const [verifyBusy, setVerifyBusy] = useState(false);

  const [signPayload, setSignPayload] = useState('{"sub":"alice","exp":1893456000}');
  const [signSecret, setSignSecret] = useState("secret");
  const [signAlg, setSignAlg] = useState("HS256");
  const [signResult, setSignResult] = useState<string | null>(null);
  const [signBusy, setSignBusy] = useState(false);

  const decode = (raw: string): void => {
    setBusy(true);
    setError(null);
    getJwtBridge()
      .decode(raw)
      .then((view) => {
        setDecoded(view);
        setBusy(false);
      })
      .catch((err) => {
        setDecoded(null);
        setError(err instanceof Error ? err.message : String(err));
        setBusy(false);
      });
  };

  const verify = (): void => {
    if (!token.trim() || !verifySecret.trim()) {
      setError("token and secret are required");
      return;
    }
    setVerifyBusy(true);
    setError(null);
    setVerifyResult(null);
    getJwtBridge()
      .verify(token.trim(), verifySecret)
      .then((ok) => {
        setVerifyResult(ok);
        setVerifyBusy(false);
      })
      .catch((err) => {
        setError(err instanceof Error ? err.message : String(err));
        setVerifyBusy(false);
      });
  };

  const sign = (): void => {
    if (!signSecret.trim()) {
      setError("secret is required");
      return;
    }
    setSignBusy(true);
    setError(null);
    setSignResult(null);
    getJwtBridge()
      .sign(signPayload, signSecret, signAlg)
      .then((tok) => {
        setSignResult(tok);
        setSignBusy(false);
      })
      .catch((err) => {
        setError(err instanceof Error ? err.message : String(err));
        setSignBusy(false);
      });
  };

  return (
    <div className="flex h-full min-h-0 flex-col overflow-y-auto" aria-label="JWT inspector">
      <PageHeader
        icon={ScanSearch}
        title="JWT Inspector"
        description="Decode, verify HMAC, and sign — local, no network (HS256/HS384/HS512/none)."
      />
      <div className="flex flex-col gap-3 p-4">
        <div className="flex gap-1.5 border-b border-border pb-2">
          {// SAFETY: literal array is exactly Tab union
          (["decode", "verify", "sign"] as Tab[]).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setActive(t)}
              className={cn(
                "rounded px-3 py-1 text-xs font-medium",
                active === t ? "bg-primary text-primary-foreground" : "border border-border hover:bg-muted",
              )}
            >
              {t === "decode" ? "Decode" : t === "verify" ? "Verify" : "Sign"}
            </button>
          ))}
        </div>

        <div className="flex flex-col gap-2 rounded-md border border-border p-3">
          <label htmlFor="jwt-token" className="text-xs font-medium">
            Token (Bearer prefix optional)
          </label>
          <Textarea
            id="jwt-token"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            rows={3}
            spellCheck={false}
            className="resize-y font-mono text-xs"
            placeholder="eyJhbGciOi…"
          />
          <div className="flex gap-1.5">
            <Button size="sm" disabled={token.trim() === "" || busy} onClick={() => decode(token.trim())}>
              {busy ? <Spinner data-icon="inline-start" /> : <ScanSearch data-icon="inline-start" />}
              Decode
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                const activeId = useWorkspaceStore.getState().activeTabId ?? "";
                const auth = useRequestStore.getState().responses[activeId]?.response?.headers?.authorization?.[0];
                const bearer = extractBearer(auth);
                if (!bearer) {
                  setError("No Bearer token in the active tab's response headers.");
                  return;
                }
                setToken(bearer);
                void decode(bearer);
              }}
            >
              <ClipboardPaste data-icon="inline-start" />
              From last response
            </Button>
          </div>
        </div>

        {active === "decode" && (
          <>
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            {decoded && (
              <div className="flex min-h-0 flex-col gap-2 pb-4">
                <div className="flex items-center gap-2">
                  <span className="text-xs text-muted-foreground">alg:</span>
                  <Badge variant="secondary">{decoded.alg || "?"}</Badge>
                  {expiryBadge(decoded.expiry)}
                </div>
                <div className="rounded-md border border-border p-2">
                  <p className={cn("pb-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground")}>Header</p>
                  <ClaimsTable claims={decoded.header} />
                </div>
                <div className="rounded-md border border-border p-2">
                  <p className="pb-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Payload (claims)</p>
                  <ClaimsTable claims={decoded.payload} epochs={{ exp: decoded.expiry.exp, iat: decoded.expiry.iat }} />
                </div>
              </div>
            )}
          </>
        )}

        {active === "verify" && (
          <div className="flex flex-col gap-2 rounded-md border border-border p-3">
            <label className="text-xs font-medium">HMAC Secret</label>
            <Input value={verifySecret} onChange={(e) => setVerifySecret(e.target.value)} placeholder="secret" spellCheck={false} className="font-mono text-xs" type="password" />
            <Button size="sm" onClick={verify} disabled={verifyBusy || !token.trim() || !verifySecret.trim()} className="w-fit gap-1.5">
              {verifyBusy ? <Spinner data-icon="inline-start" /> : <ShieldCheck data-icon="inline-start" />}
              Verify — secret
            </Button>
            {verifyResult !== null && (
              <Badge variant={verifyResult ? "outline" : "destructive"} className={verifyResult ? "border-status-ok/40 text-status-ok w-fit" : "w-fit"}>
                {verifyResult ? "Valid" : "Invalid"} — HMAC {verifyResult ? "matches" : "mismatch"}
              </Badge>
            )}
          </div>
        )}

        {active === "sign" && (
          <div className="flex flex-col gap-2 rounded-md border border-border p-3">
            <label className="text-xs font-medium">Payload JSON</label>
            <Textarea value={signPayload} onChange={(e) => setSignPayload(e.target.value)} rows={4} spellCheck={false} className="font-mono text-xs" placeholder='{"sub":"alice"}' />
            <div className="grid grid-cols-2 gap-2">
              <label className="flex flex-col gap-1">
                <span className="text-xs font-medium">Secret</span>
                <Input value={signSecret} onChange={(e) => setSignSecret(e.target.value)} placeholder="secret" spellCheck={false} className="font-mono text-xs" type="password" />
              </label>
              <label className="flex flex-col gap-1">
                <span className="text-xs font-medium">Alg</span>
                <select value={signAlg} onChange={(e) => setSignAlg(e.target.value)} className="h-8 rounded-md border border-input bg-transparent px-2 text-xs">
                  <option value="HS256">HS256</option>
                  <option value="HS384">HS384</option>
                  <option value="HS512">HS512</option>
                  <option value="none">none</option>
                </select>
              </label>
            </div>
            <Button size="sm" onClick={sign} disabled={signBusy || !signSecret.trim()} className="w-fit gap-1.5">
              {signBusy ? <Spinner data-icon="inline-start" /> : <PenLine data-icon="inline-start" />}
              Sign
            </Button>
            {signResult && (
              <div className="rounded border bg-muted/40 p-2">
                <p className="pb-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Token</p>
                <pre className="break-all font-mono text-xs">{signResult}</pre>
              </div>
            )}
          </div>
        )}

        {active !== "decode" && error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
      </div>
    </div>
  );
}
