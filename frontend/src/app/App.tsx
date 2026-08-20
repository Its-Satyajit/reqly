import { useEffect } from "react";
import logoDark from "../assets/logo-dark.svg";
import logoLight from "../assets/logo-light.svg";
import { RequestTabs, RunView, ThemeToggle, WorkspaceSidebar } from "../components";
import { EnvironmentsView } from "../features/environments-view/EnvironmentsView";
import { RequestEditor } from "../features/request-editor/RequestEditor";
import { ResponseViewer } from "../features/response-viewer/ResponseViewer";
import { useThemeStore, useWorkspaceStore } from "../stores";
import { NEW_REQUEST_TAB_ID } from "../stores/useRequestStore";
import "../index.css"

export function App() {
	const theme = useThemeStore((s) => s.theme);
	const activeView = useWorkspaceStore((s) => s.activeView);
	const environments = useWorkspaceStore((s) => s.environments);
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);
	const environmentsError = useWorkspaceStore((s) => s.environmentsError);
	const setActiveEnvironment = useWorkspaceStore((s) => s.setActiveEnvironment);
	const refreshEnvironments = useWorkspaceStore((s) => s.refreshEnvironments);
	const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);
	const openTabs = useWorkspaceStore((s) => s.openTabs);
	const activeTabId = useWorkspaceStore((s) => s.activeTabId);

	const activeEnvironment = environments.find(
		(e) => e.id === activeEnvironmentId,
	);
	const activeTab = openTabs.find((t) => t.id === activeTabId);

	useEffect(() => {
		void refreshEnvironments();
		void refreshWorkspace();
		// Ensure the default scratchpad tab exists on first load.
		const { openTabs, openTab } = useWorkspaceStore.getState();
		if (openTabs.length === 0) {
			openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" });
		}
	}, [refreshEnvironments, refreshWorkspace]);

	const onSelectEnvironment = async (name: string) => {
		const envAdapter = useWorkspaceStore.getState().envAdapter;
		setActiveEnvironment(name || null);
		try {
			await envAdapter.setActive(name);
		} catch {
			// Persist failed: re-sync from the source of truth.
			await refreshEnvironments();
			return;
		}
		await refreshEnvironments();
	};

	return (
		<div className="flex h-screen flex-col">
			<header className="flex h-12 shrink-0 items-center justify-between border-b border-border px-4">
				<div className="flex items-center gap-2">
					<img
						src={theme === "dark" ? logoDark : logoLight}
						alt="Reqly"
						className="size-6"
					/>
					<h1 className="text-sm font-semibold">Reqly</h1>
				</div>
				<div className="flex items-center gap-2">
					<select
						value={activeEnvironment?.id ?? ""}
						onChange={(e) => void onSelectEnvironment(e.target.value)}
						title={environmentsError ?? "Select the active environment"}
						className="rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground"
					>
						<option value="">No environment</option>
						{environments.map((env) => (
							<option key={env.id} value={env.id}>
								{env.name}
							</option>
						))}
					</select>
					<ThemeToggle />
				</div>
			</header>
			<div className="flex min-h-0 flex-1">
				<WorkspaceSidebar />
				<main className="flex min-w-0 flex-1 flex-col lg:flex-row">
					{activeView === "environments" ? (
						<section className="min-h-0 flex-1 overflow-y-auto">
							<EnvironmentsView />
						</section>
					) : (
						<section className="flex min-h-0 flex-1 flex-col">
								<RequestTabs />
								{activeTab?.kind === "run" ? (
									<div className="min-h-0 flex-1">
										<RunView />
									</div>
								) : (
									<div className="flex min-h-0 flex-1 flex-col lg:flex-row">
										<div className="min-h-0 flex-1 border-r border-border lg:w-1/2">
											<RequestEditor />
										</div>
										<div className="min-h-0 flex-1 lg:w-1/2">
											<ResponseViewer />
										</div>
									</div>
								)}
							</section>
					)}
				</main>
			</div>
		</div>
	);
}
