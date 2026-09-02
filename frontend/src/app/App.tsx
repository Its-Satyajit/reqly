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
import { ContextSidebar } from "../components/shell/ContextSidebar";
import { RequestTabs } from "../components/RequestTabs";
import { RunView } from "../components/RunView";
import { StatusBar } from "../components/shell/StatusBar";
import { TopBar } from "../components/shell/TopBar";
import { ToolRail } from "../components/shell/ToolRail";
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
import { MqttView } from "../features/mqtt-view/MqttView";
import { SocketIOView } from "../features/socketio-view/SocketIOView";
import { PolicyRbacView } from "../features/governance/PolicyRbacView";
import { AuditView } from "../features/audit-view/AuditView";
import { SsoScimView } from "../features/sso-view/SsoScimView";
import { CollabView } from "../features/collab-view/CollabView";
import { AutomationView } from "../features/automation-view/AutomationView";
import { WorkflowView } from "../features/workflow-view/WorkflowView";
import { MonitorView } from "../features/monitor-view/MonitorView";
import { ChangelogView } from "../features/changelog-view/ChangelogView";
import { SpecEditorView } from "../features/spec-editor/SpecEditorView";
import { CommandPalette } from "../features/command-palette/CommandPalette";
import { BottomPanel } from "../components/shell/BottomPanel";
import { registerDefaultPaletteProviders } from "../lib/paletteProviders";
import { RequestEditor } from "../features/request-editor/RequestEditor";
import { ResponseViewer } from "../features/response-viewer/ResponseViewer";
import { useWorkspaceStore } from "../stores/useWorkspaceStore";
import { useShellStore } from "../stores/useShellStore";
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

type WorkspaceView = ReturnType<typeof useWorkspaceStore.getState>["activeView"];
type RequestTab = ReturnType<typeof useWorkspaceStore.getState>["openTabs"][number];

const SCROLLABLE_VIEW_RENDERERS = {
	environments: () => (
		<ErrorBoundary label="Environments">
			<EnvironmentsView />
		</ErrorBoundary>
	),
	mocks: () => (
		<ErrorBoundary label="Mock server">
			<MocksView />
		</ErrorBoundary>
	),
	diff: () => (
		<ErrorBoundary label="API diff">
			<DiffView />
		</ErrorBoundary>
	),
	jwt: () => (
		<ErrorBoundary label="JWT inspector">
			<JwtInspector />
		</ErrorBoundary>
	),
	graphql: () => (
		<ErrorBoundary label="GraphQL browser">
			<GraphqlBrowser />
		</ErrorBoundary>
	),
	runners: () => (
		<ErrorBoundary label="Runners">
			<RunnersPanel />
		</ErrorBoundary>
	),
	explorer: () => (
		<ErrorBoundary label="OpenAPI explorer">
			<OpenapiExplorer />
		</ErrorBoundary>
	),
	docs: () => (
		<ErrorBoundary label="Docs generator">
			<DocsView />
		</ErrorBoundary>
	),
	policy: () => (
		<ErrorBoundary label="Policy">
			<PolicyRbacView />
		</ErrorBoundary>
	),
	sso: () => (
		<ErrorBoundary label="SSO">
			<SsoScimView />
		</ErrorBoundary>
	),
	collab: () => (
		<ErrorBoundary label="Collaboration">
			<CollabView />
		</ErrorBoundary>
	),
	automation: () => (
		<ErrorBoundary label="Automation">
			<AutomationView />
		</ErrorBoundary>
	),
	workflow: () => (
		<ErrorBoundary label="Workflow">
			<WorkflowView />
		</ErrorBoundary>
	),
	monitor: () => (
		<ErrorBoundary label="Monitor">
			<MonitorView />
		</ErrorBoundary>
	),
	changelog: () => (
		<ErrorBoundary label="Changelog">
			<ChangelogView />
		</ErrorBoundary>
	),
	settings: () => (
		<ErrorBoundary label="Settings">
			<SettingsView />
		</ErrorBoundary>
	),
} satisfies Record<string, () => React.ReactNode>;

