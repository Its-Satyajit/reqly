import {
	FileDown,
	FolderOpen,
	FolderPlus,
	History,
	RefreshCw,
	ChevronDown,
	Search,
	Settings,
	SquareArrowOutDownLeft,
	Trash2,
} from "lucide-react";
import logoDark from "../../assets/logo-dark.svg";
import logoLight from "../../assets/logo-light.svg";
import { Button } from "../ui/button";
import {
	Menubar,
	MenubarMenu,
	MenubarTrigger,
	MenubarContent,
	MenubarItem,
	MenubarSeparator,
	MenubarShortcut,
} from "../ui/menubar";
import { ImportDialog, ExportDialog } from "../../features";
import { CreateWorkspaceModal } from "../CreateWorkspaceModal";
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
	const currentWorkspace = useWorkspaceStore((s) => s.currentWorkspace);
	const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);
	const setImportOpen = useImportStore((s) => s.setOpen);
	const setExportOpen = useExportStore((s) => s.setOpen);
	const openFolder = useWorkspaceBootstrapStore((s) => s.openFolder);
	const openDirect = useWorkspaceBootstrapStore((s) => s.openDirect);
	const setCreateModalOpen = useWorkspaceBootstrapStore((s) => s.setCreateModalOpen);
	const recentWorkspaces = useWorkspaceBootstrapStore((s) => s.recentWorkspaces);
	const clearRecentWorkspaces = useWorkspaceBootstrapStore((s) => s.clearRecentWorkspaces);
	const requestView = useWorkspaceStore((s) => s.requestView);

	return (
		<header className="flex h-10 shrink-0 items-center justify-between border-b border-border bg-background px-3 select-none">
			<div className="flex min-w-0 items-center gap-2">
				<img
					src={resolvedTheme === "atlas-dark" ? logoDark : logoLight}
					alt="Reqly"
					className="size-4 shrink-0"
				/>

				<Menubar className="border-0 bg-transparent p-0 h-auto">
					<MenubarMenu>
						<MenubarTrigger className="group flex min-w-0 items-center gap-1.5 rounded px-2 py-1 text-xs font-mono font-medium text-foreground transition-colors hover:bg-muted data-popup-open:bg-muted cursor-pointer">
							<span className="truncate max-w-40 font-semibold">
								{workspaceName ?? "Reqly Workspace"}
							</span>
							<ChevronDown className="size-3 shrink-0 text-muted-foreground transition-transform group-hover:text-foreground" aria-hidden />
						</MenubarTrigger>

						<MenubarContent align="start" className="min-w-64">
							<div className="px-2 py-1.5 text-[11px] text-muted-foreground border-b border-border/50 mb-1">
								<div className="font-semibold text-foreground truncate">{workspaceName ?? "No workspace open"}</div>
								<div className="font-mono text-[10px] text-muted-foreground truncate">{currentWorkspace?.path ?? "—"}</div>
							</div>

							<MenubarItem onClick={() => void openFolder()}>
								<FolderOpen className="size-3.5" />
								<span>Open Folder…</span>
								<MenubarShortcut>⌘O</MenubarShortcut>
							</MenubarItem>

							<MenubarItem onClick={() => setCreateModalOpen(true)}>
								<FolderPlus className="size-3.5" />
								<span>Create Workspace…</span>
								<MenubarShortcut>⌘N</MenubarShortcut>
							</MenubarItem>

							<MenubarItem onClick={() => void refreshWorkspace()}>
								<RefreshCw className="size-3.5" />
								<span>Reload Workspace</span>
								<MenubarShortcut>⌘R</MenubarShortcut>
							</MenubarItem>

							<MenubarSeparator />

							<div className="px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground flex items-center justify-between">
								<span className="flex items-center gap-1.5">
									<History className="size-3" />
									Recent Workspaces
								</span>
								{recentWorkspaces.length > 0 && (
									<button
										type="button"
										onClick={(e) => {
											e.stopPropagation();
											clearRecentWorkspaces();
										}}
										className="text-[10px] lowercase text-muted-foreground hover:text-destructive flex items-center gap-0.5"
										title="Clear recent workspaces history"
									>
										<Trash2 className="size-2.5" />
										clear
									</button>
								)}
							</div>

							{recentWorkspaces.length === 0 ? (
								<div className="px-2 py-1.5 text-[11px] text-muted-foreground italic">
									No recent workspaces
								</div>
							) : (
								recentWorkspaces.map((ws) => (
									<MenubarItem
										key={ws.path}
										onClick={() => void openDirect(ws.path)}
										className="flex flex-col items-start gap-0.5 py-1.5"
									>
										<div className="flex w-full items-center justify-between">
											<span className="font-medium text-foreground truncate">{ws.name}</span>
											{currentWorkspace?.path === ws.path && (
												<span className="text-[9px] bg-primary/20 text-primary px-1 rounded">active</span>
											)}
										</div>
										<span className="text-[10px] text-muted-foreground font-mono truncate w-full">
											{ws.path}
										</span>
									</MenubarItem>
								))
							)}
						</MenubarContent>
					</MenubarMenu>
				</Menubar>
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
			<CreateWorkspaceModal />
		</header>
	);
}
