import { Check, FolderOpen, History, Import, Settings, FolderPlus } from "lucide-react";
import { Menu } from "@base-ui/react/menu";

import { useWorkspaceBootstrapStore } from "#stores/useWorkspaceBootstrap";
import { useWorkspaceStore } from "#stores";
import { useHistoryStore } from "#stores/useHistoryStore";
import { notifyError, notifySuccess } from "#lib/notify";
import { readRecents } from "#lib/workspace";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "#components/ui/tooltip";

const itemClass =
	"flex w-full cursor-default items-center gap-2 rounded-sm px-2 py-1 text-left outline-hidden select-none data-highlighted:bg-muted";
const iconClass = "size-3.5 shrink-0 text-muted-foreground";

/**
 * VS Code-style workspace menu on the header brand chip: recent workspaces
 * with a check on the active one, open/create entry points, and workspace
 *-level actions (settings, import, clear history).
 *
 * Clone Git Repository and Show in File Manager need backend support that
 * does not exist yet (no WorkspaceClone / WorkspaceReveal in AppService), so
 * they are intentionally absent.
 */
export function WorkspaceMenu({ children }: { children: React.ReactNode }) {
	const status = useWorkspaceBootstrapStore((s) => s.status);
	const openFolder = useWorkspaceBootstrapStore((s) => s.openFolder);
	const busy = useWorkspaceBootstrapStore((s) => s.busy);
	const bootstrapAdapter = useWorkspaceBootstrapStore((s) => s.adapter);
	const setActiveView = useWorkspaceStore((s) => s.setActiveView);
	const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);
	const refreshEnvironments = useWorkspaceStore((s) => s.refreshEnvironments);
	const historyClear = useHistoryStore((s) => s.clear);
	const recents = readRecents();
	const current = status?.path ?? "";

	const openRecent = async (dir: string) => {
		try {
			await bootstrapAdapter.open(dir);
			await Promise.all([refreshWorkspace(), refreshEnvironments()]);
		} catch (err) {
			notifyError(
				"Could not open workspace",
				err instanceof Error ? err.message : String(err),
			);
		}
	};

	const clearHistory = async () => {
		try {
			await historyClear(null);
			notifySuccess("Send history cleared");
		} catch (err) {
			notifyError(
				"Could not clear history",
				err instanceof Error ? err.message : String(err),
			);
		}
	};

	return (
		<Menu.Root>
			<Menu.Trigger render={<div>{children}</div>} />
			<Menu.Portal>
				<Menu.Positioner align="start" sideOffset={6} className="z-(--z-overlay)">
					<Menu.Popup className="min-w-56 rounded-md border border-border bg-popover p-1 text-xs text-popover-foreground shadow-lg ring-1 ring-foreground/10 outline-none">
						{recents.length > 0 && (
							<>
								<Menu.GroupLabel className="px-2 pt-1 pb-0.5 text-2xs tracking-widest text-muted-foreground uppercase">
									Recent workspaces
								</Menu.GroupLabel>
								{recents.map((dir) => (
									<Menu.Item
										key={dir}
										disabled={busy}
										className={itemClass}
										onClick={() => void openRecent(dir)}
									>
										<span className="flex size-3.5 shrink-0 items-center justify-center">
											{dir === current && (
												<Check className="size-3 text-status-ok" aria-hidden />
											)}
										</span>
										<span className="truncate" title={dir}>
											{dir.split(/[\\/]/).filter(Boolean).pop() ?? dir}
										</span>
									</Menu.Item>
								))}
								<Menu.Separator className="-mx-1 my-1 h-px bg-border" />
							</>
						)}
						<Menu.Item disabled={busy} className={itemClass} onClick={() => void openFolder()}>
							<FolderOpen className={iconClass} aria-hidden />
							Open folder…
						</Menu.Item>
						<Menu.Item disabled={busy} className={itemClass} onClick={() => void openFolder()}>
							<FolderPlus className={iconClass} aria-hidden />
							Create workspace…
						</Menu.Item>
						<Menu.Item className={itemClass} onClick={() => setActiveView("importexport")}>
							<Import className={iconClass} aria-hidden />
							Import data
						</Menu.Item>
						<Menu.Separator className="-mx-1 my-1 h-px bg-border" />
						<Menu.Item className={itemClass} onClick={() => setActiveView("settings")}>
							<Settings className={iconClass} aria-hidden />
							Workspace settings
						</Menu.Item>
						<Tooltip>
							<TooltipTrigger
								render={
									<Menu.Item
										className={itemClass}
										onClick={() => void clearHistory()}
									/>
								}
							>
								<History className={iconClass} aria-hidden />
								Clear send history
							</TooltipTrigger>
							<TooltipContent side="right">
								Deletes every recorded request/response locally
							</TooltipContent>
						</Tooltip>
					</Menu.Popup>
				</Menu.Positioner>
			</Menu.Portal>
		</Menu.Root>
	);
}
