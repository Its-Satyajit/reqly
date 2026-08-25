import { useState } from "react";
import { X } from "lucide-react";
import {
	Dialog,
	DialogContent,
} from "#components/ui/dialog";
import { Button } from "#components/ui/button";
import { cn } from "#lib/utils";

const DISMISS_KEY = "reqly-welcome-dismissed.v1";

/** dismissedBefore reads the persisted don't-show-again flag. Module-level
 * because React Compiler cannot handle try/catch. */
function dismissedBefore(): boolean {
	try {
		return localStorage.getItem(DISMISS_KEY) === "1";
	} catch {
		return false;
	}
}

function persistDismissed(): void {
	try {
		localStorage.setItem(DISMISS_KEY, "1");
	} catch {
		// storage unavailable — session-only dismissal
	}
}

const STEPS = [
	{
		title: "Pick an environment",
		body: (
			<>
				— variables like{" "}
				<code className="rounded bg-muted px-1 font-data text-xs">
					{"{{baseUrl}}"}
				</code>{" "}
				resolve through 6 scopes.
			</>
		),
	},
	{
		title: "Send your first request",
		body: "— open one from Collections and hit Send.",
	},
	{
		title: "Explore the surface",
		body: "— GraphQL, gRPC, WebSocket, SSE, mocks, diffs, tests…",
	},
	{
		title: "Press ⌘K",
		body: "— jump anywhere, or trigger flows from the palette.",
	},
] as const;

/** WelcomeModal is the G-17.4.13 onboarding modal (reference
 * 00-welcome.png): shown once per install over the open workspace, with a
 * don't-show-again flag. */
export function WelcomeModal() {
	// Lazy initializer: no state-setting effect (React Compiler cannot
	// optimize that pattern).
	const [visible, setVisible] = useState(() => !dismissedBefore());
	const [dontShow, setDontShow] = useState(true);

	if (!visible) return null;

	const close = (): void => {
		if (dontShow) persistDismissed();
		setVisible(false);
	};

	return (
		<Dialog open onOpenChange={(next) => { if (!next) close(); }}>
			<DialogContent className="w-full max-w-2xl gap-6 sm:max-w-2xl">
				<button
					type="button"
					aria-label="Close welcome"
					onClick={close}
					className="absolute right-3 top-3 rounded-full border border-border p-1 text-muted-foreground transition-colors hover:text-foreground"
				>
					<X className="size-3.5" aria-hidden />
				</button>
				<div className="grid gap-6 sm:grid-cols-2">
					<div className="flex flex-col gap-4">
						<h2 className="text-base font-semibold text-foreground">
							Welcome to Reqly
						</h2>
						<svg
							aria-hidden
							viewBox="0 0 24 24"
							className="size-10 text-primary"
							fill="none"
							stroke="currentColor"
							strokeWidth="1.6"
						>
							<path d="M12 2 21 7v10l-9 5-9-5V7l9-5Z" />
							<path d="M3.3 7.3 12 12l8.7-4.7M12 12v9.5" />
						</svg>
						<p className="text-sm font-semibold text-foreground">
							Local-first. Git-native. Zero telemetry.
						</p>
						<p className="text-xs text-muted-foreground">
							This workspace mirrors plain-text files on disk. Everything you
							see lives in your repository.
						</p>
					</div>
					<ol className="flex flex-col gap-2.5 text-xs">
						{STEPS.map((step, i) => (
							<li key={step.title} className="flex gap-2">
								<span
									aria-hidden
									className={cn("shrink-0 font-semibold text-foreground")}
								>
									{i + 1}.
								</span>
								<span className="text-muted-foreground">
									<span className="font-semibold text-foreground">
										{step.title}
									</span>{" "}
									{step.body}
								</span>
							</li>
						))}
					</ol>
				</div>
				<div className="flex items-center justify-end gap-3">
					<label className="flex cursor-pointer items-center gap-1.5 text-xs text-muted-foreground">
						<input
							type="checkbox"
							checked={dontShow}
							onChange={(e) => setDontShow(e.target.checked)}
							className="size-3.5 accent-(--primary)"
						/>
						Don&apos;t show again
					</label>
					<Button variant="destructive" size="sm" onClick={close}>
						Start exploring
					</Button>
				</div>
			</DialogContent>
		</Dialog>
	);
}
