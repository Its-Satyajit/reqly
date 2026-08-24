import { Columns2, MonitorCog, Rows2, SunMoon } from 'lucide-react'

import { cn } from '#lib/utils'
import { THEMES } from '#lib/themes'
import {
	useShellStore,
	useThemeStore,
	type ResponseMode,
} from '#stores'

function SettingsGroup({
	label,
	children,
}: {
	label: string
	children: React.ReactNode
}) {
	return (
		<section aria-label={label} className="flex flex-col gap-1.5">
			<p className="text-xs font-medium text-muted-foreground">{label}</p>
			{children}
		</section>
	)
}

function AppearanceSection() {
	const preference = useThemeStore((s) => s.preference)
	const setTheme = useThemeStore((s) => s.setTheme)

	const options = [
		{ id: 'system', label: 'System', hint: 'Follow the operating system setting' },
		...THEMES.map((t) => ({
			id: t.id,
			label: t.label,
			hint: `Built-in ${t.appearance} theme`,
		})),
	]

	return (
		<SettingsGroup label="Theme">
			<div role="radiogroup" className="flex flex-col gap-1.5">
				{options.map((opt) => (
					<button
						key={opt.id}
						type="button"
						role="radio"
						aria-checked={preference === opt.id}
						// SAFETY: opt.id comes from the THEMES registry plus the literal
						// 'system' — both members of ThemePreference.
						onClick={() => setTheme(opt.id as typeof preference)}
						className={cn(
							'flex items-center justify-between rounded-lg border px-3 py-2 text-left text-xs transition-colors',
							preference === opt.id
								? 'border-ring bg-muted/60'
								: 'border-border hover:bg-muted/40',
						)}
					>
						<span>
							<span className="block font-medium text-foreground">{opt.label}</span>
							<span className="text-muted-foreground">{opt.hint}</span>
						</span>
						{opt.id === 'system' && (
							<SunMoon className="size-4 shrink-0 text-muted-foreground" aria-hidden />
						)}
					</button>
				))}
			</div>
		</SettingsGroup>
	)
}

function ResponseLayoutSection() {
	const responseMode = useShellStore((s) => s.responseMode)
	const setResponseMode = useShellStore((s) => s.setResponseMode)

	const modes: { mode: ResponseMode; label: string; icon: typeof Columns2 }[] = [
		{ mode: 'split', label: 'Beside the editor', icon: Columns2 },
		{ mode: 'inline', label: 'Below the editor', icon: Rows2 },
	]

	return (
		<SettingsGroup label="Response layout">
			<div role="radiogroup" className="flex flex-col gap-1.5">
				{modes.map(({ mode, label, icon: Icon }) => (
					<button
						key={mode}
						type="button"
						role="radio"
						aria-checked={responseMode === mode}
						onClick={() => setResponseMode(mode)}
						className={cn(
							'flex items-center gap-2 rounded-lg border px-3 py-2 text-left text-xs transition-colors',
							responseMode === mode
								? 'border-ring bg-muted/60'
								: 'border-border hover:bg-muted/40',
						)}
					>
						<Icon className="size-4 shrink-0 text-muted-foreground" aria-hidden />
						<span className="font-medium text-foreground">{label}</span>
					</button>
				))}
			</div>
		</SettingsGroup>
	)
}

/** Settings view (M44 T6): appearance + layout preferences. */
export function SettingsView() {
	return (
		<div className="mx-auto w-full max-w-md p-4">
			<h1 className="flex items-center gap-2 pb-3 text-sm font-semibold tracking-tight">
				<MonitorCog className="size-4 text-muted-foreground" aria-hidden />
				Settings
			</h1>
			<div className="flex flex-col gap-5">
				<AppearanceSection />
				<ResponseLayoutSection />
			</div>
		</div>
	)
}
