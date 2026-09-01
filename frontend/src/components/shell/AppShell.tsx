import { useState, useEffect, type ReactNode } from "react";
import { useDefaultLayout, usePanelRef } from "react-resizable-panels";
import {
	ResizableHandle,
	ResizablePanel,
	ResizablePanelGroup,
} from "../ui/resizable";
import { useBottomPanelStore } from "#stores";
import { useKeyboardMap } from "../../hooks/useKeyboardMap";
import { shellStorage } from "./storage";

export interface AppShellProps {
	topBar?: ReactNode;
	toolRail?: ReactNode;
	sidebar: ReactNode;
	children: ReactNode;
	bottom?: ReactNode;
	statusBar?: ReactNode;
}

/**
 * Shell layout for Slice 02 — 5-zone model (TopBar / ToolRail / Sidebar / Main + Bottom / StatusBar).
 * Sidebar + BottomPanel sizes persist via react-resizable-panels storage.
 * ToolRail is rendered alongside the sidebar's ResizablePanelGroup as a fixed gutter.
 */
export function AppShell({
	topBar,
	toolRail,
	sidebar,
	children,
	bottom,
	statusBar,
}: AppShellProps) {
	const sidebarLayout = useDefaultLayout({
		id: "reqly-shell-sidebar",
		storage: shellStorage,
	});
	const sidebarPanel = usePanelRef();
	const bottomLayout = useDefaultLayout({
		id: "reqly-shell-bottom",
		storage: shellStorage,
	});
	const bottomPanelRef = usePanelRef();
	const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
	const toggleSidebar = () => {
		if (sidebarCollapsed) sidebarPanel.current?.expand();
		else sidebarPanel.current?.collapse();
	};
	useKeyboardMap(toggleSidebar);

	const bottomCollapsed = useBottomPanelStore((s) => s.collapsed);
	useEffect(() => {
		if (bottomCollapsed) bottomPanelRef.current?.collapse();
		else bottomPanelRef.current?.expand();
	}, [bottomCollapsed, bottomPanelRef]);

	return (
		<div className="flex h-screen flex-col overflow-hidden">
			{topBar}
			<div className="flex min-h-0 flex-1">
				{toolRail}
				<div className="min-w-0 flex-1">
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
							minSize={200}
							maxSize="42%"
							onResize={(size) => setSidebarCollapsed(size.inPixels <= 1)}
						>
							{sidebar}
						</ResizablePanel>
						<ResizableHandle />
						<ResizablePanel id="main" minSize="35%">
							<ResizablePanelGroup
								orientation="vertical"
								defaultLayout={bottomLayout.defaultLayout}
								onLayoutChanged={bottomLayout.onLayoutChanged}
							>
								<ResizablePanel id="main-content" minSize="40%">
									<div className="h-full min-h-0 overflow-hidden">{children}</div>
								</ResizablePanel>
								<ResizableHandle />
								<ResizablePanel
									id="bottom"
									panelRef={bottomPanelRef}
									collapsible
									collapsedSize={0}
									minSize={6}
									defaultSize="28%"
									onResize={(s) => {
										if (s.inPixels <= 1)
											queueMicrotask(() => {
												if (!useBottomPanelStore.getState().collapsed)
													useBottomPanelStore.getState().setCollapsed(true);
											});
									}}
								>
									{bottom}
								</ResizablePanel>
							</ResizablePanelGroup>
						</ResizablePanel>
					</ResizablePanelGroup>
				</div>
			</div>
			{statusBar}
		</div>
	);
}
