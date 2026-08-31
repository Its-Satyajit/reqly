import {
	FileDown,
	FolderSearch,
	Search,
	Settings,
	SquareArrowOutDownLeft,
} from "lucide-react";
import logoDark from "../../assets/logo-dark.svg";
import logoLight from "../../assets/logo-light.svg";
import { Button } from "../ui/button";
import { ImportDialog, ExportDialog } from "../../features";
import {
	useCommandPaletteStore,
	useExportStore,
	useImportStore,
	useThemeStore,
	useWorkspaceStore,
} from "../../stores";
import { useWorkspaceBootstrapStore } from "../../stores/useWorkspaceBootstrap";
import { EnvironmentSelector } from "./EnvironmentSelector";

export function TopBar() {
	const resolvedTheme = useThemeStore((s) => s.resolvedTheme);
	const workspaceName = useWorkspaceStore((s) => s.workspaceTree?.name);
	const setImportOpen = useImportStore((s) => s.setOpen);
	const setExportOpen = useExportStore((s) => s.setOpen);
	const switchWorkspace = useWorkspaceBootstrapStore((s) => s.openFolder);
	const requestView = useWorkspaceStore((s) => s.requestView);

	return (
		<header className="flex h-10 shrink-0 items-center justify-between border-b border-border bg-background px-3 select-none">
			<div className="flex min-w-0 items-center gap-2">
				<img
					src={resolvedTheme === "atlas-dark" ? logoDark : logoLight}
					alt="Reqly"
					className="size-4 shrink-0"
				/>
				<button
					type="button"
					onClick={() => void switchWorkspace()}
					title="Switch workspace folder"
					className="group flex min-w-0 items-center gap-1.5 rounded px-1.5 py-0.5 text-xs transition-colors hover:bg-muted"
				>
					<span className="truncate font-mono font-medium text-foreground">
						{workspaceName ?? "Reqly"}
					</span>
					<FolderSearch className="size-3 shrink-0 text-muted-foreground transition-colors group-hover:text-foreground" aria-hidden />
				</button>
			</div>

			<div className="flex items-center gap-1.5">
				<button
					type="button"
					onClick={() => useCommandPaletteStore.getState().setOpen(true)}
					title="Search commands (⌘K)"
					aria-label="Search"
					className="flex h-7 items-center gap-2 rounded border border-input bg-card/60 px-2.5 text-xs text-muted-foreground transition-colors hover:border-border hover:bg-muted/80 hover:text-foreground"
				>
					<Search className="size-3.5" aria-hidden />
					<span className="hidden sm:inline text-[11px]">Search commands…</span>
					<kbd className="hidden sm:inline rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground border border-border/70">⌘K</kbd>
				</button>
				<Button
					variant="ghost"
					size="xs"
					onClick={() => setImportOpen(true)}
					title="Import from cURL, OpenAPI, HAR, Postman, Insomnia, or Bruno"
					className="gap-1 text-xs text-muted-foreground hover:text-foreground"
				>
					<SquareArrowOutDownLeft className="size-3.5" aria-hidden />
					<span>Import</span>
				</Button>
				<Button
					variant="ghost"
					size="xs"
					onClick={() => setExportOpen(true)}
					title="Export as Postman, OpenAPI, HAR, or a workspace copy"
					className="gap-1 text-xs text-muted-foreground hover:text-foreground"
				>
					<FileDown className="size-3.5" aria-hidden />
					<span>Export</span>
				</Button>
			</div>

			<div className="flex items-center gap-2">
				<EnvironmentSelector />
				<div className="hidden md:flex items-center gap-1.5 border-l border-border pl-2 pr-1 text-[11px] font-mono text-muted-foreground">
					<span className="size-1.5 rounded-full bg-status-ok" aria-hidden />
					<span>Local</span>
				</div>
				<Button
					variant="ghost"
					size="icon-sm"
					onClick={() => requestView("settings")}
					title="Settings"
					aria-label="Settings"
					className="text-muted-foreground hover:text-foreground"
				>
					<Settings className="size-3.5" aria-hidden />
				</Button>
			</div>

			<ImportDialog onImported={() => void useWorkspaceStore.getState().refreshEnvironments()} />
			<ExportDialog />
		</header>
	);
}
