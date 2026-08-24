import { useShellStore, useWorkspaceStore } from '#stores'

const VIEW_LABELS = {
	requests: 'Requests',
	environments: 'Environments',
	history: 'History',
	mocks: 'Mock servers',
	diff: 'API diff',
	jwt: 'JWT inspector',
	graphql: 'GraphQL browser',
	runners: 'Runners',
	explorer: 'OpenAPI explorer',
	grpc: 'gRPC client',
	docs: 'Docs generator',
} as const

/** Shell statusbar: current view on the left, active environment on the right. */
export function StatusBar() {
	const activeView = useWorkspaceStore((s) => s.activeView)
	const environments = useWorkspaceStore((s) => s.environments)
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId)
	const inspectorOpen = useShellStore((s) => s.inspectorOpen)

	const env = environments.find((e) => e.id === activeEnvironmentId)

	return (
		<>
			<span>
				{VIEW_LABELS[activeView] ?? activeView}
				{inspectorOpen ? ' · Inspector' : ''}
			</span>
			<span className="font-data">{env ? `env: ${env.name}` : 'no environment'}</span>
		</>
	)
}
