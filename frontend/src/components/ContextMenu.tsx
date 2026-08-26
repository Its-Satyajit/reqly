import {
	ContextMenu as ContextMenuRoot,
	ContextMenuContent,
	ContextMenuTrigger,
	ContextMenuItem,
} from "#components/ui/context-menu";

export interface ContextMenuItem {
	label: string;
	onSelect: () => void;
}

interface ContextMenuProps {
	x: number;
	y: number;
	items: ContextMenuItem[];
	onClose: () => void;
}

/** Fixed-position context menu rendered at {x, y}: closes on click-away,
 * Escape, or item selection. Rendered by callers that track {x, y} in local
 * state. */
export function ContextMenu({ x, y, items, onClose }: ContextMenuProps) {
	const ref = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const close = (e: MouseEvent) => {
			// SAFETY: e.target is the DOM node the press landed on; contains()
			// only needs a Node, and every window mouse event target is one.
			if (ref.current && !ref.current.contains(e.target as Node)) onClose();
		};
		const onKey = (e: KeyboardEvent) => {
			if (e.key === "Escape") {
				e.stopPropagation();
				onClose();
			}
		};
		window.addEventListener("mousedown", close);
		window.addEventListener("keydown", onKey);
		return () => {
			window.removeEventListener("mousedown", close);
			window.removeEventListener("keydown", onKey);
		};
	}, [onClose]);

	return (
		<div
			ref={ref}
			role="menu"
			style={{ left: x, top: y }}
			className="fixed z-(--z-overlay) min-w-36 rounded-md border border-border bg-popover p-1 text-xs shadow-lg ring-1 ring-foreground/10"
		>
			<ContextMenuTrigger
				style={{ position: "fixed", left: x, top: y, width: 0, height: 0 }}
				aria-hidden
				tabIndex={-1}
			/>
			<ContextMenuContent className="min-w-36 text-xs">
				{items.map((item) => (
					<ContextMenuItem
						key={item.label}
						onSelect={() => {
							item.onSelect();
							onClose();
						}}
					>
						{item.label}
					</ContextMenuItem>
				))}
			</ContextMenuContent>
		</ContextMenuRoot>
	);
}
