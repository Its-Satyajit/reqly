import { FilePlus, ListChecks } from "lucide-react";
import { Button } from "#components/ui/button";
import { useTestStore } from "#stores/useTestStore";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";

/** TestsView is the G-17.4.15 full-page home for *.reqly-test files,
 * replacing the sidebar TESTS strip (user directive). */
export function TestsView() {
	const tests = useTestStore((s) => s.tests);
	const openPath = useTestStore((s) => s.openPath);
	const newTab = useTestStore((s) => s.newTab);
	const openTab = useWorkspaceStore((s) => s.openTab);
	const setActiveView = useWorkspaceStore((s) => s.setActiveView);

	const openTestTab = (path: string | null, title: string, fresh: boolean) => {
		const id = fresh ? `test-new-${Date.now()}` : `test-${path}`;
		if (fresh) newTab(id);
		else void openPath(id, path ?? "");
		openTab({ id, title, kind: "test", filePath: path ?? undefined });
		setActiveView("requests");
	};

	return (
		<section
			className="flex h-full min-h-0 flex-col gap-3 overflow-y-auto p-4"
			aria-label="Tests"
		>
			<div className="flex items-center justify-between">
				<h2 className="text-sm font-semibold">Tests</h2>
				<Button
					size="sm"
					onClick={() => openTestTab(null, "untitled.reqly-test", true)}
				>
					<FilePlus data-icon="inline-start" />
					New test file
				</Button>
			</div>
			{tests.length === 0 ? (
				<div className="flex flex-1 flex-col items-center justify-center gap-1.5 rounded-xl border border-border bg-card">
					<ListChecks className="size-8 text-muted-foreground/40" aria-hidden />
					<p className="text-sm font-medium text-foreground">No test files yet</p>
					<p className="max-w-xs text-center text-xs text-muted-foreground">
						Create a <code className="font-mono">*.reqly-test</code> file to
						assert against responses — they run through the same pipeline as{" "}
						<code className="font-mono">reqly test</code>.
					</p>
				</div>
			) : (
				<ul className="flex flex-col gap-1">
					{tests.map((t) => (
						<li key={t.path}>
							<button
								type="button"
								className="w-full rounded-lg border border-border bg-card px-3 py-2 text-left transition-colors hover:bg-accent"
								title={t.path}
								onClick={() => openTestTab(t.path, t.name || t.path, false)}
							>
								<span className="block truncate text-xs font-medium text-foreground">
									{t.name || t.path}
								</span>
								<span className="block truncate font-mono text-[11px] text-muted-foreground">
									{t.path}
								</span>
							</button>
						</li>
					))}
				</ul>
			)}
		</section>
	);
}
