import { useEffect, useState } from "react";
import {
	Columns2,
	Eye,
	GitBranch,
	Keyboard,
	Info,
	Lock,
	MonitorCog,
	Rows2,
	ShieldCheck,
	SunMoon,
} from "lucide-react";
import { cn } from "#lib/utils";
import { THEMES } from "#lib/themes";
import { useShellStore, useThemeStore, type ResponseMode } from "#stores";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { useHistoryStore } from "#stores/useHistoryStore";
import { SplitView, ViewShell } from "../../components/shell/ViewLayout";

type SettingsSectionId = "appearance" | "editor" | "privacy" | "storage" | "shortcuts" | "about";

const SECTIONS: { id: SettingsSectionId; label: string }[] = [
	{ id: "appearance", label: "Appearance" },
	{ id: "editor", label: "Editor" },
	{ id: "privacy", label: "Privacy" },
	{ id: "storage", label: "Storage" },
	{ id: "shortcuts", label: "Shortcuts" },
	{ id: "about", label: "About" },
];

function SettingsGroup({
	label,
	children,
}: {
	label: string;
	children: React.ReactNode;
}) {
	return (
		<section aria-label={label} className="flex flex-col gap-1.5">
			<p className="text-xs font-medium text-muted-foreground">{label}</p>
			{children}
		</section>
	)
}

/** ThemeSwatch is one clickable light/dark preview. No card chrome — the
 * preview itself is the control; selection shows as a ring. */
function ThemeSwatch({
	label,
	appearance,
	selected,
	onSelect,
}: {
	label: string;
	appearance: "light" | "dark";
	selected: boolean;
	onSelect: () => void;
}) {
	return (
		<button
			type="button"
			role="radio"
			aria-checked={selected}
			onClick={onSelect}
			className={cn(
				"flex w-40 flex-col overflow-hidden rounded-lg text-left transition-[box-shadow,outline]",
				selected
					? "ring-2 ring-primary ring-offset-2 ring-offset-background"
					: "hover:ring-1 hover:ring-border",
			)}
		>
			<span
				className={cn(
					"flex h-24 flex-col gap-1 p-2",
					appearance === "light" ? "bg-zinc-100" : "bg-zinc-900",
				)}
			>
				<span className={cn("h-3 w-full rounded-sm", appearance === "light" ? "bg-white" : "bg-zinc-800")} />
				<span className={cn("h-3 w-3/4 rounded-sm", appearance === "light" ? "bg-white" : "bg-zinc-800")} />
			</span>
			<span
				className={cn(
					"px-2 py-1 font-data text-xs",
					appearance === "light" ? "bg-zinc-200 text-zinc-700" : "bg-zinc-950 text-zinc-300",
				)}
			>
				{label}
			</span>
		</button>
	);
}

function AppearanceSection() {
	const preference = useThemeStore((s) => s.preference);
	const setTheme = useThemeStore((s) => s.setTheme);
	const options = [
		{ id: "system", label: "System", appearance: "dark" as const },
		...THEMES.map((t) => ({ id: t.id, label: t.label, appearance: t.appearance })),
	];
	return (
		<div className="flex flex-col gap-4">
			<SettingsGroup label="Theme">
				<div role="radiogroup" className="flex flex-wrap gap-3">
				{options.map((opt) => {
					// SAFETY: opt.id comes from the THEMES registry plus 'system' —
					// both members of ThemePreference.
					const themeId = opt.id as typeof preference;
					return (
						<ThemeSwatch
							key={opt.id}
							label={opt.label}
							appearance={opt.appearance}
							selected={preference === opt.id}
							onSelect={() => setTheme(themeId)}
						/>
					);
				})}
				</div>
				<p className="flex items-center gap-1.5 text-xs text-muted-foreground/70">
					<SunMoon className="size-3" aria-hidden />
					System follows the operating system setting.
				</p>
			</SettingsGroup>
			<SettingsGroup label="Accent">
				<p className="flex items-center gap-2 rounded-lg border border-primary/25 bg-primary/5 px-3 py-2 font-data text-xs text-primary">
					<Lock className="size-3.5 shrink-0" aria-hidden />
					Reqly coral — brand accent locked across themes
				</p>
			</SettingsGroup>
		</div>
	);
}

function EditorSection() {
	const responseMode = useShellStore((s) => s.responseMode);
	const setResponseMode = useShellStore((s) => s.setResponseMode);
	const modes: { mode: ResponseMode; label: string; icon: typeof Columns2 }[] = [
		{ mode: "split", label: "Beside the editor", icon: Columns2 },
		{ mode: "inline", label: "Below the editor", icon: Rows2 },
	];
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
							"flex items-center gap-2 rounded-lg border px-3 py-2 text-left text-xs transition-colors",
							responseMode === mode ? "border-ring bg-muted/60" : "border-border hover:bg-muted/40",
						)}
					>
						<Icon className="size-4 shrink-0 text-muted-foreground" aria-hidden />
						<span className="font-medium text-foreground">{label}</span>
					</button>
				))}
			</div>
		</SettingsGroup>
	);
}

