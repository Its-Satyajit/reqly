import { useEffect } from 'react'
import { Search } from 'lucide-react'

import { buildCommands, groupCommands, type Command } from '#lib/commands'
import { THEMES, type ThemePreference } from '#lib/themes'
import { WORKSPACE_VIEWS } from '#lib/views'
import {
	usePaletteStore,
	useThemeStore,
	useWorkspaceStore,
} from '#stores'
import {
	CommandDialog,
	CommandEmpty,
	CommandGroup,
	CommandInput,
	CommandItem,
	CommandList,
} from '../ui/command'
import { Button } from '../ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip'

/** Global ⌘K / Ctrl+K command palette (M44 T3), composed on cmdk. */
export function CommandPalette({
	onSelectEnvironment,
}: {
	onSelectEnvironment?: (envId: string) => void
}) {
	const open = usePaletteStore((s) => s.open)
	const close = usePaletteStore((s) => s.close)
	const toggle = usePaletteStore((s) => s.toggle)

	useEffect(() => {
		const onKeyDown = (e: KeyboardEvent) => {
			if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
				e.preventDefault()
				toggle()
			}
		}
		window.addEventListener('keydown', onKeyDown)
		return () => window.removeEventListener('keydown', onKeyDown)
	}, [toggle])

	return (
		<CommandDialog open={open} onOpenChange={(next) => (next ? undefined : close())}>
			<PaletteBody onClose={close} onSelectEnvironment={onSelectEnvironment} />
		</CommandDialog>
	)
}

function PaletteBody({
	onClose,
	onSelectEnvironment,
}: {
	onClose: () => void
	onSelectEnvironment?: (envId: string) => void
}) {
	const commands = useAppCommands(onSelectEnvironment)
	const recent = usePaletteStore((s) => s.recent)

	// With an empty query cmdk lists everything; put recently used first.
	const recentFirst = (a: Command, b: Command) =>
		Number(recent.includes(b.id)) - Number(recent.includes(a.id))
	const grouped = [...groupCommands(commands).entries()].map(([group, members]) => [
		group,
		[...members].sort(recentFirst),
	] as const)

	const run = (cmd: Command) => {
		usePaletteStore.getState().recordRun(cmd.id)
		onClose()
		cmd.run()
	}

	return (
		<>
			<CommandInput placeholder="Search or jump…" />
			<CommandList>
				<CommandEmpty>No matching commands</CommandEmpty>
				{grouped.map(([group, members]) => (
					<CommandGroup key={group} heading={group}>
						{members.map((cmd) => (
							<CommandItem
								key={cmd.id}
								value={`${cmd.label} ${(cmd.keywords ?? []).join(' ')}`}
								onSelect={() => run(cmd)}
							>
								{cmd.label}
							</CommandItem>
						))}
					</CommandGroup>
				))}
			</CommandList>
		</>
	)
}

/** Assembles palette commands from live app state. Git actions join in T4/T7. */
export function useAppCommands(
	onSelectEnvironment?: (envId: string) => void,
): Command[] {
	const setActiveView = useWorkspaceStore((s) => s.setActiveView)
	const environments = useWorkspaceStore((s) => s.environments)
	return buildCommands({
		views: WORKSPACE_VIEWS.map((v) => ({ id: v.id, label: v.label })),
		navigate: (viewId) =>
			// SAFETY: viewIds come from WORKSPACE_VIEWS, whose ids mirror the
			// WorkspaceView union one-to-one.
			setActiveView(viewId as Parameters<typeof setActiveView>[0]),
		themes: THEMES.map((t) => ({ id: t.id, label: t.label })),
		// SAFETY: ids originate from the THEMES registry itself.
		setTheme: (id) => useThemeStore.getState().setTheme(id as ThemePreference),
		environments: environments.map((e) => ({ id: e.id, name: e.name })),
		selectEnvironment: (id) => onSelectEnvironment?.(id),
	})
}

/** Header trigger button for the palette (⌘K hint). */
export function PaletteTriggerButton() {
	const openPalette = usePaletteStore((s) => s.openPalette)
	return (
		<Tooltip>
			<TooltipTrigger render={<Button variant="ghost" size="sm" onClick={openPalette} className="gap-2 text-muted-foreground" />}>
				<Search data-icon="inline-start" />
				<span>Search or jump…</span>
				<kbd className="font-data rounded border border-border px-1 text-xs">⌘K</kbd>
			</TooltipTrigger>
			<TooltipContent>Command palette (⌘K)</TooltipContent>
		</Tooltip>
	)
}
