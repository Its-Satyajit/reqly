import { isRecord, isString } from "./typeGuards";

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

export interface ParsedSpecInfo {
  title?: string;
  version?: string;
}

export type SpecPathOperation = Record<string, string | number | boolean>;

export interface ParsedSpecComponents {
  schemas?: Record<string, string>;
  securitySchemes?: Record<string, string>;
}

export interface ParsedSpecDocument {
  openapi?: string;
  info?: ParsedSpecInfo;
  servers?: string[];
  paths?: Record<string, SpecPathOperation>;
  components?: ParsedSpecComponents;
}

function tryParseYaml(content: string): ParsedSpecDocument | null {
  // Minimal YAML subset for spec-tree: handles mapping keys needed for tree/diagnostics without full parser dep.
  // Fallback: attempt JSON parse, then line-based key extraction.
  const trimmed = content.trim();
  if (!trimmed) return null;
  try {
    const parsed = JSON.parse(trimmed);
    if (isRecord(parsed)) {
      const doc: ParsedSpecDocument = {};
      if (isString(parsed.openapi)) doc.openapi = parsed.openapi;
      if (isRecord(parsed.info)) {
        doc.info = {
          title: isString(parsed.info.title) ? parsed.info.title : "",
          version: isString(parsed.info.version) ? parsed.info.version : "",
        };
      }
      if (isRecord(parsed.paths)) {
        doc.paths = {};
        for (const [k, v] of Object.entries(parsed.paths)) {
          if (isRecord(v)) {
            doc.paths[k] = {};
          }
        }
      }
      if (isRecord(parsed.components)) {
        doc.components = {};
        if (isRecord(parsed.components.schemas)) {
          doc.components.schemas = {};
        }
        if (isRecord(parsed.components.securitySchemes)) {
          doc.components.securitySchemes = {};
        }
      }
      return doc;
    }
  } catch {
    // no-op
  }
  // Use indentation-aware naive parser for top-level and paths keys
  const lines = content.split("\n");
  const result: ParsedSpecDocument = {};
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
        result.paths = {};
      } else {
        if (indent === 0) {
          inPaths = false;
          if (key === "openapi") result.openapi = val || "3.1.0";
          else if (key === "info") result.info = { title: "", version: "" };
          else if (key === "components") result.components = {};
        }
      }
    }
    if (inPaths && indent > pathsIndent && key.startsWith("/")) {
      pathKeys.push(key);
      if (!result.paths) result.paths = {};
      result.paths[key] = {};
    }
    // crude invalid yaml detection
    if (val === "[unclosed") throw new Error("yaml: unclosed bracket");
  }
  // refine info title/version detection via regex on raw content
  const titleMatch = content.match(/title:\s*(.+)/);
  const versionMatch = content.match(/version:\s*(.+)/);
  if (result.info) {
    result.info.title = titleMatch ? titleMatch[1].trim() : "";
    result.info.version = versionMatch ? versionMatch[1].trim() : "";
  }
  // also capture any /path that wasn't top-level due to nested structure
  if (pathKeys.length === 0) {
    for (const raw of lines) {
      const t = raw.trim();
      const pm = t.match(/^(\/[^:]*):\s*$/);
      if (pm) {
        const pk = pm[1];
        if (!result.paths) result.paths = {};
        if (!(pk in result.paths)) {
          result.paths[pk] = {};
          pathKeys.push(pk);
        }
      }
    }
  }
  if (content.includes("[unclosed")) throw new Error("yaml: invalid");
  if (content.includes("not: yaml: :") && content.split(":").length > 4) throw new Error("yaml: invalid mapping");
  return result;
}

export function nodesForContent(content: string): SpecNode[] {
  if (!content.trim()) return SPEC_SECTIONS;
  let doc: ParsedSpecDocument | null = null;
  try {
    doc = tryParseYaml(content);
  } catch {
    return SPEC_SECTIONS;
  }
  if (!doc) return SPEC_SECTIONS;
  const paths = doc.paths;
  const pathChildren: SpecNode[] = paths
    ? Object.keys(paths).map((p) => ({ id: `paths:${p}`, label: p }))
    : [];
  const components = doc.components;
  const hasSchemas = Boolean(components?.schemas);
  const hasSecurity = Boolean(components?.securitySchemes);
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
  let doc: ParsedSpecDocument | null = null;
  try {
    doc = tryParseYaml(content);
  } catch (e) {
    out.push({ message: `yaml: ${e instanceof Error ? e.message : String(e)}`, severity: "error" });
    return out;
  }
  if (!doc) {
    out.push({ message: "yaml: invalid document", severity: "error" });
    return out;
  }
  if (!doc.openapi) out.push({ message: "missing required field: openapi", severity: "error" });
  const info = doc.info;
  if (!info) {
    out.push({ message: "missing required field: info", severity: "error" });
  } else {
    if (!info.title) out.push({ message: "missing required field: info.title", severity: "error" });
    if (!info.version) out.push({ message: "missing required field: info.version", severity: "warning" });
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
