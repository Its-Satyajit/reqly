import type { CSSProperties, ReactNode } from "react";
import { cn } from "#lib/utils";

export interface ViewShellProps {
	/** Accessible name for the view region. */
	label?: string;
	/** Removes the default page padding for full-bleed views. */
	flush?: boolean;
	className?: string;
	children: ReactNode;
}

/** ViewShell is the page scaffold every secondary view mounts into: a full
 * height, width-constrained column with the app's standard page padding.
 * Views compose their internals inside it instead of re-rolling the
 * flex/height/padding boilerplate (which is where narrow-pane breakage
 * crept in). */
export function ViewShell({ label, flush, className, children }: ViewShellProps) {
	return (
		<div
			aria-label={label}
			className={cn(
				"flex h-full min-h-0 min-w-0 flex-col",
				!flush && "gap-4 p-4",
				className,
			)}
		>
			{children}
		</div>
	);
}

export interface SplitViewProps {
	/** The lister/navigator pane. Fluid: shrinks with the container. */
	aside: ReactNode;
	/** The primary pane; always keeps at least ~55% of the width. */
	children: ReactNode;
	/** Which edge the aside docks to. Default "left". */
	asideSide?: "left" | "right";
	/** Accessible name for the aside region. */
	asideLabel?: string;
	/** Fluid aside sizing: pixel floor, preferred fraction of the container,
	 * pixel ceiling. Defaults to a 13–20rem band at 30%. */
	asideWidth?: { min: number; preferred: number; max: number };
	asideClassName?: string;
	className?: string;
}

const DEFAULT_ASIDE_WIDTH = { min: 208, preferred: 0.3, max: 320 };

/** SplitView is the reusable two-pane view layout: a fluid aside plus a
 * main pane that never drops below ~55% of the container. The aside width
 * is a clamp() of the container, not a fixed w-64/w-96 — fixed-width
 * asides were what broke views in narrow split-screen panes. */
export function SplitView({
	aside,
	children,
	asideSide = "left",
	asideLabel,
	asideWidth = DEFAULT_ASIDE_WIDTH,
	asideClassName,
	className,
}: SplitViewProps) {
	const { min, preferred, max } = asideWidth;
	const asideStyle: CSSProperties = {
		width: `clamp(${min}px, ${Math.round(preferred * 100)}%, ${max}px)`,
	};
	const asidePane = (
		<aside
			aria-label={asideLabel}
			style={asideStyle}
			className={cn(
				"flex min-w-0 shrink-0 grow-0 flex-col gap-3 overflow-y-auto",
				asideClassName,
			)}
		>
			{aside}
		</aside>
	);
	return (
		<div className={cn("flex h-full min-h-0 min-w-0 gap-4", className)}>
			{asideSide === "left" && asidePane}
			<div className="flex min-w-0 flex-1 flex-col overflow-hidden">
				{children}
			</div>
			{asideSide === "right" && asidePane}
		</div>
	);
}
