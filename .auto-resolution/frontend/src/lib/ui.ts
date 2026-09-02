/** Shared class helpers for consistent chrome across feature panels
 * (G-4.2.x — single source instead of per-file copies). */

/** Tab-pill styling shared by the request/response section tabs. */
export const tabClass = (active: boolean) =>
	`rounded-md px-2 py-1 text-xs font-medium transition-colors ${
		active
			? "bg-muted text-foreground"
			: "text-muted-foreground hover:text-foreground"
	}`;

/** Text-input styling shared across editors and forms (focus ring included). */
export const inputClass =
	"rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus-visible:border-ring";

/** formatBytes renders a byte count as a compact human string. */
export function formatBytes(size: number): string {
	if (size < 1024) return `${size} B`;
	if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
	return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

/** Arrow-key navigation for a roving-focus tablist (G-4.3.3): ArrowLeft/
 * ArrowRight move focus and activate the neighbouring tab, Home/End jump to
 * the first/last. Attach to the element carrying role="tablist". */
export function handleTabArrowKeys(
	e: React.KeyboardEvent<HTMLElement>,
): void {
	const list = e.currentTarget;
	const tabs = Array.from(
		list.querySelectorAll<HTMLButtonElement>('[role="tab"]'),
	);
	if (tabs.length === 0) return;
	// SAFETY: document.activeElement is a tab button while the handler runs —
	// focus lives inside this tablist because the event originated there.
	const current = tabs.indexOf(document.activeElement as HTMLButtonElement);
	let next = -1;
	if (e.key === "ArrowRight") next = (current + 1 + tabs.length) % tabs.length;
	else if (e.key === "ArrowLeft") next = (current - 1 + tabs.length) % tabs.length;
	else if (e.key === "Home") next = 0;
	else if (e.key === "End") next = tabs.length - 1;
	else return;
	e.preventDefault();
	tabs[next]?.focus();
	tabs[next]?.click();
}
