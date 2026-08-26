import { Columns2, Rows2 } from 'lucide-react'

import { useShellStore, type ResponseMode } from '#stores/useShellStore'
import { CycleToggle } from '../ui/cycle-toggle'

/** Split/inline response layout preference (M44 T5) — one toggle that shows
 * the layout it will switch to. */
export function ResponseModeToggle({ className }: { className?: string }) {
	const responseMode = useShellStore((s) => s.responseMode)
	const setResponseMode = useShellStore((s) => s.setResponseMode)

	return (
		<CycleToggle<ResponseMode>
			value={responseMode}
			onChange={setResponseMode}
			ariaLabel="Response layout"
			className={className}
			options={[
				{
					value: 'split',
					label: 'response beside editor',
					icon: <Columns2 className="size-4" aria-hidden />,
				},
				{
					value: 'inline',
					label: 'response below editor',
					icon: <Rows2 className="size-4" aria-hidden />,
				},
			]}
		/>
	)
}
