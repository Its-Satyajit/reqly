import type {
	AuthAdapter,
	CollectionsAdapter,
	EnvAdapter,
	RequestInput,
	RequestSender,
	ResponseData,
	RunReport,
	RunStep,
	RunTestResult,
} from "@reqly/frontend";
import {
	serializeBody,
	useAuthStore,
	useRequestStore,
	useWorkspaceStore,
} from "@reqly/frontend";
import { Events } from "@wailsio/runtime";
import { AppService } from "../bindings/github.com/Its-Satyajit/reqly/apps/desktop/backend/index";

/**
 * wailsSender executes requests through the Go core via the generated Wails
 * bindings, then normalizes the core response into the shared ResponseData
 * shape the UI renders.
 */
export const wailsSender: RequestSender = async (
	req: RequestInput,
): Promise<ResponseData> => {
	const headers = (req.headers ?? []).map(({ key, value }) => ({ key, value }));
	const { body, contentType } = serializeBody(req);
	const hasManualType = headers.some(
		(h) => h.key.toLowerCase() === "content-type",
	);
	if (contentType && !hasManualType)
		headers.push({ key: "Content-Type", value: contentType });

	const res = await AppService.SendRequest(
		{
			method: req.method,
			url: req.url,
			headers,
			query: (req.params ?? []).map(({ key, value }) => ({ key, value })),
			body,
			auth: req.auth,
		} as never,
		{
			env: req.env ?? "",
			requestPath: req.requestPath ?? "",
		} as never,
	);
	if (!res) {
		throw new Error("core returned an empty response");
	}
	return res as ResponseData;
};

/**
 * wailsAuthAdapter executes auth actions through the Go core's AuthService
 * via the generated Wails bindings, shaping the results into the shared
 * AuthAdapter contract the auth panel renders.
 */
export const wailsAuthAdapter: AuthAdapter = {
	login: async (config, flow) => {
		const tok = await AppService.AuthLogin(config, flow);
		if (!tok) {
			throw new Error("login returned no token");
		}
		return { accessToken: tok.AccessToken };
	},
	status: async () => {
		const status = await AppService.AuthStatus();
		if (!status) {
			throw new Error("core returned an empty status");
		}
		return {
			backend: status.backend,
			tokens: (status.tokens ?? []).map((t) => ({
				endpoint: t.endpoint,
				grantType: t.grantType,
				expiry: t.expiry,
				accessToken: t.accessToken,
				hasRefresh: t.hasRefresh,
				state: t.state,
			})),
		};
	},
	logout: async () => AppService.AuthLogout(),
};

/**
 * normalizeVariables coerces the generated bindings' nullable/undefined-valued
 * map into a plain string map the shared types expect.
 */
const normalizeVariables = (
	v: Record<string, string | undefined> | null | undefined,
) => {
	const out: Record<string, string> = {};
	for (const [k, val] of Object.entries(v ?? {})) {
		if (typeof val === "string") out[k] = val;
	}
	return out;
};

export const wailsEnvAdapter: EnvAdapter = {
	list: async () => {
		const data = await AppService.EnvList();
		if (!data) {
			throw new Error("core returned an empty environment list");
		}
		return {
			active: data.active ?? "",
			environments: (data.environments ?? []).map((e) => ({
				name: e.name,
				description: e.description ?? "",
				variables: normalizeVariables(e.variables),
				secrets: e.secrets ?? [],
			})),
		};
	},
	read: async (name) => {
		const env = await AppService.EnvRead(name);
		if (!env) {
			throw new Error(`core returned an empty environment "${name}"`);
		}
		return {
			name: env.name,
			description: env.description ?? "",
			variables: normalizeVariables(env.variables),
			secrets: env.secrets ?? [],
		};
	},
	setActive: async (name) => {
		await AppService.EnvSetActive(name);
	},
	create: async (name, description, variables) => {
		await AppService.EnvCreate(name, description, variables);
	},
	update: async (name, description, variables) => {
		await AppService.EnvUpdate(name, description, variables);
	},
	updateSecrets: async (name, values, remove) => {
		await AppService.EnvUpdateSecrets(name, values, remove);
	},
	delete: async (name) => {
		await AppService.EnvDelete(name);
	},
};

/**
 * wailsCollectionsAdapter loads the workspace's collection tree through the
 * Go core's WorkspaceService via the generated Wails bindings. The generated
 * models are nullable; normalize them to the shared tree shapes.
 */
type WailsTree = NonNullable<
	Awaited<ReturnType<typeof AppService.WorkspaceLoad>>
>;
type WailsCollection = NonNullable<WailsTree["collections"]>[number];
type WailsFolder = NonNullable<WailsCollection["folders"]>[number];
type WailsRequest = NonNullable<WailsCollection["requests"]>[number];

const normalizeFolder = (
	f: WailsFolder,
): import("@reqly/frontend").WorkspaceFolder => ({
	name: f.name,
	path: f.path,
	folders: (f.folders ?? []).map(normalizeFolder),
	requests: (f.requests ?? []).map(normalizeRequest),
});

const normalizeRequest = (
	r: WailsRequest,
): import("@reqly/frontend").WorkspaceRequest => ({
	name: r.name,
	path: r.path,
});

type WailsOpened = NonNullable<
	Awaited<ReturnType<typeof AppService.WorkspaceOpenRequest>>
>;

const normalizeAuth = (
	a: { type?: string; config?: Record<string, string | undefined> | null } | null | undefined,
): import("@reqly/frontend").RequestAuth | undefined =>
	a
		? {
				type: a.type,
				config: a.config ? normalizeVariables(a.config) : undefined,
			}
		: undefined;

