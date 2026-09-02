import { GitBranch, GitCommitHorizontal, Circle } from "lucide-react";
import { cn } from "#lib/utils";
import { useWorkspaceStore } from "../../stores/useWorkspaceStore";

export function StatusBar() {
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);
	const environments = useWorkspaceStore((s) => s.environments);

	// TODO: Wire to ShellAdapter for real git status (branch, ahead/behind, dirty)
	const gitBranch = "main";
	const gitAheadBehind = 0;
	const gitDirty = false;

	const activeEnvironment = environments.find(
		(env) => env.id === activeEnvironmentId,
	);

	return (
		<footer className="flex h-6 shrink-0 items-center justify-between border-t border-border bg-card/25 px-3 font-mono text-[11px] text-muted-foreground select-none">
			<div className="flex items-center gap-3">
				<div className="flex items-center gap-1.5" title={`Git branch: ${gitBranch}`}>
					<GitBranch className="size-3 text-muted-foreground" aria-hidden />
					<span className="text-foreground/80">{gitBranch}</span>
				</div>
				{gitAheadBehind > 0 && (
					<div className="flex items-center gap-1">
						<GitCommitHorizontal className="size-3" aria-hidden />
						<span className="tabular-nums">{gitAheadBehind}</span>
					</div>
				)}
				<div className="flex items-center gap-1">
					<Circle
						className={cn(
							"size-1.5",
							gitDirty ? "fill-status-warn text-status-warn" : "fill-status-ok text-status-ok",
						)}
						aria-label={gitDirty ? "Dirty working tree" : "Clean working tree"}
						aria-hidden
					/>
					<span className="text-[10px] uppercase tracking-wide">
						{gitDirty ? "dirty" : "clean"}
					</span>
				</div>
			</div>
			<div className="flex items-center gap-2">
				{activeEnvironment ? (
					<div className="flex items-center gap-1.5">
						<span className="text-[10px] uppercase tracking-wide text-muted-foreground">env:</span>
						<span className="font-semibold text-primary">{activeEnvironment.name}</span>
					</div>
				) : (
					<span className="text-[10px] text-muted-foreground">no active env</span>
				)}
			</div>
		</footer>
	);
}
