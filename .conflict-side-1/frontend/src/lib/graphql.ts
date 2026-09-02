export interface GqlArg {
  name: string;
  type?: GqlTypeRef | null;
  def?: string;
}

export interface GqlTypeRef {
  name?: string;
  kind?: string;
  of?: GqlTypeRef | null;
}

export interface GqlField {
  name: string;
  description?: string;
  type?: GqlTypeRef | null;
  args?: GqlArg[];
  deprecated?: boolean;
}

export interface GqlType {
  kind: string;
  name: string;
  description?: string;
  fields?: GqlField[];
  enumValues?: string[];
}

export interface GqlSchema {
  query?: GqlType | null;
  mutation?: GqlType | null;
  subscription?: GqlType | null;
  types?: GqlType[] | null;
}

/** gqlTypeRef renders a wrapped reference GraphQL-style. */
export function gqlTypeRef(ref?: GqlTypeRef | null): string {
  if (!ref) return "?";
  if (ref.kind === "LIST") return `[${gqlTypeRef(ref.of ?? undefined)}]`;
  if (ref.kind === "NON_NULL") {
    const inner = gqlTypeRef(ref.of ?? undefined);
    return inner.endsWith("!") ? inner : `${inner}!`;
  }
  return ref.name ?? "?";
}

export interface GqlAdapter {
  introspect(input: {
    endpoint: string;
    headers?: { key: string; value: string }[];
    timeoutSec?: number;
  }): Promise<GqlSchema>;
  parse(input: { schemaPath: string; typeFilter?: string }): Promise<string>;
}

// Bridge registry, same pattern as the other feature adapters.
let bridge: GqlAdapter | null = null;

/** setGqlBridge installs the host adapter (called once from the bridge). */
export function setGqlBridge(adapter: GqlAdapter): void {
  bridge = adapter;
}

export function getGqlBridge(): GqlAdapter {
  if (!bridge) throw new Error("graphql browser is not available in this build");
  return bridge;
}

/** ROOT_SECTIONS orders the schema tree's root operations first. */
export const ROOT_SECTIONS = ["Query", "Mutation", "Subscription"] as const;