const FULL_VIEW_RENDERERS = {
	history: () => (
		<ErrorBoundary label="History">
			<HistoryView />
		</ErrorBoundary>
	),
	grpc: () => (
		<ErrorBoundary label="gRPC client">
			<GrpcTab tabId="grpc" />
		</ErrorBoundary>
	),
	websocket: () => (
		<ErrorBoundary label="WebSocket">
			<WebSocketPage />
		</ErrorBoundary>
	),
	sse: () => (
		<ErrorBoundary label="SSE">
			<SSEPage />
		</ErrorBoundary>
	),
	mqtt: () => (
		<ErrorBoundary label="MQTT">
			<MqttView />
		</ErrorBoundary>
	),
	socketio: () => (
		<ErrorBoundary label="Socket.IO">
			<SocketIOView />
		</ErrorBoundary>
	),
	audit: () => (
		<ErrorBoundary label="Audit">
			<AuditView />
		</ErrorBoundary>
	),
	"spec-editor": () => (
		<ErrorBoundary label="Spec Editor">
			<SpecEditorView />
		</ErrorBoundary>
	),
} satisfies Record<string, () => React.ReactNode>;

function WorkspaceMain({
	activeView,
	activeTab,
	splitOrientation,
	splitLayout,
}: {
	activeView: WorkspaceView;
	activeTab: RequestTab | undefined;
	splitOrientation: "horizontal" | "vertical";
	splitLayout: ReturnType<typeof useDefaultLayout>;
}) {
	if (activeView === "home") {
		return (
			<ErrorBoundary label="Workspace home">
				<HomeView />
			</ErrorBoundary>
		);
	}

	if (activeView in SCROLLABLE_VIEW_RENDERERS) {
		// SAFETY: in operator narrows activeView to known scrollable keys, validated by satisfies Record
		const scrollable = SCROLLABLE_VIEW_RENDERERS[activeView as keyof typeof SCROLLABLE_VIEW_RENDERERS];
		return <section className="h-full min-h-0 overflow-y-auto">{scrollable()}</section>;
	}

	if (activeView in FULL_VIEW_RENDERERS) {
		// SAFETY: in operator narrows activeView to known full keys
		const full = FULL_VIEW_RENDERERS[activeView as keyof typeof FULL_VIEW_RENDERERS];
		return <section className="h-full min-h-0">{full()}</section>;
	}

	return (
		<section className="flex h-full min-h-0 flex-col">
			<ErrorBoundary label="Request tabs">
				<RequestTabs />
			</ErrorBoundary>
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
					<ResizablePanelGroup
						orientation={splitOrientation}
						defaultLayout={splitLayout.defaultLayout}
						onLayoutChanged={splitLayout.onLayoutChanged}
					>
						<ResizablePanel id="editor" defaultSize="50%" minSize="25%">
							<div className="h-full min-h-0 min-w-0 border-r border-border">
								<ErrorBoundary label="Request editor">
									<RequestEditor />
								</ErrorBoundary>
							</div>
						</ResizablePanel>
						<ResizableHandle withHandle />
						<ResizablePanel id="viewer" defaultSize="50%" minSize="25%">
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
	);
}

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

	const splitOrientation = useShellStore((s) => s.responseMode);
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

	const isWorkspaceHome = activeView === "home";

	return (
		<ErrorBoundary variant="root">
			<Toaster />
			<CrashOverlay />
			<AppShell
				topBar={
					<ErrorBoundary label="Top bar">
						<TopBar />
					</ErrorBoundary>
				}
				toolRail={
					<ErrorBoundary label="Tool rail">
						<ToolRail collapsed={toolRailCollapsed} onToggleCollapse={toggleToolRail} />
					</ErrorBoundary>
				}
				sidebar={
					isWorkspaceHome ? null : (
						<ErrorBoundary label="Context sidebar">
							<ContextSidebar />
						</ErrorBoundary>
					)
				}
				bottom={
					<ErrorBoundary label="Bottom panel">
						<BottomPanel />
					</ErrorBoundary>
				}
				statusBar={
					<ErrorBoundary label="Status bar">
						<StatusBar />
					</ErrorBoundary>
				}
			>
				<div className="h-full min-h-0 overflow-hidden">
					<WorkspaceMain activeView={activeView} activeTab={activeTab} splitOrientation={splitOrientation} splitLayout={splitLayout} />
				</div>
			</AppShell>
			<ErrorBoundary label="Command palette">
				<CommandPalette />
			</ErrorBoundary>
			<AlertDialog
				open={pendingView != null}
				onOpenChange={(open) => {
					if (!open) cancelPendingView();
				}}
			>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Discard unsaved environment changes?</AlertDialogTitle>
						<AlertDialogDescription>Switching views discards edits that were never saved.</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Keep editing</AlertDialogCancel>
						<AlertDialogAction onClick={confirmPendingView}>Discard changes</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</ErrorBoundary>
	);
}
