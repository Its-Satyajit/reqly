import { Columns2, Rows2 } from 'lucide-react'

import { cn } from '#lib/utils'
import { useShellStore, type ResponseMode } from '#stores/useShellStore'
import { Button } from '../ui/button'
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from '../ui/tooltip'

const MODES: { mode: ResponseMode; label: string; icon: typeof Columns2 }[] = [
	{ mode: 'split', label: 'Response beside editor', icon: Columns2 },
	{ mode: 'inline', label: 'Response below editor', icon: Rows2 },
]

/** Split/inline response layout preference (M44 T5). */
export function ResponseModeToggle({ className }: { className?: string }) {
	const responseMode = useShellStore((s) => s.responseMode)
	const setResponseMode = useShellStore((s) => s.setResponseMode)

	return (
		<div
			className={cn('flex items-center gap-0.5', className)}
			role="radiogroup"
			aria-label="Response layout"
		>
			{MODES.map(({ mode, label, icon: Icon }) => (
				<Tooltip key={mode}>
					<TooltipTrigger
						render={
							<Button
								variant="ghost"
								size="icon-sm"
								onClick={() => setResponseMode(mode)}
								aria-checked={responseMode === mode}
								className={cn(
									responseMode === mode && 'bg-muted text-foreground',
								)}
							>
								<Icon className="size-4" aria-hidden />
							</Button>
						}
					/>
					<TooltipContent>{label}</TooltipContent>
				</Tooltip>
			))}
		</div>
	)
}
