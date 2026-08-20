import { create } from "zustand";
import type {
	CollectionsAdapter,
	RunEvent,
	RunReport,
	RunStep,
} from "../lib/collections";
import { useWorkspaceStore } from "./useWorkspaceStore";

/** collectionRunAdapter resolves the active CollectionsAdapter, which carries
 * the run/cancelRun contract, from the workspace store where the host bridge
 * registers it. */
const collectionRunAdapter = (): CollectionsAdapter =>
	useWorkspaceStore.getState().workspaceAdapter;

interface CollectionRunState {
	/** Active run id, null when idle. */
	runId: string | null;
	/** The collection/folder Request Path being run (used as the tab title). */
	path: string | null;
	failFast: boolean;
	steps: RunStep[];
	report: RunReport | null;
	error: string | null;
	running: boolean;

	/** startRun begins a collection/folder run and streams its events into the
	 * store until it reports done or fails. */
	startRun: (path: string, env: string | null, failFast: boolean) => Promise<void>;
	/** cancelRun aborts the in-flight run; its late events still settle the
	 * store so the UI shows the final state. */
	cancelRun: () => Promise<void>;
	reset: () => void;
}

export const useCollectionRunStore = create<CollectionRunState>((set, get) => ({
	runId: null,
	path: null,
	failFast: false,
	steps: [],
	report: null,
	error: null,
	running: false,

	startRun: async (path, env, failFast) => {
		set({
			runId: null,
			path,
			failFast,
			steps: [],
			report: null,
			error: null,
			running: true,
		});
		const onEvent = (event: RunEvent) => {
			const state = get();
			switch (event.type) {
				case "step":
					set({ steps: [...state.steps, event.step] });
					break;
				case "done":
					set({
						report: event.report,
						running: false,
						runId: null,
						path: null,
					});
					break;
				case "error":
					set({
						error: event.message,
						running: false,
						runId: null,
						path: null,
					});
					break;
			}
		};
		try {
			const id = await collectionRunAdapter().run(path, env, failFast, onEvent);
			set({ runId: id });
		} catch (err) {
			set({
				error: err instanceof Error ? err.message : String(err),
				running: false,
				path: null,
			});
		}
	},

	cancelRun: async () => {
		const { runId } = get();
		if (!runId) return;
		try {
			await collectionRunAdapter().cancelRun(runId);
		} catch (err) {
			set({ error: err instanceof Error ? err.message : String(err) });
		}
	},

	reset: () =>
		set({
			runId: null,
			path: null,
			failFast: false,
			steps: [],
			report: null,
			error: null,
			running: false,
		}),
}));