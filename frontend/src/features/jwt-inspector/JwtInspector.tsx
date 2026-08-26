import { useState } from "react";
import { ClipboardPaste, ScanSearch } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Card, CardContent } from "#components/ui/card";
import { cn } from "#lib/utils";
import { Spinner } from "#components/ui/spinner";
import { Textarea } from "#components/ui/textarea";
import {
  extractBearer,
  formatRemaining,
  getJwtBridge,
  type JwtClaim,
  type JwtTokenView,
} from "#lib/jwt";
import { useRequestStore } from "#stores/useRequestStore";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { ViewShell } from "../../components/shell/ViewLayout";

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

export function JwtInspector() {
  const [token, setToken] = useState("");
  const [decoded, setDecoded] = useState<JwtTokenView | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Promise-chain form keeps hook updates out of try/catch, which the React
  // Compiler cannot model.
  const decode = (raw: string): void => {
    setBusy(true);
    setError(null);
    getJwtBridge()
      .decode(raw)
      .then((view) => {
        setDecoded(view);
        setBusy(false);
      })
      .catch((error) => {
        setDecoded(null);
        setError(error instanceof Error ? error.message : String(error));
        setBusy(false);
      });
  };

  return (
    <ViewShell label="JWT inspector">
      <h2 className="text-sm font-semibold">JWT Inspector</h2>

      <Card size="sm">
        <CardContent className="flex flex-col gap-2">
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
          <Button
            size="sm"
            disabled={token.trim() === "" || busy}
            onClick={() => decode(token.trim())}
          >
            {busy ? (
              <Spinner data-icon="inline-start" />
            ) : (
              <ScanSearch data-icon="inline-start" />
            )}
            Decode
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              const activeId = useWorkspaceStore.getState().activeTabId ?? "";
              const auth =
                useRequestStore.getState().responses[activeId]?.response?.headers
                  ?.authorization?.[0];
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
        </CardContent>
      </Card>

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
            <p
              className={cn(
                "pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground",
              )}
            >
              Header
            </p>
            <ClaimsTable claims={decoded.header} />
          </div>
          <div className="rounded-md border border-border p-2">
            <p className="pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Payload (claims)
            </p>
            <ClaimsTable
              claims={decoded.payload}
              epochs={{ exp: decoded.expiry.exp, iat: decoded.expiry.iat }}
            />
          </div>
        </div>
      )}
    </ViewShell>
  );
}
