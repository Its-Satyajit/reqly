// Reqly - Dataset Picker for Data-driven Testing (§56.5)
// SPDX-License-Identifier: Apache-2.0
import { useCallback, useState } from "react";
import { Upload, FileText, AlertCircle, X } from "lucide-react";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Textarea } from "#components/ui/textarea";
import { useDatasetStore } from "#stores/useDatasetStore";

export function DatasetPicker() {
  const dataset = useDatasetStore((s) => s.dataset);
  const loadDataset = useDatasetStore((s) => s.loadDataset);
  const clearDataset = useDatasetStore((s) => s.clearDataset);
  const getValidationErrors = useDatasetStore((s) => s.getValidationErrors);
  const [rawInput, setRawInput] = useState("");
  const [fileName, setFileName] = useState<string | null>(null);

  const errors = getValidationErrors();

  const handleFile = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file) return;
      const text = await file.text();
      setRawInput(text);
      setFileName(file.name);
      loadDataset(text, file.name);
      e.target.value = "";
    },
    [loadDataset],
  );

  const handleTextChange = useCallback(
    (value: string) => {
      setRawInput(value);
      if (value.trim() === "") {
        clearDataset();
        setFileName(null);
        return;
      }
      // Use fileName hint if available, otherwise auto-detect via content
      loadDataset(value, fileName ?? "dataset.csv");
    },
    [loadDataset, clearDataset, fileName],
  );

  const handleClear = useCallback(() => {
    setRawInput("");
    setFileName(null);
    clearDataset();
  }, [clearDataset]);

  return (
    <div className="space-y-3" aria-label="Dataset picker">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium">Dataset (CSV or JSON Array)</span>
        {dataset && (
          <Button variant="ghost" size="sm" className="h-6 px-2 text-xs" onClick={handleClear} aria-label="Clear dataset">
            <X className="mr-1 size-3" />
            Clear
          </Button>
        )}
      </div>

      <div className="flex items-center gap-2">
        <label
          htmlFor="dataset-file"
          className="flex cursor-pointer items-center gap-1.5 rounded border border-border px-2 py-1 text-xs hover:bg-muted"
        >
          <Upload className="size-3" />
          Load file
          <Input id="dataset-file" type="file" accept=".csv,.json" className="hidden" onChange={handleFile} aria-label="Load dataset file" />
        </label>
        {fileName && (
          <span className="flex items-center gap-1 text-xs text-muted-foreground">
            <FileText className="size-3" />
            {fileName}
          </span>
        )}
        {dataset && (
          <span className="ml-auto text-xs text-muted-foreground">
            {dataset.rows.length} rows • {dataset.columns.length} cols • {dataset.source.toUpperCase()}
          </span>
        )}
      </div>

      <Textarea
        value={rawInput}
        onChange={(e) => handleTextChange(e.target.value)}
        placeholder="Paste CSV (id,name&#10;1,Alice) or JSON array ([{&quot;id&quot;:1}])…"
        rows={5}
        className="font-mono text-xs"
        aria-label="Dataset raw content"
      />

      {errors.length > 0 && (
        <div className="rounded border border-destructive/30 bg-destructive/10 p-2" role="alert">
          <p className="flex items-center gap-1 text-xs font-medium text-destructive">
            <AlertCircle className="size-3" />
            Validation
          </p>
          <ul className="mt-1 list-disc pl-4 text-xs text-destructive">
            {errors.map((err, i) => (
              <li key={i}>{err}</li>
            ))}
          </ul>
        </div>
      )}

      {dataset && dataset.rows.length > 0 && (
        <div className="overflow-auto rounded border border-border">
          <table className="w-full text-xs">
            <thead className="bg-muted/50">
              <tr>
                {dataset.columns.map((col) => (
                  <th key={col.name} className="px-2 py-1 text-left font-medium">
                    {col.name}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {dataset.rows.slice(0, 5).map((row) => (
                <tr key={row.index} className="border-t border-border">
                  {dataset.columns.map((col) => (
                    <td key={col.name} className="px-2 py-1 font-mono">
                      {row.values[col.name] ?? ""}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
          {dataset.rows.length > 5 && (
            <p className="border-t border-border bg-muted/20 px-2 py-1 text-center text-xs text-muted-foreground">
              …and {dataset.rows.length - 5} more rows
            </p>
          )}
        </div>
      )}
    </div>
  );
}
