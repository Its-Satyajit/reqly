import { useEffect } from "react";
import { GitBranch, Leaf, ShieldCheck, Terminal } from "lucide-react";
import { workspaceViewLabel } from "#lib/views";
import { type WorkspaceFolder, type WorkspaceTree } from "#lib/collections";
import { cn } from "#lib/utils";
import { useGitStore, usePaletteStore, useShellStore, useWorkspaceStore } from "#stores";

/** countTreeRequests totals requests across every collection/folder. */
function countTreeRequests(tree: WorkspaceTree | null): number {
	if (!tree) return 0;
	let count = 0;
	const walkFolder = (f: WorkspaceFolder) => {
		count += f.requests.length;
		for (const child of f.folders) walkFolder(child);
	};
	for (const c of tree.collections) {
		count += c.requests.length;
		for (const f of c.folders) walkFolder(f);
	}
	return count;
}

/** Shell statusbar (G-17.3.5): branch / environment / request count /
 * zero-telemetry on the left, console-style slot on the right. */
export function StatusBar() {
	const activeView = useWorkspaceStore((s) => s.activeView);
	const environments = useWorkspaceStore((s) => s.environments);
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);
	const workspaceTree = useWorkspaceStore((s) => s.workspaceTree);
	const branch = useGitStore((s) => s.status?.branch ?? "");
	const repoFound = useGitStore((s) => s.status?.repoFound ?? false);
	const gitLoading = useGitStore((s) => s.loading);
	const refreshGit = useGitStore((s) => s.refresh);
	const inspectorOpen = useShellStore((s) => s.inspectorOpen);
	const toggleInspector = useShellStore((s) => s.toggleInspector);
	const paletteOpen = usePaletteStore((s) => s.open);

	useEffect(() => {
		if (!gitLoading && repoFound === false) void refreshGit();
	}, [gitLoading, repoFound, refreshGit]);

	const env = environments.find((e) => e.id === activeEnvironmentId);
	const requestCount = countTreeRequests(workspaceTree);

	return (
		<>
			<span className="flex min-w-0 items-center gap-3">
				<span className="flex items-center gap-1.5">
					<GitBranch className="size-3 shrink-0" aria-hidden />
					<span className="font-data">{branch || workspaceViewLabel(activeView)}</span>
				</span>
				<span className="flex items-center gap-1.5">
					<Leaf className="size-3 shrink-0 text-status-ok" aria-hidden />
					<span>{env ? env.name : "no environment"}</span>
				</span>
				<span className="font-data tabular-nums">{requestCount} requests</span>
				<span className="hidden items-center gap-1.5 text-muted-foreground/70 sm:flex">
					<ShieldCheck className="size-3 shrink-0 text-status-ok" aria-hidden />
					Zero telemetry · local-first
				</span>
				{paletteOpen ? <span className="font-data">· Palette</span> : null}
			</span>
			<span className="flex items-center gap-1 rounded-md border border-border bg-muted/30 px-1 py-0.5 font-data text-[11px]">
				<button
					type="button"
					onClick={toggleInspector}
					className={cn(
						"flex items-center gap-1 rounded px-1.5 py-px hover:bg-muted",
						inspectorOpen ? "text-foreground" : "text-muted-foreground",
					)}
				>
					<Terminal className="size-3" aria-hidden />
					Console
				</button>
				<span className="px-1 text-muted-foreground/60">Test output</span>
				<span className="px-1 text-muted-foreground/60">Network</span>
			</span>
		</>
	);
}
