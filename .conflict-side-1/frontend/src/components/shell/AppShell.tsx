import { useState, type ReactNode } from 'react'
import { PanelLeftClose, PanelLeftOpen } from 'lucide-react'
import { useDefaultLayout, usePanelRef } from 'react-resizable-panels'

import {
	ResizableHandle,
	ResizablePanel,
	ResizablePanelGroup,
} from '../ui/resizable'
import { Button } from '../ui/button'
import { ThemeToggle } from '../ThemeToggle'
import { useShellStore } from '#stores'
import { shellStorage } from './storage'

export interface AppShellProps {
	/** Brand block rendered at the far left of the header. */
	brand?: ReactNode;
	/** Icon activity rail rendered on the left edge of the body (G-17.3.1). */
	rail?: ReactNode;
	/** Center header slot — command-palette trigger lands here (M44 T3). */
	headerCenter?: ReactNode;
	/** Right-aligned header actions (environment pill, settings, …). */
	headerActions?: ReactNode;
	sidebar: ReactNode;
	/** Main column content: tabstrip + view host. */
	children: ReactNode;
	/** Right-hand inspector mount point content; hidden while closed. */
	inspector?: ReactNode;
	/** Commit-strip slot above the statusbar; populated by M44 T7. */
	commitStrip?: ReactNode;
	statusbar?: ReactNode;
}

/**
 * Atlas-style application shell (M44 T2): header / [sidebar | main |
 * inspector] body / statusbar, mirroring the ui-demos atlas layout.
 * Sidebar gutter sizes persist via react-resizable-panels' default-layout
 * storage; the inspector is collapsible and driven by useShellStore.
 */
export function AppShell({
	brand,
	rail,
	headerCenter,
	headerActions,
	sidebar,
	children,
	inspector,
	commitStrip,
	statusbar,
}: AppShellProps) {
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
	const inspectorOpen = useShellStore((s) => s.inspectorOpen);

	return (
		<div className="flex h-screen flex-col overflow-hidden">
			<header className="flex h-11 shrink-0 items-center justify-between border-b border-border px-3">
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
					{brand}
				</div>
				{headerCenter}
				<div className="flex items-center gap-2">
					{headerActions}
					<ThemeToggle />
				</div>
			</header>
			<div className="flex min-h-0 flex-1">
				{rail}
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
						{sidebar}
					</ResizablePanel>
					<ResizableHandle data-split="side" />
					<ResizablePanel id="main" minSize="35%">
						{children}
					</ResizablePanel>
					{inspector !== undefined && (
						<>
							{inspectorOpen && <ResizableHandle data-split="inspector" />}
							<ResizablePanel
								id="inspector"
								collapsible
								collapsedSize={0}
								defaultSize="25%"
								minSize={220}
								maxSize="45%"
							>
								{inspector}
							</ResizablePanel>
						</>
					)}
				</ResizablePanelGroup>
			</div>
			{commitStrip}
			{statusbar !== undefined && (
				<footer className="flex h-6 shrink-0 items-center justify-between border-t border-border px-3 text-xs text-muted-foreground">
					{statusbar}
				</footer>
			)}
		</div>
	);
}
