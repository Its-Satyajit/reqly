export interface DatasetColumn {
  name: string;
  index: number;
}

export interface DatasetRow {
  index: number;
  values: { [key: string]: string };
}

export interface Dataset {
  name: string;
  columns: DatasetColumn[];
  rows: DatasetRow[];
  source: "csv" | "json";
  rawContent: string;
}

/** Parse a single CSV line respecting quoted fields. */
function parseCsvLine(line: string): string[] {
  const fields: string[] = [];
  let current = "";
  let inQuotes = false;
  for (let i = 0; i < line.length; i++) {
    const ch = line[i];
    if (inQuotes) {
      if (ch === '"') {
        if (i + 1 < line.length && line[i + 1] === '"') {
          current += '"';
          i++;
        } else {
          inQuotes = false;
        }
      } else {
        current += ch;
      }
    } else {
      if (ch === '"') {
        inQuotes = true;
      } else if (ch === ",") {
        fields.push(current);
        current = "";
      } else {
        current += ch;
      }
    }
  }
  fields.push(current);
  return fields;
}

export function parseCsv(content: string, name = "dataset"): Dataset {
  if (content.trim() === "") {
    return { name, columns: [], rows: [], source: "csv", rawContent: content };
  }

  // First pass: split into logical lines respecting quoted newlines
  const logicalLines: string[] = [];
  let current = "";
  let inQuotes = false;
  for (let i = 0; i < content.length; i++) {
    const ch = content[i];
    if (inQuotes) {
      if (ch === '"') {
        if (i + 1 < content.length && content[i + 1] === '"') {
          current += '"';
          i++;
        } else {
          inQuotes = false;
          current += ch;
        }
      } else {
        if (ch === "\n") {
          current += "\n"; // preserve newline inside quoted field
        } else if (ch === "\r") {
          // skip \r
        } else {
          current += ch;
        }
      }
    } else {
      if (ch === '"') {
        inQuotes = true;
        current += ch;
      } else if (ch === "\n" || ch === "\r") {
        if (current.length > 0 || logicalLines.length === 0) {
          logicalLines.push(current);
        }
        current = "";
        if (ch === "\r" && i + 1 < content.length && content[i + 1] === "\n") i++;
      } else {
        current += ch;
      }
    }
  }
  if (current.length > 0 || logicalLines.length === 0) logicalLines.push(current);

  const headerFields = parseCsvLine(logicalLines[0]);
  const columns: DatasetColumn[] = headerFields.map((h, i) => ({
    name: h.trim(),
    index: i,
  }));

  const rows: DatasetRow[] = [];
  for (let i = 1; i < logicalLines.length; i++) {
    if (logicalLines[i].trim() === "") continue;
    const fields = parseCsvLine(logicalLines[i]);
    const values: { [key: string]: string } = {};
    for (let j = 0; j < columns.length; j++) {
      values[columns[j].name] = fields[j] ?? "";
    }
    rows.push({ index: rows.length, values });
  }

  return { name, columns, rows, source: "csv", rawContent: content };
}

/** A row object parsed from JSON — string values for all fields. */
type JsonRow = { [key: string]: string };

export function parseJsonDataset(content: string, name = "dataset"): Dataset {
  const parsed: unknown = JSON.parse(content);
  if (!Array.isArray(parsed)) throw new Error("Dataset JSON must be an array of objects");
  if (parsed.length === 0) {
    return { name, columns: [], rows: [], source: "json", rawContent: content };
  }

  // SAFETY: we validated Array.isArray above; elements are user-provided objects
  // that we immediately convert to string values — no unsafe access pattern.
  const objects = parsed as JsonRow[];
  const keys = Object.keys(objects[0]);
  const columns: DatasetColumn[] = keys.map((k, i) => ({ name: k, index: i }));

  const rows: DatasetRow[] = objects.map((obj, i) => {
    const values: JsonRow = {};
    for (const key of keys) {
      values[key] = String(obj[key] ?? "");
    }
    return { index: i, values };
  });

  return { name, columns, rows, source: "json", rawContent: content };
}

export function getRowVariables(dataset: Dataset, rowIndex: number): { [key: string]: string } {
  const row = dataset.rows[rowIndex];
  return row ? { ...row.values } : {};
}

export function validateDataset(dataset: Dataset): string[] {
  const errors: string[] = [];
  if (dataset.rows.length === 0) errors.push("Dataset has no data rows");
  const names = dataset.columns.map((c) => c.name);
  const seen = new Set<string>();
  for (const n of names) {
    if (seen.has(n)) errors.push(`Duplicate column name: "${n}"`);
    seen.add(n);
  }
  return errors;
}
