export interface SchemaDocField {
  name: string;
  type: string;
  description?: string;
  isRequired: boolean;
  args?: SchemaDocArg[];
}

export interface SchemaDocArg {
  name: string;
  type: string;
  description?: string;
  defaultValue?: string;
}

export interface SchemaDocType {
  name: string;
  kind: "type" | "interface" | "enum" | "input" | "scalar";
  description?: string;
  fields?: SchemaDocField[];
  values?: string[];
}

export interface SchemaDocTree {
  types: SchemaDocType[];
  queryType?: string;
  mutationType?: string;
  subscriptionType?: string;
}

/** Parse a minimal GraphQL SDL string into a SchemaDocTree. */
export function parseGraphQlSchema(sdl: string): SchemaDocTree {
  const types: SchemaDocType[] = [];
  let queryType: string | undefined;
  let mutationType: string | undefined;
  let subscriptionType: string | undefined;

  // Match type definitions: type Name { ... }
  const typeRegex = /type\s+(\w+)\s*\{([^}]*)\}/g;
  let match: RegExpExecArray | null;
  while ((match = typeRegex.exec(sdl)) !== null) {
    const name = match[1];
    const body = match[2];
    const fields = parseFields(body);
    types.push({ name, kind: "type", fields });
    if (name === "Query") queryType = name;
    if (name === "Mutation") mutationType = name;
    if (name === "Subscription") subscriptionType = name;
  }

  // Match enum definitions: enum Name { A B C }
  const enumRegex = /enum\s+(\w+)\s*\{([^}]*)\}/g;
  while ((match = enumRegex.exec(sdl)) !== null) {
    const name = match[1];
    const values = match[2]
      .split(/\s+/)
      .map((v) => v.trim())
      .filter(Boolean);
    types.push({ name, kind: "enum", values });
  }

  return { types, queryType, mutationType, subscriptionType };
}

function parseFields(body: string): SchemaDocField[] {
  const lines = body.split("\n").map((l) => l.trim()).filter(Boolean);
  const fields: SchemaDocField[] = [];
  for (const line of lines) {
    // Skip comments
    if (line.startsWith("#")) continue;
    // Match "name: Type" or "name(args): Type"
    const fieldMatch = line.match(/^(\w+)(?:\(([^)]*)\))?\s*:\s*(.+?)(?:\s*(!))?$/);
    if (fieldMatch) {
      const name = fieldMatch[1];
      const argsRaw = fieldMatch[2];
      const type = fieldMatch[3].trim();
      const isRequired = fieldMatch[4] === "!";
      const args = argsRaw ? parseArgs(argsRaw) : undefined;
      fields.push({ name, type, isRequired, args });
    }
  }
  return fields;
}

function parseArgs(argsRaw: string): SchemaDocArg[] {
  return argsRaw.split(",").map((a) => {
    const [name, type] = a.split(":").map((s) => s.trim());
    return { name: name ?? "", type: type ?? "String", isRequired: false };
  });
}

/** Search types by name, field name, or description. */
export function searchSchemaDocs(tree: SchemaDocTree, query: string): SchemaDocType[] {
  const lower = query.toLowerCase();
  return tree.types.filter(
    (t) =>
      t.name.toLowerCase().includes(lower) ||
      (t.description?.toLowerCase().includes(lower) ?? false) ||
      t.fields?.some(
        (f) =>
          f.name.toLowerCase().includes(lower) ||
          f.type.toLowerCase().includes(lower) ||
          (f.description?.toLowerCase().includes(lower) ?? false),
      ) ||
      t.values?.some((v) => v.toLowerCase().includes(lower)),
  );
}

/** Render a SchemaDocType as a Markdown documentation string. */
export function renderDocMarkdown(type: SchemaDocType): string {
  const lines: string[] = [];
  lines.push(`## ${type.name}`);
  if (type.description) lines.push(`\n${type.description}`);

  if (type.kind === "enum" && type.values) {
    lines.push("\n**Values:**");
    for (const v of type.values) {
      lines.push(`- \`${v}\``);
    }
  }

  if (type.fields && type.fields.length > 0) {
    lines.push("\n**Fields:**\n");
    lines.push("| Name | Type | Required |");
    lines.push("|------|------|----------|");
    for (const f of type.fields) {
      const req = f.isRequired ? "yes" : "no";
      lines.push(`| \`${f.name}\` | \`${f.type}\` | ${req} |`);
    }
  }

  return lines.join("\n");
}
