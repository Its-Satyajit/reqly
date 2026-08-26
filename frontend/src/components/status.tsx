import { cn } from "#lib/utils";
import {
	DOT_CLASSES,
	METHOD_CLASSES,
	STATUS_CLASSES,
	statusTier,
} from "../lib/status";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "#components/ui/tooltip";

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
		<Tooltip>
			<TooltipTrigger
				render={
					<span
						className={cn(
							"inline-flex items-center gap-1.5 rounded-full border px-2 py-px font-mono text-xs font-medium tabular-nums",
							STATUS_CLASSES[tier],
							className,
						)}
						aria-label={
							status == null || status === 0
								? "Request failed — no HTTP response"
								: `HTTP ${status}`
						}
					/>
				}
			>
				<span
					aria-hidden
					className={cn("size-1.5 rounded-full", DOT_CLASSES[tier])}
				/>
				{label}
			</TooltipTrigger>
			<TooltipContent>
				{status == null || status === 0
					? "Request failed — no HTTP response"
					: `HTTP ${status}`}
			</TooltipContent>
		</Tooltip>
	);
}

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
				"font-mono text-xs font-semibold tracking-wide",
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
