import { create } from "zustand";
import type { BodyType } from "../lib/body";
import {
	type CollectionsAdapter,
	fallbackCollectionsAdapter,
	type ResolvedRequestInput,
	type WorkspaceTree,
} from "../lib/collections";
import { type EnvAdapter, fallbackEnvAdapter } from "../lib/env";
import {
	NEW_REQUEST_TAB_ID,
	type TabDraft,
	useRequestStore,
} from "./useRequestStore";

export interface Workspace {
	id: string;
	name: string;
	path: string;
}

export interface Environment {
	id: string;
	name: string;
	description: string;
	variables: Record<string, string>;
	secrets: string[];
}

export interface RequestTab {
	id: string;
	title: string;
	/** Workspace-relative Request Path when opened from a collection. */
	requestPath?: string;
	/** Tab kind: "request" (default) renders the request editor; "run" renders
	 * the collection Run View. */
	kind?: "request" | "run";
}

export type WorkspaceView = "requests" | "environments";

interface WorkspaceState {
	currentWorkspace: Workspace | null;
	selectedCollectionId: string | null;
	openTabs: RequestTab[];
	activeTabId: string | null;
	activeView: WorkspaceView;
	activeEnvironmentId: string | null;
	environments: Environment[];
	environmentsError: string | null;
	envAdapter: EnvAdapter;
	workspaceTree: WorkspaceTree | null;
	workspaceError: string | null;
	workspaceAdapter: CollectionsAdapter;
	expanded: Record<string, boolean>;
	dirtyEditors: Record<string, boolean>;
	hasUnsavedEnvChanges: boolean;
	/** Transient error opening a specific request; never replaces the tree. */
	openError: string | null;

	setCurrentWorkspace: (workspace: Workspace | null) => void;
	selectCollection: (id: string | null) => void;
	setActiveView: (view: WorkspaceView) => void;
	setEditorDirty: (key: string, dirty: boolean) => void;
	openTab: (
		tab: RequestTab,
		seed?: Partial<import("./useRequestStore").TabDraft>,
	) => void;
	closeTab: (id: string) => void;
	setActiveTab: (id: string | null) => void;
	/** Open a collection request by Request Path into a tab, seeding the draft
	 * from its resolved form and recording its variable chain + env pill. */
	openRequest: (path: string) => Promise<void>;
	setActiveEnvironment: (id: string | null) => void;
	setEnvironments: (environments: Environment[]) => void;
	setEnvAdapter: (adapter: EnvAdapter) => void;
	setWorkspaceAdapter: (adapter: CollectionsAdapter) => void;
	refreshWorkspace: () => Promise<void>;
	toggleExpanded: (path: string) => void;
	refreshEnvironments: () => Promise<void>;
}

const toEnvironment = (
	name: string,
	src: {
		description?: string;
		variables?: Record<string, string>;
		secrets?: string[];
	},
): Environment => ({
	id: name,
	name,
	description: src.description ?? "",
	variables: src.variables ?? {},
	secrets: src.secrets ?? [],
});

/** bodyTypeFor infers the editor's body type from an opened request's body and
 * Content-Type header, so what the tab shows matches what the core will send. */
export const bodyTypeFor = (req: ResolvedRequestInput): BodyType => {
	if (!req.body) return "none";
	const contentType = req.headers
		.find((h) => h.key.toLowerCase() === "content-type")
		?.value.toLowerCase();
	if (contentType?.includes("json")) return "json";
	if (contentType?.includes("xml")) return "xml";
	if (contentType?.includes("multipart/form-data")) return "form-data";
	if (contentType?.includes("urlencoded")) return "urlencoded";
	return "raw";
};

/** draftFromOpened maps an opened request's resolved fields onto the editor
 * draft shape, preserving placeholders for send-time interpolation. */
export const draftFromOpened = (opened: {
	request: ResolvedRequestInput;
}): Partial<TabDraft> => {
	const toRows = (rows: { key: string; value: string }[]) =>
		rows.map(({ key, value }) => ({ key, value, enabled: true }));
	return {
		method: opened.request.method,
		url: opened.request.url,
		params: toRows(opened.request.query),
		headers: toRows(opened.request.headers),
		bodyType: bodyTypeFor(opened.request),
		body: opened.request.body,
	};
};

