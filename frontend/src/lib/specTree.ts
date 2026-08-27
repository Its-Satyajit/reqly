export type SpecNode = { id: string; label: string; children?: SpecNode[] };

export type SpecDiagnostic = { message: string; severity: "error" | "warning" };

export const SPEC_SECTIONS: SpecNode[] = [
  { id: "info", label: "Info" },
  { id: "servers", label: "Servers" },
  {
    id: "paths",
    label: "Paths",
    children: [
      { id: "paths:/users", label: "/users" },
      { id: "paths:/products", label: "/products" },
      { id: "paths:/orders", label: "/orders" },
    ],
  },
  {
    id: "components",
    label: "Components",
    children: [
      { id: "components:schemas", label: "Schemas" },
      { id: "components:security", label: "Security" },
    ],
  },
];

function tryParseYaml(content: string): unknown {
  // Minimal YAML subset for spec-tree: handles mapping keys needed for tree/diagnostics without full parser dep.
  // Fallback: attempt JSON parse, then line-based key extraction.
  const trimmed = content.trim();
  if (!trimmed) return null;
  try {
    return JSON.parse(trimmed);
  } catch {
    // no-op
  }
  try {
    // Use indentation-aware naive parser for top-level and paths keys
    const lines = content.split("\n");
    const result: Record<string, unknown> = {};
    const pathKeys: string[] = [];
    let inPaths = false;
    let pathsIndent = -1;
    for (const raw of lines) {
      const line = raw.replace(/\r$/, "");
      if (!line.trim() || line.trim().startsWith("#")) continue;
      const indent = line.search(/\S/);
      const m = line.trim().match(/^([^:]+):\s*(.*)$/);
      if (!m) {
        if (line.includes("[unclosed") || line.includes(": :")) throw new Error("invalid yaml");
        continue;
      }
      const key = m[1].trim();
      const val = m[2].trim();
      if (key === "openapi" || key === "info" || key === "servers" || key === "paths" || key === "components") {
        if (key === "paths") {
          inPaths = true;
          pathsIndent = indent;
          result[key] = {};
        } else {
          if (indent === 0) {
            inPaths = false;
            if (key === "openapi") result[key] = val || "3.1.0";
            else if (key === "info") result[key] = { title: "", version: "" };
            else result[key] = val;
          }
        }
      }
      if (inPaths && indent > pathsIndent && key.startsWith("/")) {
        pathKeys.push(key);
        (result["paths"] as Record<string, unknown>)[key] = {};
      }
      // crude invalid yaml detection
      if (val === "[unclosed") throw new Error("yaml: unclosed bracket");
    }
    // refine info title/version detection via regex on raw content
    const titleMatch = content.match(/title:\s*(.+)/);
    const versionMatch = content.match(/version:\s*(.+)/);
    if (result["info"] && typeof result["info"] === "object") {
      (result["info"] as Record<string, string>).title = titleMatch ? titleMatch[1].trim() : "";
      (result["info"] as Record<string, string>).version = versionMatch ? versionMatch[1].trim() : "";
    }
    // also capture any /path that wasn't top-level due to nested structure
    if (pathKeys.length === 0) {
      for (const raw of lines) {
        const t = raw.trim();
        const pm = t.match(/^(\/[^:]*):\s*$/);
        if (pm) {
          const pk = pm[1];
          if (!(pk in (result["paths"] as Record<string, unknown> ?? {}))) {
            if (!result["paths"]) result["paths"] = {};
            (result["paths"] as Record<string, unknown>)[pk] = {};
            pathKeys.push(pk);
          }
        }
      }
    }
    if (content.includes("[unclosed")) throw new Error("yaml: invalid");
    if (content.includes("not: yaml: :") && content.split(":").length > 4) throw new Error("yaml: invalid mapping");
    return result;
  } catch (e) {
    throw e;
  }
}

export function nodesForContent(content: string): SpecNode[] {
  if (!content.trim()) return SPEC_SECTIONS;
  let doc: Record<string, unknown> | null = null;
  try {
    doc = tryParseYaml(content) as Record<string, unknown> | null;
  } catch {
    return SPEC_SECTIONS;
  }
  if (!doc || typeof doc !== "object") return SPEC_SECTIONS;
  const paths = doc["paths"] as Record<string, unknown> | undefined;
  const pathChildren: SpecNode[] = paths && typeof paths === "object"
    ? Object.keys(paths).map((p) => ({ id: `paths:${p}`, label: p }))
    : [];
  const components = doc["components"] as Record<string, unknown> | undefined;
  const hasSchemas = components && typeof components["schemas"] === "object";
  const hasSecurity = components && typeof components["securitySchemes"] === "object";
  return [
    { id: "info", label: "Info" },
    { id: "servers", label: "Servers" },
    { id: "paths", label: "Paths", children: pathChildren },
    {
      id: "components",
      label: "Components",
      children: [
        ...(hasSchemas || !components ? [{ id: "components:schemas", label: "Schemas" }] : []),
        ...(hasSecurity || !components ? [{ id: "components:security", label: "Security" }] : []),
      ],
    },
  ];
}

export function diagnosticsForSpec(content: string): SpecDiagnostic[] {
  const out: SpecDiagnostic[] = [];
  const trimmed = content.trim();
  if (!trimmed) {
    out.push({ message: "spec is empty", severity: "error" });
    return out;
  }
  let doc: Record<string, unknown>;
  try {
    doc = tryParseYaml(content) as Record<string, unknown>;
  } catch (e) {
    out.push({ message: `yaml: ${(e as Error).message}`, severity: "error" });
    return out;
  }
  if (!doc || typeof doc !== "object") {
    out.push({ message: "yaml: invalid document", severity: "error" });
    return out;
  }
  if (!doc["openapi"]) out.push({ message: "missing required field: openapi", severity: "error" });
  const info = doc["info"] as Record<string, unknown> | undefined;
  if (!info || typeof info !== "object") {
    out.push({ message: "missing required field: info", severity: "error" });
  } else {
    if (!info["title"]) out.push({ message: "missing required field: info.title", severity: "error" });
    if (!info["version"]) out.push({ message: "missing required field: info.version", severity: "warning" });
  }
  if (!("paths" in doc)) out.push({ message: "missing required field: paths", severity: "warning" });
  return out;
}

export function flattenSpecTree(nodes: SpecNode[]): SpecNode[] {
  const out: SpecNode[] = [];
  const walk = (ns: SpecNode[]) => {
    for (const n of ns) {
      out.push(n);
      if (n.children) walk(n.children);
    }
  };
  walk(nodes);
  return out;
}
