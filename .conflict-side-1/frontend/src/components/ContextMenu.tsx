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
	return (
		<ContextMenuRoot
			open
			onOpenChange={(open) => {
				if (!open) onClose();
			}}
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
