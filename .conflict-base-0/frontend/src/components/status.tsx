import { cn } from "#lib/utils";

/** Tailwind keeps literal classes only — the ramp is enumerated, not interpolated. */
const STATUS_CLASSES = {
  info: "text-status-info bg-status-info/10 border-status-info/25",
  ok: "text-status-ok bg-status-ok/10 border-status-ok/25",
  redirect: "text-status-redirect bg-status-redirect/10 border-status-redirect/25",
  warn: "text-status-warn bg-status-warn/10 border-status-warn/25",
  error: "text-status-error bg-status-error/10 border-status-error/25",
} as const;

const DOT_CLASSES = {
  info: "bg-status-info",
  ok: "bg-status-ok",
  redirect: "bg-status-redirect",
  warn: "bg-status-warn",
  error: "bg-status-error",
} as const;

export type StatusTier = keyof typeof STATUS_CLASSES;

export function statusTier(status: number | null | undefined): StatusTier {
  if (status == null || status === 0) return "error";
  if (status < 200) return "info";
  if (status < 300) return "ok";
  if (status < 400) return "redirect";
  if (status < 500) return "warn";
  return "error";
}

/**
 * The one status device used everywhere — response header, run steps, history
 * rows. A colored dot plus tabular code so meaning never rides color alone.
 */
export function StatusPill({
  status,
  className,
}: {
  status: number | null | undefined;
  className?: string;
}) {
  const tier = statusTier(status);
  const label =
    status == null || status === 0 ? "ERR" : String(status);
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded border px-1.5 py-0.5 font-mono text-[11px] font-semibold tabular-nums leading-none tracking-tight select-none",
        STATUS_CLASSES[tier],
        className,
      )}
      title={
        status == null || status === 0
          ? "Request failed — no HTTP response"
          : `HTTP ${status}`
      }
    >
      <span aria-hidden className={cn("size-1.5 rounded-full", DOT_CLASSES[tier])} />
      {label}
    </span>
  );
}

const METHOD_CLASSES = {
  GET: "text-method-get",
  POST: "text-method-post",
  PUT: "text-method-put",
  PATCH: "text-method-put",
  DELETE: "text-method-delete",
  HEAD: "text-status-info",
  OPTIONS: "text-status-info",
} as const;

export function MethodLabel({
  method,
  className,
}: {
  method: string;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "font-mono text-[11px] font-semibold tracking-wide",
        // SAFETY: unknown methods miss the lookup and fall back to muted text
        METHOD_CLASSES[method.toUpperCase() as keyof typeof METHOD_CLASSES] ??
          "text-muted-foreground",
        className,
      )}
    >
      {method.toUpperCase()}
    </span>
  );
}
