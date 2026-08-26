import type { ReactNode } from "react"

import { cn } from "#lib/utils"
import { Button } from "#components/ui/button"
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "#components/ui/tooltip"

export interface CycleToggleOption<T extends string> {
	value: T
	/** Shown in the tooltip when this option is the switch target. */
	label: string
	icon: ReactNode
}

interface CycleToggleProps<T extends string> {
	value: T
	options: [CycleToggleOption<T>, CycleToggleOption<T>, ...CycleToggleOption<T>[]]
	onChange: (next: T) => void
	ariaLabel: string
	className?: string
}

/**
 * One button for a two-state layout choice (e.g. split vs stacked panes).
 * Renders the icon of the state it will switch TO and cycles through
 * `options` on click, so a single control replaces a radio pair wherever
 * two column/row layouts share a surface.
 */
export function CycleToggle<T extends string>({
	value,
	options,
	onChange,
	ariaLabel,
	className,
}: CycleToggleProps<T>) {
	const currentIndex = options.findIndex((o) => o.value === value)
	const next = options[(currentIndex + 1) % options.length]

	return (
		<Tooltip>
			<TooltipTrigger
				render={
					<Button
						variant="ghost"
						size="icon-sm"
						onClick={() => onChange(next.value)}
						aria-label={`${ariaLabel}: ${next.label}`}
						title={undefined}
						className={cn("text-muted-foreground hover:text-foreground", className)}
					/>
				}
			>
				{next.icon}
			</TooltipTrigger>
			<TooltipContent>{`Switch to ${next.label}`}</TooltipContent>
		</Tooltip>
	)
}