const normalizeOpenedRequest = (
	o: WailsOpened,
): import("@reqly/frontend").OpenedRequest => ({
	path: o.path,
	name: o.name,
	request: {
		method: o.request?.method ?? "GET",
		url: o.request?.url ?? "",
		headers: (o.request?.headers ?? []).map(({ key, value }) => ({
			key,
			value,
		})),
		query: (o.request?.query ?? []).map(({ key, value }) => ({ key, value })),
		body: o.request?.body ?? "",
		auth: normalizeAuth(o.request?.auth),
	},
	fileRequest: {
		method: o.fileRequest?.method ?? "GET",
		url: o.fileRequest?.url ?? "",
		headers: (o.fileRequest?.headers ?? []).map(({ key, value }) => ({
			key,
			value,
		})),
		query: (o.fileRequest?.query ?? []).map(({ key, value }) => ({ key, value })),
		body: o.fileRequest?.body ?? "",
		auth: normalizeAuth(o.fileRequest?.auth),
	},
	variables: (o.variables ?? []).map(({ name, value, scope }) => ({
		name,
		value,
		scope,
	})),
	fileEnv: o.fileEnv ?? "",
	version: o.version ?? "",
});

/**
 * Collection-run events carry the core's RunStep/RunReport DTOs (serialized
 * over the Wails event channel, so they never surface in the generated
 * bindings). These normalizers coerce the loosely-typed payloads into the
 * shared run types the run view renders.
 */
const normalizeRunTestResult = (t: {
	name?: string;
	passed?: boolean;
}): RunTestResult => ({
	name: t.name ?? "",
	passed: !!t.passed,
});

const normalizeRunStep = (s: {
	name?: string;
	requestPath?: string;
	passed?: boolean;
	requestError?: string;
	response?: ResponseData | null;
	tests?: { name?: string; passed?: boolean }[];
	logs?: string[];
}): RunStep => ({
	name: s.name ?? "",
	requestPath: s.requestPath ?? "",
	passed: !!s.passed,
	requestError: s.requestError ?? "",
	response: s.response ?? null,
	tests: (s.tests ?? []).map(normalizeRunTestResult),
	logs: s.logs ?? [],
});

const normalizeRunReport = (r: {
	steps?: unknown[];
	started?: string;
	finished?: string;
	total?: number;
	passed?: number;
	failed?: number;
	durationMs?: number;
	ok?: boolean;
}): RunReport => ({
	steps: (r.steps ?? []).map((s) => normalizeRunStep(s as Parameters<typeof normalizeRunStep>[0])),
	started: r.started ?? "",
	finished: r.finished ?? "",
	total: r.total ?? 0,
	passed: r.passed ?? 0,
	failed: r.failed ?? 0,
	durationMs: r.durationMs ?? 0,
	ok: !!r.ok,
});

export const wailsCollectionsAdapter: CollectionsAdapter = {
	load: async () => {
		const tree = await AppService.WorkspaceLoad();
		if (!tree) {
			throw new Error("core returned an empty workspace tree");
		}
		return {
			name: tree.name ?? "",
			path: tree.path ?? "",
			collections: (tree.collections ?? []).map((c) => ({
				name: c.name,
				path: c.path,
				folders: (c.folders ?? []).map(normalizeFolder),
				requests: (c.requests ?? []).map(normalizeRequest),
			})),
		};
	},
	open: async (path) => {
		const opened = await AppService.WorkspaceOpenRequest(path);
		if (!opened) {
			throw new Error("core returned an empty opened request");
		}
		return normalizeOpenedRequest(opened);
	},
	save: async (path, draft, expectedVersion) => {
		const version = await AppService.WorkspaceSaveRequest(
			path,
			{
				method: draft.method,
				url: draft.url,
				headers: (draft.headers ?? []).map(({ key, value }) => ({
					key,
					value,
				})),
				query: (draft.query ?? []).map(({ key, value }) => ({ key, value })),
				body: draft.body,
				auth: draft.auth,
			} as never,
			expectedVersion,
		);
		if (!version) {
			throw new Error("core returned an empty version after save");
		}
		return version;
	},
	run: async (path, env, failFast, onEvent) => {
		const id = await AppService.WorkspaceRunCollection(path, env ?? "", failFast);
		if (!id) {
			throw new Error("core returned an empty run id");
		}
		// Subscribe before returning so no streamed event is missed: the run
		// executes on the core's goroutine and the first step can arrive while
		// the binding's response is still in flight.
		const offStep = Events.On(`reqly.run.${id}.step`, (e) => {
			onEvent({ type: "step", step: normalizeRunStep(e.data) });
		});
		const offDone = Events.On(`reqly.run.${id}.done`, (e) => {
			onEvent({ type: "done", report: normalizeRunReport(e.data) });
			offStep();
			offDone();
		});
		const offError = Events.On(`reqly.run.${id}.error`, (e) => {
			onEvent({ type: "error", message: String(e.data ?? "") });
			offStep();
			offDone();
			offError();
		});
		return id;
	},
	cancelRun: async (id) => {
		await AppService.WorkspaceRunCancel(id);
	},
};

/**
 * Wires the Go core behind the shared request, auth, and environment stores.
 * Called once from the host entry point, before the React tree mounts.
 */
export function initRequestBridge(): void {
	useRequestStore.getState().setSender(wailsSender);
	useAuthStore.getState().setAdapter(wailsAuthAdapter);
	useWorkspaceStore.getState().setEnvAdapter(wailsEnvAdapter);
	useWorkspaceStore.getState().setWorkspaceAdapter(wailsCollectionsAdapter);
}
