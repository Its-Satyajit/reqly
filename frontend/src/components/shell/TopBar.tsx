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
		<header className="flex h-11 shrink-0 items-center gap-2 border-b border-border px-3">
			<div className="flex min-w-0 items-center gap-2">
				<img
					src={resolvedTheme === "atlas-dark" ? logoDark : logoLight}
					alt="Reqly"
					className="size-5 shrink-0"
				/>
				<button
					type="button"
					onClick={() => void switchWorkspace()}
					title="Switch workspace folder"
					className="flex min-w-0 items-center gap-1.5 rounded-md px-1.5 py-1 text-sm transition-colors hover:bg-muted"
				>
					<span className="truncate font-semibold">{workspaceName ?? "Reqly"}</span>
					<FolderSearch className="size-3.5 shrink-0 text-muted-foreground" aria-hidden />
				</button>
			</div>

			<div className="mx-auto flex items-center gap-1">
				<Button variant="ghost" size="sm" onClick={() => useCommandPaletteStore.getState().setOpen(true)} title="Search commands (⌘K)" aria-label="Search">
					<Search className="size-4" aria-hidden />
					<span className="hidden sm:inline">Search</span>
					<kbd className="hidden sm:inline ml-1 rounded bg-muted px-1 text-[11px]">⌘K</kbd>
				</Button>
				<Button
					variant="ghost"
					size="sm"
					onClick={() => setImportOpen(true)}
					title="Import from cURL, OpenAPI, HAR, Postman, Insomnia, or Bruno"
				>
					<SquareArrowOutDownLeft className="size-4" aria-hidden />
					Import
				</Button>
				<Button
					variant="ghost"
					size="sm"
					onClick={() => setExportOpen(true)}
					title="Export as Postman, OpenAPI, HAR, or a workspace copy"
				>
					<FileDown className="size-4" aria-hidden />
					Export
				</Button>
			</div>

			<div className="flex items-center gap-1.5">
				<EnvironmentSelector />
				<div className="hidden sm:flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] text-muted-foreground" title="Local-first Git storage (all files up-to-date)">
					<span className="size-1.5 rounded-full bg-status-ok" aria-hidden />
					<span>Saved</span>
				</div>
				<Button
					variant="ghost"
					size="sm"
					onClick={() => requestView("settings")}
					title="Settings"
					aria-label="Settings"
				>
					<Settings className="size-4" aria-hidden />
				</Button>
			</div>

			<ImportDialog onImported={() => void useWorkspaceStore.getState().refreshEnvironments()} />
			<ExportDialog />
		</header>
	);
}