export const useWorkspaceStore = create<WorkspaceState>((set, get) => ({
	currentWorkspace: null,
	selectedCollectionId: null,
	openTabs: [],
	activeTabId: null,
	activeView: "requests",
	activeEnvironmentId: null,
	environments: [],
	environmentsError: null,
	envAdapter: fallbackEnvAdapter,
	workspaceTree: null,
	workspaceError: null,
	workspaceAdapter: fallbackCollectionsAdapter,
	expanded: {},
	dirtyEditors: {},
	hasUnsavedEnvChanges: false,
	openError: null,

	setCurrentWorkspace: (currentWorkspace) => set({ currentWorkspace }),
	selectCollection: (selectedCollectionId) => set({ selectedCollectionId }),
	setActiveView: (activeView) => set({ activeView }),
	setEditorDirty: (key, dirty) =>
		set((state) => {
			const dirtyEditors = { ...state.dirtyEditors, [key]: dirty };
			return {
				dirtyEditors,
				hasUnsavedEnvChanges: Object.values(dirtyEditors).some(Boolean),
			};
		}),

	openTab: (tab, seed) =>
		set((state) => {
			const exists = state.openTabs.some((t) => t.id === tab.id);
			const openTabs = exists ? state.openTabs : [...state.openTabs, tab];
			if (!exists && tab.kind !== "run") {
				useRequestStore.getState().ensureDraft(tab.id, seed);
			}
			return {
				openTabs,
				activeTabId: tab.id,
			};
		}),

	closeTab: (id) =>
		set((state) => {
			const index = state.openTabs.findIndex((t) => t.id === id);
			let openTabs = state.openTabs.filter((t) => t.id !== id);
			useRequestStore.getState().removeTab(id);
			// Closing the last tab restores the default scratchpad.
			if (openTabs.length === 0) {
				openTabs = [{ id: NEW_REQUEST_TAB_ID, title: "New Request" }];
				useRequestStore.getState().ensureDraft(NEW_REQUEST_TAB_ID);
			}
			const activeTabId =
				state.activeTabId === id
					? (openTabs[Math.max(0, index - 1)]?.id ?? null)
					: state.activeTabId;
			return { openTabs, activeTabId };
		}),

	setActiveTab: (activeTabId) => {
		if (activeTabId) {
			const tab = get().openTabs.find((t) => t.id === activeTabId);
			if (!tab || tab.kind !== "run") {
				useRequestStore.getState().ensureDraft(activeTabId);
			}
		}
		set({ activeTabId });
	},

	openRequest: async (path) => {
		const { workspaceAdapter } = get();
		try {
			const opened = await workspaceAdapter.open(path);
			set({ openError: null });
			get().openTab(
				{ id: opened.path, title: opened.name, requestPath: opened.path },
				draftFromOpened(opened),
			);
			useRequestStore.getState().setMeta(opened.path, {
				requestPath: opened.path,
				name: opened.name,
				variables: opened.variables,
				env: opened.fileEnv || undefined,
				auth: opened.request.auth,
			});
		} catch (err) {
			set({ openError: err instanceof Error ? err.message : String(err) });
		}
	},

	setActiveEnvironment: (activeEnvironmentId) => set({ activeEnvironmentId }),

	setEnvironments: (environments) => set({ environments }),

	setEnvAdapter: (envAdapter) => set({ envAdapter }),

	setWorkspaceAdapter: (workspaceAdapter) => set({ workspaceAdapter }),

	refreshWorkspace: async () => {
		const { workspaceAdapter } = get();
		try {
			const tree = await workspaceAdapter.load();
			set({
				workspaceTree: tree,
				workspaceError: null,
				openError: null,
				currentWorkspace: tree.name
					? { id: tree.path, name: tree.name, path: tree.path }
					: null,
			});
		} catch (err) {
			set({ workspaceError: err instanceof Error ? err.message : String(err) });
		}
	},

	toggleExpanded: (path) =>
		set((state) => ({
			expanded: { ...state.expanded, [path]: !state.expanded[path] },
		})),

	refreshEnvironments: async () => {
		const { envAdapter } = get();
		try {
			const data = await envAdapter.list();
			set({
				environments: data.environments.map((e) => toEnvironment(e.name, e)),
				activeEnvironmentId: data.active || null,
				environmentsError: null,
			});
		} catch (err) {
			set({
				environmentsError: err instanceof Error ? err.message : String(err),
			});
		}
	},
}));
