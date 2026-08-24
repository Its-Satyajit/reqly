import { useState } from "react";
import { RefreshCw } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { cn } from "#lib/utils";
import {
  getGqlBridge,
  gqlTypeRef,
  type GqlField,
  type GqlSchema,
  type GqlType,
} from "#lib/graphql";

function FieldRow({
  field,
  onNavigate,
}: {
  field: GqlField;
  onNavigate: (typeName: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const returnType = gqlTypeRef(field.type);
  const targetName = field.type?.of?.name ?? field.type?.name;
  return (
    <li className="rounded border border-border/60 px-2 py-1 text-xs">
      <button type="button" className="w-full text-left" onClick={() => setOpen(!open)}>
        <span className="font-mono">{field.name}</span>
        <span className="ml-1.5 font-mono text-status-info">{returnType}</span>
        {field.deprecated && (
          <Badge variant="outline" className="ml-1.5 text-warning">deprecated</Badge>
        )}
      </button>
      {open && (
        <div className="mt-1 flex flex-col gap-1 pl-4">
          {field.description && (
            <p className="text-[11px] text-muted-foreground">{field.description}</p>
          )}
          {(field.args ?? []).length > 0 && (
            <p className="font-mono text-[11px]">
              args:{" "}
              {(field.args ?? [])
                .map((a) => `${a.name}: ${gqlTypeRef(a.type)}`)
                .join(", ")}
            </p>
          )}
          {targetName && (
            <Button
              variant="ghost"
              size="sm"
              className="self-start font-mono"
              onClick={() => onNavigate(targetName)}
            >
              → {targetName}
            </Button>
          )}
        </div>
      )}
    </li>
  );
}

function TypeSection({
  title,
  typ,
  selected,
  onNavigate,
}: {
  title: string;
  typ?: GqlType | null;
  selected: string | null;
  onNavigate: (typeName: string) => void;
}) {
  if (!typ) return null;
  return (
    <div className={cn("rounded-md border p-2", selected === typ.name && "border-primary")}>
      <p className="pb-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
        {title} · <span className="font-mono normal-case">{typ.name}</span>
      </p>
      <ul className="flex flex-col gap-1">
        {(typ.fields ?? []).map((f) => (
          <FieldRow key={`${typ.name}-${f.name}`} field={f} onNavigate={onNavigate} />
        ))}
      </ul>
    </div>
  );
}

export function GraphqlBrowser() {
  const [endpoint, setEndpoint] = useState("");
  const [authHeader, setAuthHeader] = useState("");
  const [schema, setSchema] = useState<GqlSchema | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<string | null>(null);

  const introspect = (): void => {
    if (endpoint.trim() === "") return;
    setBusy(true);
    setError(null);
    const headers =
      authHeader.trim() === ""
        ? undefined
        : [{ key: "Authorization", value: authHeader.trim() }];
    getGqlBridge()
      .introspect({ endpoint: endpoint.trim(), headers, timeoutSec: 30 })
      .then((s) => {
        setSchema(s);
        setBusy(false);
      })
      .catch((error) => {
        setError(error instanceof Error ? error.message : String(error));
        setBusy(false);
      });
  };

  const navigate = (typeName: string) => setSelected(typeName);
  const selectedType: GqlType | undefined = schema?.types?.find((t) => t.name === selected);

  return (
    <section className="flex h-full min-h-0 flex-col gap-3 overflow-y-auto p-4" aria-label="GraphQL schema browser">
      <h2 className="text-sm font-semibold">GraphQL Schema Browser</h2>

      <div className="flex flex-wrap items-end gap-2">
        <div className="flex min-w-64 flex-1 flex-col gap-1">
          <label htmlFor="gql-endpoint" className="text-xs font-medium">Endpoint</label>
          <Input
            id="gql-endpoint"
            value={endpoint}
            onChange={(e) => setEndpoint(e.target.value)}
            placeholder="https://api.example.com/graphql"
            spellCheck={false}
            className="font-mono text-xs"
          />
        </div>
        <div className="flex min-w-48 flex-col gap-1">
          <label htmlFor="gql-auth" className="text-xs font-medium">Bearer token (optional)</label>
          <Input
            id="gql-auth"
            value={authHeader}
            onChange={(e) => setAuthHeader(e.target.value)}
            placeholder="eyJ…"
            spellCheck={false}
            type="password"
            className="font-mono text-xs"
          />
        </div>
        <Button size="sm" disabled={busy || endpoint.trim() === ""} onClick={introspect}>
          {busy ? <Spinner data-icon="inline-start" /> : <RefreshCw data-icon="inline-start" />}
          Introspect
        </Button>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {schema && (
        <div className="flex min-h-0 flex-col gap-2 pb-4">
          <div className="flex flex-col gap-2">
            {(
              [
                ["Query", schema.query],
                ["Mutation", schema.mutation],
                ["Subscription", schema.subscription],
              ] as const
            ).map(([title, root]) =>
              root == null ? null : (
                <TypeSection
                  key={title}
                  title={title}
                  typ={root}
                  selected={selected}
                  onNavigate={navigate}
                />
              ),
            )}
          </div>

          <details className="rounded-md border border-border p-2" open={selected != null}>
            <summary className="cursor-pointer text-xs font-medium">
              Types ({schema.types?.length ?? 0})
            </summary>
            <ul className="flex flex-wrap gap-1 pt-2">
              {(schema.types ?? []).map((t) => (
                <li key={t.name}>
                  <button
                    type="button"
                    onClick={() => setSelected(t.name)}
                    className={cn(
                      "rounded border px-1.5 py-0.5 font-mono text-[11px]",
                      selected === t.name
                        ? "border-primary text-primary"
                        : "border-border text-muted-foreground hover:text-foreground",
                    )}
                  >
                    {t.name}
                    <span className="ml-1 text-[10px] opacity-60">{t.kind}</span>
                  </button>
                </li>
              ))}
            </ul>
          </details>

          {selectedType && (
            <div className="rounded-md border border-border p-2">
              <p className="pb-1 text-xs font-medium">
                <span className="font-mono">{selectedType.name}</span>{" "}
                <span className="text-muted-foreground">{selectedType.kind}</span>
              </p>
              {selectedType.description && (
                <p className="pb-1 text-[11px] text-muted-foreground">{selectedType.description}</p>
              )}
              {(selectedType.enumValues?.length ?? 0) > 0 && (
                <p className="font-mono text-[11px]">
                  enum: {selectedType.enumValues!.join(" | ")}
                </p>
              )}
              <ul className="flex flex-col gap-1 pt-1">
                {(selectedType.fields ?? []).map((f) => (
                  <FieldRow
                    key={`${selectedType.name}-${f.name}`}
                    field={f}
                    onNavigate={navigate}
                  />
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </section>
  );
}
