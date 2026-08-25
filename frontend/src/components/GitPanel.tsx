import { useEffect, useState } from "react";
import {
	GitBranch,
	GitCommitHorizontal,
	Loader2,
	Plus,
	RotateCcw,
	ShieldCheck,
	TriangleAlert,
	X,
} from "lucide-react";

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
	const conflicts = useGitStore((s) => s.conflicts);
	const worktrees = useGitStore((s) => s.worktrees);
	const resolveSide = useGitStore((s) => s.resolveSide);
	const mergeAbort = useGitStore((s) => s.mergeAbort);
	const removeWorktree = useGitStore((s) => s.removeWorktree);
	const recentCommits = useGitStore((s) => s.recentCommits);

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

	if (!status?.repoFound) {
		return (
			<div className="flex h-full items-center justify-center p-4">
				<p className="text-xs text-muted-foreground">
					Not a git repository — collections still save as plain files.
				</p>
			</div>
		);
	}

	const files = status.files ?? [];
	const stagedFiles = files.filter((f) => f.staged);
	const unstagedFiles = files.filter((f) => !f.staged);
	const staged = stagedFiles.map((f) => f.path);
	const canCommit = staged.length > 0 && message.trim().length > 0 && conflicts.length === 0;

	const stageAll = () => void stage(unstagedFiles.map((f) => f.path));

	const onCommit = async () => {
		if (await commit(message.trim())) setMessage("");
	};

	const fileRow = (f: GitFileStatus, action: "stage" | "unstage") => (
		<li key={f.path} className="flex items-center gap-2 rounded-md px-2 py-1 text-xs hover:bg-muted/50">
			<span aria-hidden className={cn("font-data shrink-0 font-medium", statusTone(f))}>
				{glyphOf(f)}
			</span>
			<span className="min-w-0 flex-1 truncate font-data" title={f.path}>
				{f.path}
			</span>
			{(f.adds > 0 || f.dels > 0) && (
				<span className="flex shrink-0 gap-1 font-data text-[10px] tabular-nums">
					{f.adds > 0 && (
						<span className="text-status-ok" title={`${f.adds} added`}>
							+{f.adds}
						</span>
					)}
					{f.dels > 0 && (
						<span className="text-status-error" title={`${f.dels} removed`}>
							−{f.dels}
						</span>
					)}
				</span>
			)}
			<Button
				variant="ghost"
				size="icon-xs"
				aria-label={action === "stage" ? `Stage ${f.path}` : `Unstage ${f.path}`}
				onClick={() => (action === "stage" ? void stage([f.path]) : void unstage([f.path]))}
			>
				{action === "stage" ? <Plus className="size-3.5" /> : <X className="size-3.5" />}
			</Button>
		</li>
	);

	return (
		<div className="flex h-full min-h-0 flex-col gap-3 lg:flex-row" aria-label="Source control">
			<div className="flex w-full max-w-sm shrink-0 flex-col gap-2 rounded-xl border border-border bg-card p-3">
				<div className="flex items-center justify-between">
					<p className="font-data text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
						Working tree {files.length > 0 ? `(${files.length})` : ""}
					</p>
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
					<p role="alert" className="text-xs text-destructive">
						{error}
					</p>
				)}

				{conflicts.length > 0 && (
					<div className="flex flex-col gap-1 rounded-lg border border-status-error/40 bg-status-error/10 p-2" role="alert">
						<p className="flex items-center gap-1.5 text-xs font-medium text-status-error">
							<TriangleAlert className="size-3.5" aria-hidden />
							UNMERGED — resolve before commit
						</p>
						{conflicts.map((c) => (
							<div key={c.path} className="flex items-center justify-between gap-2">
								<span className="truncate font-data text-xs">{c.path}</span>
								<span className="flex shrink-0 gap-1">
									<Button variant="outline" size="xs" onClick={() => void resolveSide(c.path, "ours")}>
										Ours
									</Button>
									<Button variant="outline" size="xs" onClick={() => void resolveSide(c.path, "theirs")}>
										Theirs
									</Button>
								</span>
							</div>
						))}
						<Button variant="outline" size="xs" className="self-start" onClick={() => void mergeAbort()}>
							Abort merge
						</Button>
					</div>
				)}

				{files.length === 0 ? (
					<p className="py-2 text-xs text-muted-foreground">Working tree clean</p>
				) : (
					<div className="min-h-0 flex-1 overflow-y-auto">
						{stagedFiles.length > 0 ? (
							<div className="flex flex-col gap-0.5">
								<div className="flex items-center justify-between px-2">
									<p className="font-data text-[10px] uppercase tracking-widest text-muted-foreground">
										Staged ({stagedFiles.length})
									</p>
								</div>
								<ul>{stagedFiles.map((f) => fileRow(f, "unstage"))}</ul>
							</div>
						) : null}
						{unstagedFiles.length > 0 ? (
							<div className="flex flex-col gap-0.5 pt-1">
								<div className="flex items-center justify-between px-2">
									<p className="font-data text-[10px] uppercase tracking-widest text-muted-foreground">
										Unstaged ({unstagedFiles.length})
									</p>
									<Button variant="ghost" size="xs" onClick={stageAll}>
										Stage all
									</Button>
								</div>
								<ul>{unstagedFiles.map((f) => fileRow(f, "stage"))}</ul>
							</div>
						) : null}
					</div>
				)}

				<div className="flex shrink-0 flex-col gap-1.5 border-t border-border pt-2">
					<textarea
						value={message}
						onChange={(e) => setMessage(e.target.value)}
						placeholder="Commit message (feat(scope): summary)"
						aria-label="Commit message"
						rows={2}
						className="resize-none rounded-md border border-input bg-background px-2 py-1.5 text-xs text-foreground placeholder:text-muted-foreground"
					/>
					<Button
						variant="destructive"
						size="sm"
						disabled={!canCommit}
						onClick={() => void onCommit()}
						title={conflicts.length > 0 ? "Resolve conflicts before committing" : undefined}
					>
						<GitCommitHorizontal data-icon="inline-start" />
						Commit {staged.length > 0 ? `${staged.length} file${staged.length === 1 ? "" : "s"}` : ""}
					</Button>
				</div>
			</div>

			<div className="flex min-w-0 flex-1 flex-col gap-3 overflow-y-auto">
				<div className="flex flex-col gap-1 rounded-xl border border-border bg-card p-3">
					<p className="font-data text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
						Recent commits — {status.branch || "HEAD"}
					</p>
					{recentCommits.length === 0 ? (
						<p className="text-xs text-muted-foreground">No commits yet.</p>
					) : (
						<ul className="flex flex-col gap-0.5">
							{recentCommits.map((t) => (
								<li key={t.hash} className="flex items-center gap-2 text-xs">
									<Hash hash={t.hash} />
									<span className="min-w-0 flex-1 truncate text-foreground">{t.subject}</span>
								</li>
							))}
						</ul>
					)}
				</div>

				<div className="flex flex-col gap-1 rounded-xl border border-border bg-card p-3">
					<p className="font-data text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
						Branch
					</p>
					<p className="flex items-center gap-1.5 font-data text-xs text-foreground">
						<GitBranch className="size-3.5" aria-hidden />
						{status.branch || "(no branch)"}
					</p>
					{worktrees.length > 1 ? (
						<details className="text-xs text-muted-foreground">
							<summary className="cursor-pointer select-none py-0.5">
								Worktrees ({worktrees.length})
							</summary>
							<ul className="flex flex-col gap-0.5 pt-1">
								{worktrees.map((t) => (
									<li key={t.path} className="flex items-center justify-between gap-2">
										<span className="truncate font-data" title={t.path}>
											{t.branch || t.path}
										</span>
										{!t.isCurrent && !t.isBare ? (
											<Button
												variant="ghost"
												size="icon-xs"
												aria-label={`Remove worktree ${t.path}`}
												onClick={() => void removeWorktree(t.path)}
											>
												<X className="size-3" />
											</Button>
										) : null}
									</li>
								))}
							</ul>
						</details>
					) : null}
				</div>

				<div className="flex items-start gap-2 rounded-xl border border-border bg-card p-3 text-[11px] text-muted-foreground">
					<ShieldCheck className="mt-0.5 size-3.5 shrink-0" aria-hidden />
					Everything here reads the plain-text worktree. No network calls unless
					you push from outside — Reqly never talks to remotes.
				</div>
			</div>
		</div>
	);
}

/** Shell commit strip (M44 T7): staged summary + recent commits. */
export function CommitStrip() {
	const status = useGitStore((s) => s.status)
	const recentCommits = useGitStore((s) => s.recentCommits)
	const conflicts = useGitStore((s) => s.conflicts)

	if (!status?.repoFound) return null

	const files = status.files ?? []
	const stagedCount = files.filter((f) => f.staged).length
	const latest = recentCommits[0]

	return (
		<div className="flex h-6 shrink-0 items-center justify-between border-t border-border px-3 text-xs text-muted-foreground">
			<span>
				{files.length > 0 ? (
					<>
						<span className="text-status-warn">{stagedCount}</span> of{" "}
						{files.length} changed file(s) staged
					</>
				) : (
					"Working tree clean"
				)}
			</span>
			{conflicts.length > 0 && (
				<span className="font-medium text-status-error">
					{conflicts.length} conflict(s)
				</span>
			)}
			{latest && (
				<span className="font-data truncate" title={latest.subject}>
					<Hash hash={latest.hash} /> {latest.subject}
				</span>
			)}
		</div>
	)
}

function Hash({ hash, className }: { hash: string; className?: string }) {
	return <span className={cn("mr-1 text-primary", className)}>{hash}</span>
}
