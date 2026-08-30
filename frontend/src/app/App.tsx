import { useEffect, useState } from "react";
import { ErrorBoundary } from "../components/ErrorBoundary";
import { CrashOverlay } from "../components/CrashOverlay";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "#components/ui/alert-dialog";
import {
	ContextSidebar,
	RequestTabs,
	RunView,
	StatusBar,
	TopBar,
	ToolRail,
} from "../components";
import { AppShell } from "../components/shell/AppShell";
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
import { HomeView } from "../features/workspace-home/HomeView";
import { SettingsView } from "../features/settings-view/SettingsView";
import { WebSocketPage, SSEPage } from "../features/realtime-pages/RealtimePage";
import { SpecEditorView } from "../features/spec-editor/SpecEditorView";
import { CommandPalette } from "../features/command-palette/CommandPalette";
import { BottomPanel } from "../components/shell/BottomPanel";
import { registerDefaultPaletteProviders } from "../lib/paletteProviders";
import { RequestEditor } from "../features/request-editor/RequestEditor";
import { ResponseViewer } from "../features/response-viewer/ResponseViewer";
import { useWorkspaceStore } from "../stores";
import { useWorkspaceBootstrapStore } from "../stores/useWorkspaceBootstrap";
import { WorkspaceEmptyState } from "../features/workspace-bootstrap/WorkspaceEmptyState";
import { NEW_REQUEST_TAB_ID } from "../stores/useRequestStore";
import { useDefaultLayout } from "react-resizable-panels";
import { armDebugCrashTrigger, installCrashReporter } from "../lib/crashReporter";
import { shellStorage } from "../components/shell/storage";
import {
	ResizableHandle,
	ResizablePanel,
	ResizablePanelGroup,
} from "../components/ui/resizable";

export function App() {
	const bootChecked = useWorkspaceBootstrapStore((s) => s.checked);
	const bootFound = useWorkspaceBootstrapStore((s) => s.status?.found ?? false);
	const initBootstrap = useWorkspaceBootstrapStore((s) => s.init);
	const refreshEnvironments = useWorkspaceStore((s) => s.refreshEnvironments);
	const activeView = useWorkspaceStore((s) => s.activeView);
	const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);
	const openTabs = useWorkspaceStore((s) => s.openTabs);
	const activeTabId = useWorkspaceStore((s) => s.activeTabId);
	const activeTab = openTabs.find((t) => t.id === activeTabId);

	useEffect(() => {
		installCrashReporter();
		return armDebugCrashTrigger();
	}, []);

	useEffect(() => {
		void initBootstrap();
	}, [initBootstrap]);
	useEffect(() => { registerDefaultPaletteProviders(); }, []);

	useEffect(() => {
		void refreshEnvironments();
		void refreshWorkspace();
		const { openTabs, openTab } = useWorkspaceStore.getState();
		if (openTabs.length === 0) {
			openTab({ id: NEW_REQUEST_TAB_ID, title: "New Request" });
		}
	}, [refreshEnvironments, refreshWorkspace]);

	const [toolRailCollapsed, setToolRailCollapsed] = useState(false);
	const toggleToolRail = () => setToolRailCollapsed((prev) => !prev);

	const [splitOrientation, setSplitOrientation] = useState<"horizontal" | "vertical">("horizontal");
	const splitLayout = useDefaultLayout({
		id: "reqly-shell-split",
		storage: shellStorage,
	});

	const pendingView = useWorkspaceStore((s) => s.pendingView);
	const confirmPendingView = useWorkspaceStore((s) => s.confirmPendingView);
	const cancelPendingView = useWorkspaceStore((s) => s.cancelPendingView);

	if (!bootChecked) {
		return (
			<ErrorBoundary variant="root">
				<div className="flex min-h-screen items-center justify-center bg-background" />
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

	return (
		<ErrorBoundary variant="root">
			<Toaster />
			<CrashOverlay />
			<AppShell
				topBar={<TopBar />}
				toolRail={<ToolRail collapsed={toolRailCollapsed} onToggleCollapse={toggleToolRail} />}
				sidebar={
					<ErrorBoundary label="Context sidebar">
						<ContextSidebar />
					</ErrorBoundary>
				}
				bottom={<BottomPanel />}
				statusBar={
					<ErrorBoundary label="Status bar">
						<StatusBar />
					</ErrorBoundary>
				}
			>
				<div className="h-full min-h-0 overflow-hidden">
					{activeView === "home" ? (
						<ErrorBoundary label="Workspace home">
							<HomeView />
						</ErrorBoundary>
					) : activeView === "environments" ? (
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
					) : activeView === "websocket" ? (
						<section className="h-full min-h-0">
							<ErrorBoundary label="WebSocket"><WebSocketPage /></ErrorBoundary>
						</section>
					) : activeView === "sse" ? (
						<section className="h-full min-h-0">
							<ErrorBoundary label="SSE"><SSEPage /></ErrorBoundary>
						</section>
					) : activeView === "settings" ? (
						<section className="h-full min-h-0 overflow-y-auto">
							<ErrorBoundary label="Settings"><SettingsView /></ErrorBoundary>
						</section>
					) : activeView === "spec-editor" ? (
						<section className="h-full min-h-0">
							<ErrorBoundary label="Spec Editor"><SpecEditorView /></ErrorBoundary>
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
								<div className="flex min-h-0 min-w-0 flex-1 flex-col">
									<div className="flex shrink-0 justify-end border-b border-border px-1 py-0.5">
										<button
											type="button"
											onClick={() => setSplitOrientation((o) => (o === "horizontal" ? "vertical" : "horizontal"))}
											className="rounded px-1.5 py-0.5 text-[10px] text-muted-foreground hover:bg-muted hover:text-foreground"
											title="Toggle Request/Response split orientation"
										>
											{splitOrientation === "horizontal" ? "↔ Split" : "↕ Split"}
										</button>
									</div>
									<ResizablePanelGroup
										orientation={splitOrientation}
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
				</div>
			</AppShell>
			<CommandPalette />
			<AlertDialog
				open={pendingView != null}
				onOpenChange={(open) => {
					if (!open) cancelPendingView();
				}}
			>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Discard unsaved environment changes?</AlertDialogTitle>
						<AlertDialogDescription>
							Switching views discards edits that were never saved.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Keep editing</AlertDialogCancel>
						<AlertDialogAction onClick={confirmPendingView}>
							Discard changes
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</ErrorBoundary>
	);
}
