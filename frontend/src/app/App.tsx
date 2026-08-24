import { useEffect } from "react";
import logoDark from "../assets/logo-dark.svg";
import logoLight from "../assets/logo-light.svg";
import { ErrorBoundary } from "../components/ErrorBoundary";
import { CrashOverlay } from "../components/CrashOverlay";
import { CompactSelect } from "../components/CompactSelect";
import {
	ResizableHandle,
	ResizablePanel,
	ResizablePanelGroup,
} from "../components/ui/resizable";
import { AppShell } from "../components/shell/AppShell";
import { StatusBar } from "../components/shell/StatusBar";
import { RequestTabs, RunView, WorkspaceSidebar } from "../components";
import { RealtimeTab } from "../features/realtime-view/RealtimeTab";
import { TestTab } from "../features/test-runner/TestTab";
import { Toaster } from "../components/ui/toast";
import { EnvironmentsView } from "../features/environments-view/EnvironmentsView";
import { DiffView } from "../features/diff-view/DiffView";
import { GraphqlBrowser } from "../features/graphql-browser/GraphqlBrowser";
import { JwtInspector } from "../features/jwt-inspector/JwtInspector";
import { OpenapiExplorer } from "../features/openapi-explorer/OpenapiExplorer";
import { DocsView } from "../features/docs-view/DocsView";
import { GrpcTab } from "../features/grpc-view/GrpcTab";
import { RunnersPanel } from "../features/runners-panel/RunnersPanel";
import { MocksView } from "../features/mock-view/MocksView";
import { HistoryView } from "../features/history-view/HistoryView";
import { RequestEditor } from "../features/request-editor/RequestEditor";
import { ResponseViewer } from "../features/response-viewer/ResponseViewer";
import { useThemeStore, useWorkspaceBootstrapStore, useWorkspaceStore } from "../stores";
import { WorkspaceEmptyState } from "../features/workspace-bootstrap/WorkspaceEmptyState";
import { NEW_REQUEST_TAB_ID, tabIsDirty, useRequestStore } from "../stores/useRequestStore";
import { useDefaultLayout } from "react-resizable-panels";
import { armDebugCrashTrigger, installCrashReporter } from "../lib/crashReporter";
import { addBreadcrumb } from "../lib/crash";
import { notifyError, notifyWarning } from "../lib/notify";
import "../index.css";

const shellStorage: Pick<Storage, "getItem" | "setItem"> = {
	getItem: (key) => window.localStorage.getItem(key),
	setItem: (key, value) => window.localStorage.setItem(key, value),
};

export function App() {
	const theme = useThemeStore((s) => s.appearance);
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

	// ⌘/Ctrl+W closes the active tab (G-4.3.2). Dirty tabs are never silently
	// discarded — the user closes those via the tab's explicit ✕ confirm flow.
	useEffect(() => {
		const onKeyDown = (e: KeyboardEvent) => {
			if (!(e.metaKey || e.ctrlKey) || e.key.toLowerCase() !== "w") return;
			e.preventDefault();
			const { activeTabId, closeTab } = useWorkspaceStore.getState();
			if (!activeTabId) return;
			const req = useRequestStore.getState();
			if (tabIsDirty(req.drafts[activeTabId], req.meta[activeTabId])) {
				notifyWarning("Tab has unsaved changes", "Close it with the tab's ✕ to discard them.");
				return;
			}
			closeTab(activeTabId, { force: true });
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, []);

	useEffect(() => {
		void refreshEnvironments();
		void refreshWorkspace();
		const { openTabs, openTab } = useWorkspaceStore.getState();
		if (openTabs.length === 0) {
			openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" });
		}
	}, [refreshEnvironments, refreshWorkspace]);

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
			<AppShell
				brand={
					<>
						<img
							src={theme === "dark" ? logoDark : logoLight}
							alt="Reqly"
							className="size-6"
						/>
						<h1 className="text-sm font-semibold tracking-tight">Reqly</h1>
					</>
				}
				headerActions={
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
				}
				sidebar={
					<ErrorBoundary label="Sidebar">
						<WorkspaceSidebar />
					</ErrorBoundary>
				}
				statusbar={<StatusBar />}
			>
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
							) : activeView === "mocks" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Mock server">
										<MocksView />
									</ErrorBoundary>
								</section>
							) : activeView === "diff" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="API diff">
										<DiffView />
									</ErrorBoundary>
								</section>
							) : activeView === "jwt" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="JWT inspector">
										<JwtInspector />
									</ErrorBoundary>
								</section>
							) : activeView === "graphql" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="GraphQL browser">
										<GraphqlBrowser />
									</ErrorBoundary>
								</section>
							) : activeView === "runners" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Runners">
										<RunnersPanel />
									</ErrorBoundary>
								</section>
							) : activeView === "explorer" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="OpenAPI explorer">
										<OpenapiExplorer />
									</ErrorBoundary>
								</section>
							) : activeView === "grpc" ? (
								<section className="h-full min-h-0">
									<ErrorBoundary label="gRPC client">
										<GrpcTab tabId="grpc" />
									</ErrorBoundary>
								</section>
							) : activeView === "docs" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Docs generator">
										<DocsView />
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
								) : activeTab?.kind === "realtime" ? (
									<div className="min-h-0 min-w-0 flex-1">
										<ErrorBoundary label="Realtime client">
											<RealtimeTab tabId={activeTab.id} />
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
			</AppShell>
		</ErrorBoundary>
	);
}