function PrivacySection() {
	return (
		<div className="flex flex-col gap-2 text-xs text-muted-foreground">
			<p className="flex items-center gap-2 text-sm font-medium text-status-ok">
				<ShieldCheck className="size-4" aria-hidden />
				Zero telemetry · local-first
			</p>
			<p>
				Reqly ships with no analytics, no crash reporting, and no accounts.
				Requests, responses, headers, credentials, and history never leave
				your machine — traffic goes straight from the app to the APIs you
				call.
			</p>
			<p>History and cookies live in a local SQLite database inside the workspace.</p>
		</div>
	);
}

function StorageSection() {
	const workspace = useWorkspaceStore((s) => s.currentWorkspace);
	const historyCount = useHistoryStore((s) => s.pool.length);
	const loadPool = useHistoryStore((s) => s.loadPool);
	const poolLoaded = useHistoryStore((s) => s.poolLoaded);
	useEffect(() => {
		if (!poolLoaded) void loadPool();
	}, [poolLoaded, loadPool]);
	return (
		<dl className="grid max-w-md grid-cols-[10rem_1fr] gap-x-3 gap-y-1.5 font-data text-xs">
			<dt className="text-muted-foreground">WORKSPACE</dt>
			<dd className="truncate text-foreground">{workspace?.path ?? "none open"}</dd>
			<dt className="text-muted-foreground">FORMATS</dt>
			<dd className="text-foreground">JSON/YAML collections · Git-versioned</dd>
			<dt className="text-muted-foreground">HISTORY ENTRIES</dt>
			<dd className="text-foreground tabular-nums">{historyCount}</dd>
		</dl>
	);
}

const SHORTCUTS = [
	{ keys: "⌘K", label: "Open the command palette" },
	{ keys: "⌘S", label: "Save the open request" },
	{ keys: "⌘↩", label: "Send the open request" },
];

function ShortcutsSection() {
	return (
		<ul className="flex max-w-md flex-col gap-1">
			{SHORTCUTS.map((s) => (
				<li
					key={s.keys}
					className="flex items-center justify-between rounded-lg border border-border px-3 py-2 text-xs"
				>
					<span className="text-muted-foreground">{s.label}</span>
					<kbd className="rounded border border-border bg-muted/40 px-1.5 py-0.5 font-data text-xs">
						{s.keys}
					</kbd>
				</li>
			))}
		</ul>
	);
}

function AboutSection() {
	const workspace = useWorkspaceStore((s) => s.currentWorkspace);
	return (
		<div className="flex max-w-md flex-col gap-2 text-xs text-muted-foreground">
			<p className="flex items-center gap-2 text-sm font-medium text-foreground">
				<GitBranch className="size-4" aria-hidden />
				Reqly — local-first, Git-native API client
			</p>
			<p>
				Desktop shell: Wails v3 · Core: Go · Collections, environments, and
				tests are plain-text files versioned with Git in{" "}
				<span className="font-data">{workspace?.path ?? "your workspace"}</span>.
			</p>
		</div>
	);
}

/** Settings view (M44 T6, restyled G-17.4.7): settings nav beside per-section
 * content — appearance (theme cards + locked accent), editor layout,
 * privacy, storage, shortcuts, about. */
export function SettingsView() {
	const [section, setSection] = useState<SettingsSectionId>("appearance");
	return (
		<ViewShell label="Settings">
			<SplitView
				asideLabel="Settings sections"
				aside={
						<nav aria-label="Settings sections" className="flex flex-col gap-0.5">
					<p className="px-2 pb-1 font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
						Settings
					</p>
					{SECTIONS.map((s) => (
						<button
							key={s.id}
							type="button"
							onClick={() => setSection(s.id)}
							aria-current={section === s.id ? "page" : undefined}
							className={cn(
								"flex items-center gap-2 rounded-lg px-2.5 py-1.5 text-left text-xs transition-colors",
								section === s.id
									? "bg-primary/10 font-medium text-primary"
									: "text-muted-foreground hover:bg-accent hover:text-foreground",
							)}
						>
							{s.id === "appearance" ? (
								<Eye className="size-3.5" aria-hidden />
							) : s.id === "shortcuts" ? (
								<Keyboard className="size-3.5" aria-hidden />
							) : s.id === "about" ? (
								<Info className="size-3.5" aria-hidden />
							) : (
								<MonitorCog className="size-3.5" aria-hidden />
							)}
							{s.label}
						</button>
					))}
					</nav>
				}
			>
			<div className="min-w-0 flex-1 overflow-y-auto">
				<div className="max-w-xl border-t border-border pt-4">
					{section === "appearance" ? <AppearanceSection /> : null}
					{section === "editor" ? <EditorSection /> : null}
					{section === "privacy" ? <PrivacySection /> : null}
					{section === "storage" ? <StorageSection /> : null}
					{section === "shortcuts" ? <ShortcutsSection /> : null}
					{section === "about" ? <AboutSection /> : null}
				</div>
			</div>
			</SplitView>
		</ViewShell>
	);
}
