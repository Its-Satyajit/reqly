import { create } from "zustand";
import type {
	CollectionsAdapter,
	RunEvent,
	RunReport,
	RunStep,
} from "../lib/collections";
import { addBreadcrumb } from "../lib/crash";
import { useWorkspaceStore, registerTabCloseListener } from "./useWorkspaceStore";

/** RUN_TAB_ID is the fixed tab id of the Run View (only one run happens at a
 * time, so a single persistent run tab is enough). */
export const RUN_TAB_ID = "run";

/** collectionRunAdapter resolves the active CollectionsAdapter, which carries
 * the run/cancelRun contract, from the workspace store where the host bridge
 * registers it. */
const collectionRunAdapter = (): CollectionsAdapter =>
	useWorkspaceStore.getState().workspaceAdapter;

interface CollectionRunState {
	/** Active run id, null when idle. */
	runId: string | null;
	/** The collection/folder Request Path being (or last) run — kept after
	 * completion so the report stays visible until the next run starts. */
	path: string | null;
	/** The environment pill used for the run ("" for the workspace default). */
	env: string | null;
	failFast: boolean;
	steps: RunStep[];
	report: RunReport | null;
	error: string | null;
	running: boolean;
	/** Bumped on every startRun; events from older generations are dropped so
	 * a superseded run's late steps/done never mutate the current one. */
	generation: number;

	startRun: (path: string, env: string | null, failFast: boolean) => Promise<void>;
	cancelRun: () => Promise<void>;
	/** exportReport serializes the finished run as "json" or "junit"; returns
	 * the export (path + rendered text) or null when there is nothing to
	 * export / the adapter does not support it. */
	exportReport: (
		format: "json" | "junit",
	) => Promise<{ format: string; path: string; content: string } | null>;
	toggleFailFast: () => void;
	reset: () => void;
}

export const useCollectionRunStore = create<CollectionRunState>((set, get) => ({
	runId: null,
	path: null,
	env: null,
	failFast: false,
	steps: [],
	report: null,
	error: null,
	running: false,
	generation: 0,

	startRun: async (path, env, failFast) => {
		const state = get();
		if (state.running && state.runId) {
			set({ error: "A collection run is already in progress." });
			return;
		}
		const generation = state.generation + 1;
		addBreadcrumb("run-start", path);
		set({
			generation,
			runId: null,
			path,
			env,
			failFast,
			steps: [],
			report: null,
			error: null,
			running: true,
		});
		const onEvent = (event: RunEvent) => {
			if (get().generation !== generation) return;
			switch (event.type) {
				case "step":
					set({ steps: [...get().steps, event.step] });
					break;
				case "done":
					set({ report: event.report, running: false, runId: null });
					break;
				case "error":
					set({ error: event.message, running: false, runId: null });
					break;
			}
		};
		try {
			const id = await collectionRunAdapter().run(path, env, failFast, onEvent);
			if (get().generation !== generation) return;
			set({ runId: id });
		} catch (err) {
			if (get().generation !== generation) return;
			set({
				error: err instanceof Error ? err.message : String(err),
				running: false,
			});
		}
	},

	cancelRun: async () => {
		const { runId, generation } = get();
		if (!runId) return;
		addBreadcrumb("run-cancel");
		try {
			await collectionRunAdapter().cancelRun(runId);
		} catch (err) {
			if (get().generation !== generation) return;
			set({ error: err instanceof Error ? err.message : String(err) });
		}
	},

	exportReport: async (format) => {
		const report = get().report;
		if (!report || get().running) return null;
		const adapter = collectionRunAdapter();
		if (!adapter.exportReport) return null;
		return adapter.exportReport(format, {
			path: get().path ?? undefined,
			started: report.started,
			finished: report.finished,
			durationMs: report.durationMs,
			steps: report.steps.map((st) => ({
				name: st.name,
				requestPath: st.requestPath,
				passed: st.passed,
				requestError: st.requestError || undefined,
				durationMs: st.response?.durationMs ?? undefined,
				tests: st.tests.map((t) => ({ name: t.name, passed: t.passed })),
			})),
		});
	},

	toggleFailFast: () => set((state) => ({ failFast: !state.failFast })),

	reset: () =>
		set((state) => ({
			runId: null,
			path: null,
			env: null,
			failFast: state.failFast,
			steps: [],
			report: null,
			error: null,
			running: false,
			generation: state.generation + 1,
		})),
}));

registerTabCloseListener((tab) => {
	if (tab.kind === "run") {
		useCollectionRunStore.getState().reset();
	}
});
