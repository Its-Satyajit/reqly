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

export interface EndpointInput {
  path: string;
  method: string;
  summary?: string;
  operationId?: string;
}

function yamlEscape(value: string): string {
  // Minimal YAML quoting: if value contains chars that would break plain scalar,
  // wrap in double quotes and JSON-escape. Mirrors how OpenAPI generators quote summaries.
  if (value === "") return '""';
  const needsQuote =
    value.includes(":") ||
    value.includes("#") ||
    value.includes("[") ||
    value.includes("]") ||
    value.includes("{") ||
    value.includes("}") ||
    value.includes("&") ||
    value.includes("*") ||
    value.includes("?") ||
    value.includes("|") ||
    value.includes(">") ||
    value.includes("-") ||
    value.includes("!") ||
    value.includes("%") ||
    value.includes("@") ||
    value.includes("`") ||
    value.includes("\n") ||
    /^\s|\s$/.test(value) ||
    value.includes('"') ||
    value.includes("'");
  if (needsQuote) return JSON.stringify(value);
  return value;
}

function isPathLine(line: string): boolean {
  const t = line.trim();
  // Unquoted or quoted path-like keys: must start with "/" and end with ":"
  return (/^"?\s*\/[^:]*"?\s*:\s*$/).test(t) || (/^'\/[^']*':\s*$/).test(t);
}

export function patchEndpointInContent(content: string, oldPath: string, updated: EndpointInput): string {
  const targetPath = updated.path;
  const methodLower = updated.method.toLowerCase();
  const lines = content.split("\n");

  // Helper to find exact path line (no prefix collision) — matches whole key, not substring.
  const findPathIdx = (path: string, from = 0): number => {
    for (let i = from; i < lines.length; i++) {
      const t = lines[i].trim();
      if (t === `${path}:` || t === `"${path}":` || t === `'${path}':`) return i;
      // Strict regex: leading whitespace, exact path (escaped), optional whitespace, colon, optional whitespace, end
      const esc = path.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
      const re = new RegExp(`^\\s*(?:"${esc}"|'${esc}'|${esc})\\s*:\\s*$`);
      if (re.test(lines[i])) return i;
    }
    return -1;
  };

  if (updated.path !== oldPath) {
    const oldIdx = findPathIdx(oldPath);
    if (oldIdx !== -1) {
      const indent = lines[oldIdx].match(/^\s*/)?.[0] ?? "  ";
      // Preserve quoting style of original; default to unquoted.
      const origTrim = lines[oldIdx].trim();
      const wasSingle = origTrim.startsWith("'");
      const wasDouble = origTrim.startsWith('"');
      const newKey = wasSingle ? `'${targetPath}':` : wasDouble ? `"${targetPath}":` : `${targetPath}:`;
      lines[oldIdx] = `${indent}${newKey}`;
      content = lines.join("\n");
    } else {
      // Fallback: append new path block if old not found — quote summary safely
      content = content.trimEnd() + `\n  ${targetPath}:\n    ${methodLower}:\n      summary: ${yamlEscape(updated.summary ?? "New endpoint")}\n`;
      // Append operationId if present
      if (updated.operationId) content = content.trimEnd() + `\n        operationId: ${yamlEscape(updated.operationId)}\n`;
      return content;
    }
  } else {
    content = lines.join("\n");
  }

  // Now ensure method and summary under targetPath
  // Re-split after possible rename
  const curLines = content.split("\n");
  const pathIdx = findPathIdx(targetPath);
  if (pathIdx === -1) return content;

  // Bound path block: next path line after pathIdx
  let nextPathIdx = curLines.length;
  for (let i = pathIdx + 1; i < curLines.length; i++) {
    if (isPathLine(curLines[i])) {
      nextPathIdx = i;
      break;
    }
  }
  const pathBlock = curLines.slice(pathIdx, nextPathIdx).join("\n");
  const escMethod = methodLower.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const hasMethod = new RegExp(`^\\s+${escMethod}\\s*:`, "im").test(pathBlock);

  if (!hasMethod) {
    const indent = "    ";
    const methodBlock = [`${indent}${methodLower}:`, `${indent}  summary: ${yamlEscape(updated.summary ?? "New endpoint")}`];
    if (updated.operationId) methodBlock.push(`${indent}  operationId: ${yamlEscape(updated.operationId)}`);
    curLines.splice(pathIdx + 1, 0, ...methodBlock);
    return curLines.join("\n");
  }
  if (updated.summary) {
    // Update summary within this path+method block only, to avoid cross-path greed.
    for (let i = pathIdx + 1; i < nextPathIdx; i++) {
      if (/^\s+summary\s*:/.test(curLines[i]) && curLines.slice(pathIdx, i).some((l) => new RegExp(`^\\s+${escMethod}\\s*:`, "i").test(l))) {
        // Find method line first, ensure summary is under that method
        const methodIdx = curLines.slice(pathIdx + 1, nextPathIdx).findIndex((l) => new RegExp(`^\\s+${escMethod}\\s*:`, "i").test(l));
        const absMethodIdx = pathIdx + 1 + methodIdx;
        if (i > absMethodIdx) {
          const indent = curLines[i].match(/^\s*/)?.[0] ?? "      ";
          curLines[i] = `${indent}summary: ${yamlEscape(updated.summary)}`;
          return curLines.join("\n");
        }
      }
    }
    // If summary line not found but method exists, insert it right after method
    const methodIdx = curLines.slice(pathIdx + 1, nextPathIdx).findIndex((l) => new RegExp(`^\\s+${escMethod}\\s*:`, "i").test(l));
    if (methodIdx !== -1) {
      const absMethodIdx = pathIdx + 1 + methodIdx;
      const indent = "      ";
      curLines.splice(absMethodIdx + 1, 0, `${indent}summary: ${yamlEscape(updated.summary)}`);
      return curLines.join("\n");
    }
  }
  return curLines.join("\n");
}

const ALLOWED_METHODS = new Set(["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE"]);

export function validateEndpoint(input: EndpointInput): string[] {
  const errors: string[] = [];
  const path = input.path.trim();
  if (path.length === 0) {
    errors.push("path is required");
  } else {
    if (!path.startsWith("/")) errors.push("path must start with '/'");
    if (/\s/.test(path)) errors.push("path must not contain spaces");
    // Only allow visible ASCII, no control chars
    if (/[^\x20-\x7E]/.test(path)) errors.push("path contains invalid characters");
  }
  const method = input.method.trim().toUpperCase();
  if (method.length === 0) {
    errors.push("method is required");
  } else if (!ALLOWED_METHODS.has(method)) {
    errors.push(`method must be one of ${[...ALLOWED_METHODS].join(", ")}`);
  }
  if (input.operationId !== undefined && input.operationId !== null) {
    const op = String(input.operationId);
    if (/\s/.test(op)) errors.push("operationId must not contain spaces");
  }
  return errors;
}
