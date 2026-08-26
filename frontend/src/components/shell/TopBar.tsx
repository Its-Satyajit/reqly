import {
	FileDown,
	FolderSearch,
	PanelLeftClose,
	PanelLeftOpen,
	SquareArrowOutDownLeft,
} from "lucide-react";
import logoDark from "../../assets/logo-dark.svg";
import logoLight from "../../assets/logo-light.svg";
import { Button } from "../ui/button";
import { CompactSelect } from "../CompactSelect";
import { ImportDialog, ExportDialog } from "../../features";
import {
	useExportStore,
	useImportStore,
	useThemeStore,
	useWorkspaceStore,
} from "../../stores";
import { useWorkspaceBootstrapStore } from "../../stores/useWorkspaceBootstrap";
import { addBreadcrumb } from "../../lib/crash";
import { notifyError } from "../../lib/notify";

interface TopBarProps {
	sidebarCollapsed: boolean;
	onToggleSidebar: () => void;
}

export function TopBar({ sidebarCollapsed, onToggleSidebar }: TopBarProps) {
	const resolvedTheme = useThemeStore((s) => s.resolvedTheme);
	const workspaceName = useWorkspaceStore((s) => s.workspaceTree?.name);
	const environments = useWorkspaceStore((s) => s.environments);
	const environmentsError = useWorkspaceStore((s) => s.environmentsError);
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);
	const refreshEnvironments = useWorkspaceStore((s) => s.refreshEnvironments);
	const setActiveEnvironment = useWorkspaceStore((s) => s.setActiveEnvironment);
	const setImportOpen = useImportStore((s) => s.setOpen);
	const setExportOpen = useExportStore((s) => s.setOpen);
	const switchWorkspace = useWorkspaceBootstrapStore((s) => s.openFolder);

	const onSelectEnvironment = async (name: string) => {
		addBreadcrumb("env-switch", name || "none");
		const envAdapter = useWorkspaceStore.getState().envAdapter;
		setActiveEnvironment(name || null);
		try {
			await envAdapter.setActive(name);
		} catch (err) {
			notifyError(
				"Could not save the active environment",
				err instanceof Error ? err.message : String(err),
			);
			await refreshEnvironments();
			return;
		}
		await refreshEnvironments();
	};

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

			<div className="flex items-center gap-2">
				<CompactSelect
					value={activeEnvironmentId ?? ""}
					onChange={(next) => void onSelectEnvironment(next)}
					ariaLabel={
						environmentsError ?? "Select the active environment"
					}
					options={[
						{ value: "", label: "No environment" },
						...environments.map((env) => ({
							value: env.id,
							label: env.name,
						})),
					]}
				/>
				<Button
					variant="ghost"
					size="icon-sm"
					onClick={onToggleSidebar}
					aria-label={sidebarCollapsed ? "Show sidebar" : "Hide sidebar"}
					aria-pressed={!sidebarCollapsed}
					title={sidebarCollapsed ? "Show sidebar (Ctrl+B)" : "Hide sidebar (Ctrl+B)"}
				>
					{sidebarCollapsed ? (
						<PanelLeftOpen className="size-4" aria-hidden />
					) : (
						<PanelLeftClose className="size-4" aria-hidden />
					)}
				</Button>
			</div>

			<ImportDialog onImported={() => void refreshEnvironments()} />
			<ExportDialog />
		</header>
	);
}
