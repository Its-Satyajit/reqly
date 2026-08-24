import { useEffect, useState } from "react";
import { GitBranch, Loader2, RotateCcw } from "lucide-react";

import { cn } from "#lib/utils";
import type { GitFileStatus } from "#lib/gitclient";
import { useGitStore } from "#stores";
import { Button } from "./ui/button";

/** Porcelain code → display glyph; unknown codes render verbatim. */
function glyphCode(code: string): string {
	return code === "?" ? "?" : code === " " ? "" : code;
}

function glyphOf(f: GitFileStatus): string {
	if (f.x === "?" && f.y === "?") return "?";
	return glyphCode(f.x) + glyphCode(f.y);
}

/** Status color ramp — semantic tokens, never color alone (glyph + text). */
function statusTone(f: GitFileStatus): string {
	const code = f.x !== " " ? f.x : f.y;
	switch (code) {
		case "A":
		case "?":
			return "text-status-ok";
		case "U":
		case "B":
			return "text-status-error";
		case "M":
		case "R":
		case "C":
			return "text-status-warn";
		case "D":
			return "text-status-error";
		default:
			return "text-muted-foreground";
	}
}

/**
 * Pending-changes strip inside the collections sidebar (M44 T4):
 * branch chip, per-file status glyphs, stage/unstage on click, commit box.
 */
export function GitPanel() {
	const status = useGitStore((s) => s.status);
	const loading = useGitStore((s) => s.loading);
	const error = useGitStore((s) => s.error);
	const refresh = useGitStore((s) => s.refresh);
	const stage = useGitStore((s) => s.stage);
	const unstage = useGitStore((s) => s.unstage);
	const commit = useGitStore((s) => s.commit);

	const [message, setMessage] = useState("");

	useEffect(() => {
		void refresh();
		// Keep status live while the app is open: poll lightly and refresh on focus.
		const interval = setInterval(() => void refresh(), 4000);
		const onFocus = () => void refresh();
		window.addEventListener("focus", onFocus);
		return () => {
			clearInterval(interval);
			window.removeEventListener("focus", onFocus);
		};
	}, [refresh]);

	if (!status?.repoFound) return null;

	const files = status.files ?? [];
	const staged = files.flatMap((f) => (f.staged ? [f.path] : []));
	const canCommit = staged.length > 0 && message.trim().length > 0;

	const toggle = (f: GitFileStatus) => {
		if (f.staged) void unstage([f.path]);
		else void stage([f.path]);
	};

	const onCommit = async () => {
		if (await commit(message.trim())) setMessage("");
	};

	return (
		<section aria-label="Source control" className="flex flex-col gap-1 border-t border-border pt-2">
			<div className="flex items-center justify-between px-2">
				<span className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground">
					<GitBranch className="size-3.5" aria-hidden />
					<span className="font-data">{status.branch || "(no branch)"}</span>
				</span>
				<Button
					variant="ghost"
					size="icon-xs"
					onClick={() => void refresh()}
					aria-label="Refresh git status"
					title="Refresh"
				>
					{loading ? (
						<Loader2 className="size-3.5 animate-spin" aria-hidden />
					) : (
						<RotateCcw className="size-3.5" aria-hidden />
					)}
				</Button>
			</div>

			{error && (
				<p role="alert" className="px-2 py-1 text-xs text-destructive">
					{error}
				</p>
			)}

			{files.length === 0 ? (
				<p className="px-2 pb-1 text-xs text-muted-foreground">
					Working tree clean
				</p>
			) : (
				<ul className="max-h-44 overflow-y-auto" aria-label="Changed files">
					{files.map((f) => (
						<li key={f.path}>
							<button
								type="button"
								onClick={() => toggle(f)}
								className={cn(
									"flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-xs hover:bg-muted/50",
									f.staged && "bg-muted/30",
								)}
								title={f.staged ? "Click to unstage" : "Click to stage"}
							>
								<span
									aria-hidden
									className={cn(
										"flex size-3 shrink-0 items-center justify-center rounded-[3px] border border-input",
										f.staged && "bg-primary text-primary-foreground",
									)}
								>
									{f.staged ? "✓" : ""}
								</span>
								<span
									className={cn("font-data shrink-0 font-medium", statusTone(f))}
								>
									{glyphOf(f)}
								</span>
								<span className="truncate">{f.path}</span>
							</button>
						</li>
					))}
				</ul>
			)}

			{files.length > 0 && (
				<div className="flex flex-col gap-1 px-1 pb-1">
					<textarea
						value={message}
						onChange={(e) => setMessage(e.target.value)}
						placeholder={`Commit ${staged.length} staged file(s)…`}
						aria-label="Commit message"
						rows={2}
						className="resize-none rounded-md border border-input bg-input-bg px-2 py-1 text-xs outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/50"
					/>
					<Button size="sm" disabled={!canCommit} onClick={() => void onCommit()}>
						Commit{staged.length > 0 ? ` (${staged.length})` : ""}
					</Button>
				</div>
			)}
		</section>
	);
}
