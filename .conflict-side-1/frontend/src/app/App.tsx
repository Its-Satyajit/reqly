import { GitBranch } from "lucide-react";
import { useEffect } from "react";
import logoDark from "../assets/logo-dark.svg";
import logoLight from "../assets/logo-light.svg";
import { ErrorBoundary } from "../components/ErrorBoundary";
import { CrashOverlay } from "../components/CrashOverlay";
import { cn } from "#lib/utils";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "../components/ui/select";
import {
	ResizableHandle,
	ResizablePanel,
	ResizablePanelGroup,
} from "../components/ui/resizable";
import { ActivityRail } from "../components/shell/ActivityRail";
import { AppShell } from "../components/shell/AppShell";
import { ResponseModeToggle } from "../components/shell/ResponseModeToggle";
import { CommitStrip } from "../components/GitPanel";
import {
	CommandPalette,
	PaletteTriggerButton,
} from "../components/palette/CommandPalette";
import { StatusBar } from "../components/shell/StatusBar";
import { WorkspaceMenu } from "../components/shell/WorkspaceMenu";
import { shellStorage } from "../components/shell/storage";
import { RequestTabs } from "../components/RequestTabs";
import { RunView } from "../components/RunView";
import { WorkspaceSidebar } from "../components/WorkspaceSidebar";
import { RealtimeTab } from "../features/realtime-view/RealtimeTab";
import { TestTab } from "../features/test-runner/TestTab";
import { Toaster } from "#components/ui/sonner";
import { EnvironmentsView } from "../features/environments-view/EnvironmentsView";
import { DiffView } from "../features/diff-view/DiffView";
import { GraphqlBrowser } from "../features/graphql-browser/GraphqlBrowser";
import { JwtInspector } from "../features/jwt-inspector/JwtInspector";
import { OpenapiExplorer } from "../features/openapi-explorer/OpenapiExplorer";
import { DocsView } from "../features/docs-view/DocsView";
import { GrpcTab } from "../features/grpc-view/GrpcTab";
import { RunnersPanel } from "../features/runners-panel/RunnersPanel";
import { SettingsView } from "../features/settings-view/SettingsView";
import { MocksView } from "../features/mock-view/MocksView";
import { HistoryView } from "../features/history-view/HistoryView";
import { RequestEditor } from "../features/request-editor/RequestEditor";
import { ResponseViewer } from "../features/response-viewer/ResponseViewer";
import { useGitStore } from "../stores/useGitStore";
import { useShellStore } from "../stores/useShellStore";
import { useThemeStore } from "../stores/useThemeStore";
import { useWorkspaceBootstrapStore } from "../stores/useWorkspaceBootstrap";
import { useWorkspaceStore } from "../stores/useWorkspaceStore";
import { WorkspaceEmptyState } from "../features/workspace-bootstrap/WorkspaceEmptyState";
import { WelcomeModal } from "../features/workspace-bootstrap/WelcomeModal";
import { OverviewHome } from "../features/overview-home/OverviewHome";
import { TestsView } from "../features/tests-view/TestsView";
import { RealtimeView } from "../features/realtime-view/RealtimeView";
import { ImportExportView } from "../features/import-export-view/ImportExportView";
import { CodegenView } from "../features/codegen-view/CodegenView";
import { AuthPanel } from "../features/auth-panel/AuthPanel";
import { GitPanel } from "../components/GitPanel";
import { NEW_REQUEST_TAB_ID, tabIsDirty, useRequestStore } from "../stores/useRequestStore";
import { useDefaultLayout } from "react-resizable-panels";
import { armDebugCrashTrigger, installCrashReporter } from "../lib/crashReporter";
import { addBreadcrumb } from "../lib/crash";
import { notifyError, notifyWarning } from "../lib/notify";
import "../index.css";

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
	const workspaceName = useWorkspaceStore((s) => s.workspaceTree?.name);
	const gitBranch = useGitStore((s) => s.status?.branch ?? "");

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
	// Skipped while typing in a field or when a dialog/palette owns the screen,
	// so the shortcut never fights text entry or modal flows.
	useEffect(() => {
		const onKeyDown = (e: KeyboardEvent) => {
			if (!(e.metaKey || e.ctrlKey) || e.key.toLowerCase() !== "w") return;
			// SAFETY: keydown targets are EventTargets; in a DOM window they are
			// always Nodes, so closest() is valid for the focus-context check.
			const target = e.target as HTMLElement | null;
			if (
				target?.closest(
					"input, textarea, select, [contenteditable='true'], [data-slot='dialog-content'], [data-slot='command']",
				)
			) {
				return;
			}
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
	const responseMode = useShellStore((s) => s.responseMode);

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
			<WelcomeModal />
			<CommandPalette onSelectEnvironment={(id) => void onSelectEnvironment(id)} />
			<AppShell
				rail={
					<ErrorBoundary label="Activity rail">
						<ActivityRail />
					</ErrorBoundary>
				}
				brand={
					<>
						<WorkspaceMenu>
							<div className="flex items-center gap-2 rounded-full border border-border bg-card py-1 pr-1 pl-2 hover:bg-accent">
								<img
									src={theme === "dark" ? logoDark : logoLight}
									alt="Reqly"
									className="size-5"
								/>
								<h1 className="max-w-40 truncate text-sm font-semibold tracking-tight">
									{workspaceName ?? "Reqly"}
								</h1>
								<Select
									items={[
										{ value: "", label: "No environment" },
										...environments.map((env) => ({
											value: env.id,
											label: env.name,
										})),
									]}
									value={activeEnvironment?.id ?? ""}
									onValueChange={(next) => {
										if (next !== null) void onSelectEnvironment(next);
									}}
								>
									<SelectTrigger
										aria-label={
											environmentsError ?? "Select the active environment"
										}
										className="h-6 rounded-full border-none bg-muted px-2 text-xs"
									>
										<SelectValue />
									</SelectTrigger>
									<SelectContent className="max-h-72 min-w-(--anchor-width)">
										{[
											{ value: "", label: "No environment" },
											...environments.map((env) => ({
												value: env.id,
												label: env.name,
											})),
										].map((option) => (
											<SelectItem
												key={option.value}
												value={option.value}
												className="text-xs"
											>
												{option.label}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>
						</WorkspaceMenu>
						{gitBranch && (
							<span className="font-data inline-flex items-center gap-1.5 rounded-full border border-border px-2.5 py-0.5 text-xs text-muted-foreground">
								<GitBranch className="size-3" aria-hidden />
								{gitBranch}
							</span>
						)}
					</>
				}
				headerCenter={<PaletteTriggerButton />}
				sidebar={
					<ErrorBoundary label="Sidebar">
						<WorkspaceSidebar />
					</ErrorBoundary>
				}
				commitStrip={<CommitStrip />}
				statusbar={<StatusBar />}
			>
			<section className="flex h-full min-h-0 flex-col">
				{/* Tab bar persists across every view so open request tabs stay
				 * reachable; selecting one returns to the REST client. */}
				<RequestTabs />
				{activeView === "requests" ?
					activeTab?.kind === "test" ? (
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
										<div className="flex min-h-0 min-w-0 flex-1 flex-col">
											<div className="flex shrink-0 items-center justify-end border-b border-border px-2 py-1">
												<ResponseModeToggle />
											</div>
											<ResizablePanelGroup
												orientation={responseMode === "inline" ? "vertical" : "horizontal"}
												defaultLayout={splitLayout.defaultLayout}
												onLayoutChanged={splitLayout.onLayoutChanged}
											>
												<ResizablePanel
													id="editor"
													defaultSize="50%"
													minSize="25%"
												>
													<div className={cn("h-full min-h-0 min-w-0", responseMode === "split" && "border-r border-border")}>
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
													<div className="h-full min-h-0 min-w-0 border-t border-border">
														<ErrorBoundary label="Response viewer">
															<ResponseViewer />
														</ErrorBoundary>
													</div>
												</ResizablePanel>
											</ResizablePanelGroup>
										</div>
				) : (
					<SecondaryView />
				)}
			</section>
			</AppShell>
		</ErrorBoundary>
	);
}


/** SecondaryView hosts every non-requests workspace view, keeping App's own
 * body small (react-doctor no-giant-component). */
function SecondaryView() {
	const view = useWorkspaceStore((s) => s.activeView);
	return (
		<>
							{view === "tests" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Tests">
										<TestsView />
									</ErrorBoundary>
								</section>
							) : view === "realtime" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Realtime">
										<RealtimeView />
									</ErrorBoundary>
								</section>
							) : view === "oauth" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="OAuth tokens">
										<div className="mx-auto max-w-xl p-4">
											<AuthPanel />
										</div>
									</ErrorBoundary>
								</section>
							) : view === "git" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Git">
										<div className="p-4">
											<GitPanel />
										</div>
									</ErrorBoundary>
								</section>
							) : view === "overview" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Overview">
										<OverviewHome />
									</ErrorBoundary>
								</section>
							) : view === "environments" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Environments">
										<EnvironmentsView />
									</ErrorBoundary>
								</section>
							) : view === "history" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="History">
										<HistoryView />
									</ErrorBoundary>
								</section>
							) : view === "mocks" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Mock server">
										<MocksView />
									</ErrorBoundary>
								</section>
							) : view === "diff" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="API diff">
										<DiffView />
									</ErrorBoundary>
								</section>
							) : view === "jwt" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="JWT inspector">
										<JwtInspector />
									</ErrorBoundary>
								</section>
							) : view === "graphql" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="GraphQL browser">
										<GraphqlBrowser />
									</ErrorBoundary>
								</section>
							) : view === "runners" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Runners">
										<RunnersPanel />
									</ErrorBoundary>
								</section>
							) : view === "explorer" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="OpenAPI explorer">
										<OpenapiExplorer />
									</ErrorBoundary>
								</section>
							) : view === "grpc" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="gRPC client">
										<GrpcTab tabId="grpc" />
									</ErrorBoundary>
								</section>
							) : view === "docs" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Docs generator">
										<DocsView />
									</ErrorBoundary>
								</section>
							) : view === "importexport" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Import / Export">
										<ImportExportView />
									</ErrorBoundary>
								</section>
							) : view === "codegen" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Code generation">
										<CodegenView />
									</ErrorBoundary>
								</section>
							) : view === "settings" ? (
								<section className="h-full min-h-0 overflow-y-auto">
									<ErrorBoundary label="Settings">
										<SettingsView />
									</ErrorBoundary>
								</section>
							) : null}
		</>
	);
}
