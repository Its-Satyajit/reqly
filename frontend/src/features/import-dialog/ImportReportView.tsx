import { cn } from "#lib/utils";
import { Badge } from "#components/ui/badge";
import type { ImportReport, ImportReportEntry } from "#lib/import";

const CATEGORY_ORDER = ["auth", "script", "body", "environment", "schema", "other"];

const CATEGORY_LABELS = new Map<string, string>([
  ["auth", "Auth"],
  ["script", "Scripts"],
  ["body", "Bodies"],
  ["environment", "Environments"],
  ["schema", "Schema"],
  ["other", "Other"],
]);

interface SeverityBadge {
  label: string;
  className?: string;
  variant?: "secondary" | "destructive" | "outline";
}

const SEVERITY_BADGES = new Map<string, SeverityBadge>([
  ["translated", { label: "translated", variant: "secondary" }],
  [
    "warned",
    {
      label: "warned",
      variant: "outline",
      className: "border-status-warn/40 text-status-warn",
    },
  ],
  ["dropped", { label: "dropped", variant: "destructive" }],
]);

function severityTally(entries: ImportReportEntry[]) {
  const tally = new Map<string, number>();
  for (const entry of entries) {
    tally.set(entry.severity, (tally.get(entry.severity) ?? 0) + 1);
  }
  return tally;

}

/** Groups report entries by category in roadmap order. Module-level so the
 * component carries no manual memoization (React Compiler manages it). */
function groupByCategory(
  report: ImportReport | null | undefined,
): [string, ImportReportEntry[]][] {
  if (!report) return [];
  const byCategory = new Map<string, ImportReportEntry[]>();
  for (const entry of report.entries ?? []) {
    const list = byCategory.get(entry.category) ?? [];
    list.push(entry);
    byCategory.set(entry.category, list);
  }
  return [...byCategory.entries()].sort(
    ([a], [b]) =>
      (CATEGORY_ORDER.indexOf(a) === -1 ? CATEGORY_ORDER.length : CATEGORY_ORDER.indexOf(a)) -
      (CATEGORY_ORDER.indexOf(b) === -1 ? CATEGORY_ORDER.length : CATEGORY_ORDER.indexOf(b)),
  );
}

export function ImportReportView({ report }: { report: ImportReport | null | undefined }) {
  const groups = groupByCategory(report);

  if (!report || (report.entries?.length ?? 0) === 0) {
    return (
      <div className="rounded-md border border-border px-3 py-2 text-xs text-muted-foreground">
        Everything in this source mapped cleanly — nothing was skipped.
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2">
      {groups.map(([category, entries]) => {
        const tally = severityTally(entries);
        return (
          <div key={category} className="rounded-md border border-border">
            <div className="flex items-center gap-2 border-b border-border px-3 py-1.5">
              <span className="text-xs font-medium">
                {CATEGORY_LABELS.get(category) ?? category}
              </span>
              <span className="ml-auto flex items-center gap-1">
                {[...tally.entries()].flatMap(([severity, count]) => {
                  if (count === 0) return [];
                  const badge: SeverityBadge = SEVERITY_BADGES.get(severity) ?? {
                    label: severity,
                  };
                  return [
                    <Badge
                      key={severity}
                      variant={badge.variant ?? "outline"}
                      className={cn("tabular-nums", badge.className)}
                    >
                      {count} {badge.label}
                    </Badge>,
                  ];
                })}
              </span>
            </div>
            <ul className="flex flex-col divide-y divide-border/60">
              {entries.map((entry) => (
                <li
                  key={`${entry.severity}-${entry.itemPath}-${entry.message}`}
                  className="px-3 py-1.5 text-xs"
                >
                  {entry.itemPath !== "" && (
                    <span className="font-mono text-xs text-muted-foreground">
                      {entry.itemPath}
                      {" · "}
                    </span>
                  )}
                  {entry.message}
                </li>
              ))}
            </ul>
          </div>
        );
      })}
    </div>
  );
}
