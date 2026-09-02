import { useEffect, useState } from "react";
import { CrashReportDialog } from "#components/CrashReportDialog";
import { subscribeToCrashes } from "#lib/crashReporter";

export function CrashOverlay() {
  const [report, setReport] = useState<string | null>(null);

  useEffect(
    () =>
      subscribeToCrashes((next) => {
        setReport(next);
      }),
    [],
  );

  return (
    <CrashReportDialog
      open={report != null}
      onOpenChange={(open) => {
        if (!open) setReport(null);
      }}
      report={report ?? ""}
      onReload={() => window.location.reload()}
    />
  );
}
