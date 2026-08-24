import { useEffect, useState } from "react";
import { PanelLeftClose, PanelLeftOpen } from "lucide-react";
import logoDark from "../assets/logo-dark.svg";
import logoLight from "../assets/logo-light.svg";
import { ErrorBoundary } from "../components/ErrorBoundary";
import { CrashOverlay } from "../components/CrashOverlay";
import { CompactSelect } from "../components/CompactSelect";
import { Button } from "../components/ui/button";
import {
	ResizableHandle,
	ResizablePanel,
	ResizablePanelGroup,
} from "../components/ui/resizable";
import { RequestTabs, RunView, ThemeToggle, WorkspaceSidebar } from "../components";
import { TestTab } from "../features/test-runner/TestTab";
import { Toaster } from "../components/ui/toast";
import { EnvironmentsView } from "../features/environments-view/EnvironmentsView";
import { HistoryView } from "../features/history-view/HistoryView";
import { RequestEditor } from "../features/request-editor/RequestEditor";
import { ResponseViewer } from "../features/response-viewer/ResponseViewer";
import { useThemeStore, useWorkspaceBootstrapStore, useWorkspaceStore } from "../stores";
import { WorkspaceEmptyState } from "../features/workspace-bootstrap/WorkspaceEmptyState";
import { NEW_REQUEST_TAB_ID } from "../stores/useRequestStore";
import { useDefaultLayout, usePanelRef } from "react-resizable-panels";
import { armDebugCrashTrigger, installCrashReporter } from "../lib/crashReporter";
import { addBreadcrumb } from "../lib/crash";
import { notifyError } from "../lib/notify";
import "../index.css";

const shellStorage: Pick<Storage, "getItem" | "setItem"> = {
	getItem: (key) => window.localStorage.getItem(key),
	setItem: (key, value) => window.localStorage.setItem(key, value),
};

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

	const bootChecked = useWorkspaceBootstrapStore((s) => s.checked);
	const bootFound = useWorkspaceBootstrapStore((s) => s.status?.found ?? false);
	const initBootstrap = useWorkspaceBootstrapStore((s) => s.init);

	useEffect(() => {
		installCrashReporter();
		return armDebugCrashTrigger();
	}, []);

	useEffect(() => {
		void initBootstrap();
	}, [initBootstrap]);

	useEffect(() => {
		void refreshEnvironments();
		void refreshWorkspace();
		const { openTabs, openTab } = useWorkspaceStore.getState();
		if (openTabs.length === 0) {
			openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" });
		}
	}, [refreshEnvironments, refreshWorkspace]);

	const sidebarLayout = useDefaultLayout({
		id: "reqly-shell-sidebar",
		storage: shellStorage,
	});
	const sidebarPanel = usePanelRef();
	const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
	const toggleSidebar = () => {
		if (sidebarCollapsed) {
			sidebarPanel.current?.expand();
			return;
		}
		sidebarPanel.current?.collapse();
	};
	const splitLayout = useDefaultLayout({
		id: "reqly-shell-split",
		storage: shellStorage,
	});

	const activeEnvironment = environments.find(
		(e) => e.id === activeEnvironmentId,
	);
	const activeTab = openTabs.find((t) => t.id === activeTabId);

	if (!bootChecked) {
		return (
			<ErrorBoundary variant="root">
				<div className="flex min-h-screen items-center justify-center bg-background">
					<img src={theme === "dark" ? logoDark : logoLight} alt="Reqly" className="h-8 w-auto opacity-70" />
				</div>
			</ErrorBoundary>
		);
	}
	if (!bootFound) {
		return (
			<ErrorBoundary variant="root">
				<WorkspaceEmptyState />
			</ErrorBoundary>
		);
	}

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
		<ErrorBoundary variant="root">
			<Toaster />
			<CrashOverlay />
			<div className="flex h-screen flex-col overflow-hidden">
				<header className="flex h-11 shrink-0 items-center justify-between border-b border-border px-3">
					<div className="flex items-center gap-2">
						<img
							src={theme === "dark" ? logoDark : logoLight}
							alt="Reqly"
							className="size-6"
						/>
						<h1 className="text-sm font-semibold tracking-tight">Reqly</h1>
					</div>
					<div className="flex items-center gap-2">
						<Button
							variant="ghost"
							size="icon-sm"
							onClick={toggleSidebar}
							aria-label={sidebarCollapsed ? "Show sidebar" : "Hide sidebar"}
							aria-pressed={!sidebarCollapsed}
							title={sidebarCollapsed ? "Show sidebar" : "Hide sidebar"}
						>
							{sidebarCollapsed ? (
								<PanelLeftOpen className="size-4" aria-hidden />
							) : (
								<PanelLeftClose className="size-4" aria-hidden />
							)}
						</Button>
						<CompactSelect
							value={activeEnvironment?.id ?? ""}
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
						<ThemeToggle />
					</div>
				</header>
				<div className="min-h-0 flex-1">
					<ResizablePanelGroup
						orientation="horizontal"
						defaultLayout={sidebarLayout.defaultLayout}
						onLayoutChanged={sidebarLayout.onLayoutChanged}
					>
						<ResizablePanel
							id="sidebar"
							panelRef={sidebarPanel}
							collapsible
							collapsedSize={0}
							defaultSize="17%"
							minSize={168}
							maxSize="42%"
							onResize={(size) => setSidebarCollapsed(size.inPixels <= 1)}
						>
							<ErrorBoundary label="Sidebar">
								<WorkspaceSidebar />
							</ErrorBoundary>
						</ResizablePanel>
						<ResizableHandle />
						<ResizablePanel id="main" minSize="35%">
							{activeView === "environments" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Environments">
										<EnvironmentsView />
									</ErrorBoundary>
								</section>
							) : activeView === "history" ? (
								<section className="h-full min-h-0">
									<ErrorBoundary label="History">
										<HistoryView />
									</ErrorBoundary>
								</section>
							) : (
								<section className="flex h-full min-h-0 flex-col">
									<RequestTabs />
								{activeTab?.kind === "test" ? (
									<div className="min-h-0 min-w-0 flex-1">
										<ErrorBoundary label="Test runner">
											<TestTab tabId={activeTab.id} />
										</ErrorBoundary>
									</div>
								) : activeTab?.kind === "run" ? (
										<div className="min-h-0 min-w-0 flex-1">
											<ErrorBoundary label="Run view">
												<RunView />
											</ErrorBoundary>
										</div>
									) : (
										<div className="min-h-0 min-w-0 flex-1">
											<ResizablePanelGroup
												orientation="horizontal"
												defaultLayout={splitLayout.defaultLayout}
												onLayoutChanged={splitLayout.onLayoutChanged}
											>
												<ResizablePanel
													id="editor"
													defaultSize="50%"
													minSize="25%"
												>
													<div className="h-full min-h-0 min-w-0 border-r border-border">
														<ErrorBoundary label="Request editor">
															<RequestEditor />
														</ErrorBoundary>
													</div>
												</ResizablePanel>
												<ResizableHandle />
												<ResizablePanel
													id="viewer"
													defaultSize="50%"
													minSize="25%"
												>
													<div className="h-full min-h-0 min-w-0">
														<ErrorBoundary label="Response viewer">
															<ResponseViewer />
														</ErrorBoundary>
													</div>
												</ResizablePanel>
											</ResizablePanelGroup>
										</div>
									)}
								</section>
							)}
						</ResizablePanel>
					</ResizablePanelGroup>
				</div>
			</div>
		</ErrorBoundary>
	);
}
