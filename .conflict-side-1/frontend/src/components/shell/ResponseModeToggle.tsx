import { Columns2, Rows2 } from 'lucide-react'

import { useShellStore, type ResponseMode } from '#stores/useShellStore'
import { ToggleGroup, ToggleGroupItem } from '#components/ui/toggle-group'

/** Split/inline response layout preference (M44 T5). */
export function ResponseModeToggle({ className }: { className?: string }) {
	const responseMode = useShellStore((s) => s.responseMode)
	const setResponseMode = useShellStore((s) => s.setResponseMode)

	return (
		<ToggleGroup
			variant="default"
			size="sm"
			className={className}
			aria-label="Response layout"
			value={[responseMode]}
			onValueChange={(values) => {
				const next = values[values.length - 1]
				if (next != null) {
					// SAFETY: single-select group yields the clicked item's string id
					setResponseMode(next as ResponseMode)
				}
			}}
		>
			<ToggleGroupItem value="split" aria-label="response beside editor">
				<Columns2 className="size-4" aria-hidden />
			</ToggleGroupItem>
			<ToggleGroupItem value="inline" aria-label="response below editor">
				<Rows2 className="size-4" aria-hidden />
			</ToggleGroupItem>
		</ToggleGroup>
	)
}
