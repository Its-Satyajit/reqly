import { GitBranch, GitCommitHorizontal, Circle } from "lucide-react";
import { cn } from "#lib/utils";
import { useWorkspaceStore } from "../../stores";

export function StatusBar() {
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);
	const environments = useWorkspaceStore((s) => s.environments);

	const activeEnvironment = environments.find(
		(env) => env.id === activeEnvironmentId,
	);

	return (
		<footer className="flex h-6 shrink-0 items-center gap-3 border-t border-border bg-card/40 px-3 text-xs text-muted-foreground">
			<div className="flex items-center gap-1.5">
				<GitBranch className="size-3" aria-hidden />
				<span>main</span>
			</div>
			<div className="flex items-center gap-1.5">
				<GitCommitHorizontal className="size-3" aria-hidden />
				<span>0</span>
			</div>
			<div className="flex items-center gap-1.5">
				<Circle
					className={cn("size-2", "fill-muted-foreground")}
					aria-hidden
				/>
			</div>
			{activeEnvironment && (
				<div className="flex items-center gap-1.5 ml-auto">
					<span className="font-medium">{activeEnvironment.name}</span>
				</div>
			)}
		</footer>
	);
}
