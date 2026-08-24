import { workspaceViewLabel } from '#lib/views'
import { usePaletteStore, useShellStore, useWorkspaceStore } from '#stores'

/** Shell statusbar: current view on the left, active environment on the right. */
export function StatusBar() {
	const activeView = useWorkspaceStore((s) => s.activeView);
	const environments = useWorkspaceStore((s) => s.environments);
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);
	const inspectorOpen = useShellStore((s) => s.inspectorOpen);
	const paletteOpen = usePaletteStore((s) => s.open);

	const env = environments.find((e) => e.id === activeEnvironmentId);

	return (
		<>
			<span>
				{workspaceViewLabel(activeView)}
				{inspectorOpen ? ' · Inspector' : ''}
				{paletteOpen ? ' · Palette' : ''}
			</span>
			<span className="font-data">{env ? `env: ${env.name}` : 'no environment'}</span>
		</>
	)
}
