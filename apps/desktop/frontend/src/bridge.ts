import type {
	AuthAdapter,
	CollectionsAdapter,
	EnvAdapter,
	HistoryAdapter,
	ExportAdapter,
	ExportFormat,
	DiffAdapter,
	EnvToolsAdapter,
	GqlAdapter,
	OpenapiAdapter,
	RunnerAdapter,
	JwtAdapter,
	JwtTokenView,
	MockAdapter,
	MockStatus,
	RealtimeAdapter,
	RealtimeFrameView,
	ImportAdapter,
	ImportOutcome,
	WorkspaceBootstrapAdapter,
	RequestInput,
	RequestSender,
	ResponseData,
	RunReport,
	RunStep,
	RunTestResult,
} from "@reqly/frontend";
import {
	normalizeHeaderKeys,
	serializeBody,
	useAuthStore,
	useExportStore,
	useHistoryStore,
	useImportStore,
	useMockStore,
	useRealtimeStore,
	setDiffBridge,
	setEnvToolsBridge,
	setGqlBridge,
	setOpenapiBridge,
	setRunnerBridge,
	setJwtBridge,
	useRequestStore,
	useWorkspaceBootstrapStore,
	useWorkspaceStore,
	addGoLog,
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
	const headers = (req.headers ?? []).map(({ key, value }: { key: string; value: string }) => ({ key, value }));
	const { body, contentType } = serializeBody(req);
	const hasManualType = headers.some(
		(h) => h.key.toLowerCase() === "content-type",
	);
	if (contentType && !hasManualType)
		headers.push({ key: "Content-Type", value: contentType });

	// SAFETY: Wails binding DTO shapes verified via Go core SendRequest; ResponseData mapping is boundary-parsed
	let res;
	try {
		// SAFETY: Wails binding DTO shapes verified via Go core SendRequest; ResponseData mapping is boundary-parsed
		res = await AppService.SendRequest(
			{
				method: req.method,
				url: req.url,
				headers,
				query: (req.params ?? []).map(({ key, value }: { key: string; value: string }) => ({ key, value })),
				body,
				auth: req.auth,
				retry: req.retry ?? null,
			} as never,
			{
				env: req.env ?? "",
				requestPath: req.requestPath ?? "",
				sendId: req.sendId ?? "",
			} as never,
		);
	} catch (err) {
		// A cancelled send surfaces as a Go "context canceled" rejection; map it
		// to the neutral message the UI renders for Stop.
		const message = err instanceof Error ? err.message : String(err);
		if (message.includes("context canceled")) {
			throw new Error("Request cancelled");
		}
		throw err;
	}
	if (!res) {
		throw new Error("core returned an empty response");
	}
	// SAFETY: AppService response shape is ResponseData DTO from Go core; header
	// keys arrive in Go canonical case and are lowercased once at this boundary
	return {
		...(res as ResponseData),
		headers: normalizeHeaderKeys((res as ResponseData).headers),
	};
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
			tokens: (status.tokens ?? []).map((t: NonNullable<typeof status.tokens>[number]) => ({
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
		if (val !== undefined) out[k] = val;
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
			environments: (data.environments ?? []).map((e: { name: string; description?: string; variables?: Record<string, string | undefined> | null; secrets?: string[] | null }) => ({
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
		headers: (o.request?.headers ?? []).map(({ key, value }: { key: string; value: string }) => ({
			key,
			value,
		})),
		query: (o.request?.query ?? []).map(({ key, value }: { key: string; value: string }) => ({ key, value })),
		body: o.request?.body ?? "",
		auth: normalizeAuth(o.request?.auth),
	},
	fileRequest: {
		method: o.fileRequest?.method ?? "GET",
		url: o.fileRequest?.url ?? "",
		headers: (o.fileRequest?.headers ?? []).map(({ key, value }: { key: string; value: string }) => ({
			key,
			value,
		})),
		query: (o.fileRequest?.query ?? []).map(({ key, value }: { key: string; value: string }) => ({ key, value })),
		body: o.fileRequest?.body ?? "",
		auth: normalizeAuth(o.fileRequest?.auth),
	},
	variables: (o.variables ?? []).map(({ name, value, scope }: { name: string; value: string; scope: string }) => ({
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
	// SAFETY: run report steps are boundary-parsed from Wails event DTO; shape validated via normalizeRunStep
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
			collections: (tree.collections ?? []).map((c: WailsCollection) => ({
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
		// SAFETY: Wails WorkspaceSaveRequest DTO shape verified via Go Request struct
		const version = await AppService.WorkspaceSaveRequest(
			path,
			{
				method: draft.method,
				url: draft.url,
				headers: (draft.headers ?? []).map(({ key, value }: { key: string; value: string }) => ({
					key,
					value,
				})),
				query: (draft.query ?? []).map(({ key, value }: { key: string; value: string }) => ({ key, value })),
				body: draft.body,
				auth: draft.auth,
				retry: draft.retry ?? null,
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
		const offStep = Events.On(`reqly.run.${id}.step`, (e: { data: unknown }) => {
			// SAFETY: Wails event data is RunStep DTO from Go core; validated via normalizeRunStep
			onEvent({ type: "step", step: normalizeRunStep(e.data as Parameters<typeof normalizeRunStep>[0]) });
		});
		const offDone = Events.On(`reqly.run.${id}.done`, (e: { data: unknown }) => {
			// SAFETY: Wails event data is RunReport DTO from Go core; validated via normalizeRunReport
			onEvent({ type: "done", report: normalizeRunReport(e.data as Parameters<typeof normalizeRunReport>[0]) });
			offStep();
			offDone();
		});
		const offError = Events.On(`reqly.run.${id}.error`, (e: { data: unknown }) => {
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

const toHistoryEntry = (e: {
	ID: string;
	RequestPath: string;
	Method: string;
	URL: string;
	Env: string;
	Status: number;
	DurationMS: number;
	Size: number;
	CreatedAt: string;
}) => ({
	id: e.ID,
	requestPath: e.RequestPath,
	method: e.Method,
	url: e.URL,
	env: e.Env,
	status: e.Status,
	durationMs: e.DurationMS,
	size: e.Size,
	createdAt: e.CreatedAt,
});

export const wailsHistoryAdapter: HistoryAdapter = {
	list: async (limit, offset, status, env) => {
		const data = await AppService.HistoryList(limit, offset, status, env);
		return (data ?? []).map(toHistoryEntry);
	},
	show: async (id) => {
		const e = await AppService.HistoryShow(id);
		if (!e) throw new Error(`history entry ${id} not found`);
		return {
			...toHistoryEntry(e),
			reqHeaders: normalizeHeaderKeys(e.ReqHeaders ?? undefined),
			reqBody: e.ReqBody ?? "",
			respHeaders: normalizeHeaderKeys(e.RespHeaders ?? undefined),
			respBody: e.RespBody ?? "",
		};
	},
	search: async (query, limit) => {
		const data = await AppService.HistorySearch(query, limit);
		return (data ?? []).map(toHistoryEntry);
	},
	clear: async (env) => {
		await AppService.HistoryClear(env);
	},
	replay: async (id) => {
		const res = await AppService.HistoryReplay(id);
		if (!res) return null;
		// SAFETY: SendResponse DTO from Go core; headers lowercased at boundary
		return {
			statusCode: res.statusCode,
			durationMs: res.durationMs,
			size: res.size,
			headers: normalizeHeaderKeys(res.headers as ResponseData["headers"]),
			body: res.body,
		};
	},
	listCookies: async (env) => {
		const data = await AppService.CookieList(env);
		return (data ?? []).map((c) => ({
			name: c.Name,
			value: c.Value,
			domain: c.Domain ?? "",
			path: c.Path ?? "/",
			env: c.Env ?? "",
		}));
	},
	deleteCookie: async (name, domain, path, env) => {
		await AppService.CookieDelete(name, domain, path, env);
	},
	clearCookies: async (env) => {
		await AppService.CookieClear(env);
	},
};

type WailsImportResult = NonNullable<Awaited<ReturnType<typeof AppService.Import>>>;

function toImportOutcome(res: WailsImportResult): ImportOutcome {
	return {
		kind: res.kind,
		format: res.format,
		title: res.title,
		requestCount: res.requestCount,
		environmentCount: res.environmentCount,
		targetDir: res.targetDir,
		report: res.report ?? undefined,
		operations: res.operations ?? undefined,
	};
}

async function runImport(
	req: Parameters<typeof AppService.Import>[0],
): Promise<ImportOutcome> {
	const res = await AppService.Import(req);
	if (!res) throw new Error("import failed");
	return toImportOutcome(res);
}

export const wailsWorkspaceBootstrapAdapter: WorkspaceBootstrapAdapter = {
	status: async () => (await AppService.WorkspaceStatus()) ?? { found: false },
	restoreLast: async () => (await AppService.WorkspaceRestoreLast()) ?? { found: false },
	pickFolder: async () => {
		const dir = await AppService.WorkspacePickFolder();
		return dir ?? "";
	},
	open: async (dir) => {
		await AppService.WorkspaceOpen(dir);
	},
	create: async (dir, name) => {
		await AppService.WorkspaceCreate(dir, name ?? "");
	},
};

export const wailsImportAdapter: ImportAdapter = {
	detect: async (content) => {
		const [format, ok] = await AppService.Detect(content);
		return { format, ok };
	},
	preview: ({ content, formatHint }) =>
		runImport({ content, formatHint, dryRun: true }),
	commit: ({ content, formatHint, targetDir }) =>
		runImport({ content, formatHint, targetDir, dryRun: false }),
};

type WailsDiffResult = NonNullable<Awaited<ReturnType<typeof AppService.DiffSpecs>>>["result"];

// toDiffResultView maps the generated Change shape onto the shared view type;
// both carry identical JSON fields (type/path/from/to/severity).
function toDiffResultView(r: WailsDiffResult): import("@reqly/frontend").DiffResultView {
	const changes: import("@reqly/frontend").DiffChange[] = (r?.changes ?? []).flatMap((c) =>
		c == null
			? []
			: [
					{
						// SAFETY: the backend only emits create/update/delete.
						type: c.type as "create" | "update" | "delete",
						path: [...(c.path ?? [])],
						from: c.from,
						to: c.to,
						severity: c.severity,
					},
				],
	);
	return { hasChanges: r?.hasChanges ?? false, changes };
}

export const wailsGqlAdapter: GqlAdapter = {
	introspect: async ({ endpoint, headers, timeoutSec }) => {
		const res = await AppService.GraphqlIntrospect(endpoint, headers ?? [], timeoutSec ?? 0);
		if (!res) throw new Error("introspection failed");
		// mapRef converts the generated wrapped-type reference recursively;
		// both shapes carry name/kind/of with identical meanings.
		type RawRef = { name?: string; kind?: string; of?: RawRef | null };
		const mapRef = (ref: RawRef | null | undefined): import("@reqly/frontend").GqlTypeRef | null =>
			ref == null
				? null
				: { name: ref.name, kind: ref.kind, of: mapRef(ref.of) };
		const toGqlType = (t: typeof res.query): import("@reqly/frontend").GqlType | null => {
			if (t == null) return null;
			const gqlType: import("@reqly/frontend").GqlType = {
				kind: t.kind,
				name: t.name,
				fields: (t.fields ?? []).map((f) => ({
					name: f.name,
					// SAFETY: generated refs carry only name/kind/of.
					type: mapRef(f.type as RawRef | null),
					args: (f.args ?? []).map((a) => ({
						name: a.name,
						// SAFETY: same generated ref shape as field types.
						type: mapRef(a.type as RawRef | null),
					})),
					deprecated: f.deprecated ?? false,
				})),
			};
			if (t.description) gqlType.description = t.description;
			if (t.enumValues) gqlType.enumValues = [...t.enumValues];
			return gqlType;
		};
		return {
			query: toGqlType(res.query),
			mutation: toGqlType(res.mutation),
			subscription: toGqlType(res.subscription),
			types: (res.types ?? []).map(toGqlType).filter(
				(t): t is NonNullable<typeof t> => t != null,
			),
		};
	},
};

// RunnerStepPayload mirrors backend runnerStep JSON.
type RunnerStepPayload = {
	index: number;
	status?: number;
	error?: string;
	url?: string;
	bodyPreview?: string;
};

export const wailsRunnerAdapter: RunnerAdapter = {
	start: async (input) => {
		await AppService.RunnerStart({
			runId: input.runId,
			kind: input.kind,
			request: input.request,

			pagination: input.pagination,
			maxPagesOverride: input.maxPagesOverride,
			data: input.data,
			dataFormat: input.dataFormat,
			parallel: input.parallel,
			concurrency: input.concurrency,
		});
	},
	cancel: async (runId) => {
		AppService.RunnerCancel(runId);
	},
	listen: (runId, handlers) => {
		const offs = [
			// SAFETY: frames mirror the Go runnerStep struct field-for-field.
			Events.On(`reqly.runner.${runId}.step`, (e: { data?: RunnerStepPayload }) => {
				if (e?.data)
					handlers.onStep({
						index: e.data.index,
						status: e.data.status,
						error: e.data.error,
						url: e.data.url,
						bodyPreview: e.data.bodyPreview,
					});
			}),
			// SAFETY: done summaries are plain string-keyed maps.
			Events.On(`reqly.runner.${runId}.done`, (e: { data?: RunnerStepPayload }) => {
				if (e?.data) {
					// SAFETY: Go only emits scalar summary fields.
					handlers.onDone({ ...e.data } as import("@reqly/frontend").RunnerSummary);
				}
			}),
		];
		return () => offs.forEach((off) => off());
	},
};

export const wailsOpenapiAdapter: OpenapiAdapter = {
	explore: async (specPath) => {
		const res = await AppService.OpenapiExplore(specPath);
		if (!res) throw new Error("explore failed");
		const endpoints = (res.endpoints ?? []).map((ep) => {
			const view: import("@reqly/frontend").OpenapiEndpointView = { method: ep.method, path: ep.path };
			if (ep.operationId) view.operationId = ep.operationId;
			if (ep.tags) view.tags = [...ep.tags];
			if (ep.summary) view.summary = ep.summary;
			if (ep.requestSchema) view.requestSchema = ep.requestSchema;
								// SAFETY: responseSchemas values are JSON strings built in Go.
					if (ep.responseSchemas) {
						const schemas: Record<string, string> = {};
						for (const [status, schema] of Object.entries(ep.responseSchemas)) {
							if (schema != null) schemas[status] = schema;
						}
						view.responseSchemas = schemas;
					}
			return view;
		});
		const out: import("@reqly/frontend").OpenapiExploreResultView = {
			title: res.title,
			endpoints,
		};
		if (res.version) out.version = res.version;
		return out;
	},
	generate: async (input) => {
		const res = await AppService.OpenapiGenerateRequests(
			input.specPath,
			input.selections,
			input.dirName,
		);
		if (!res) throw new Error("generate failed");
		const out: import("@reqly/frontend").OpenapiGenerateResultView = {
			targetDir: res.targetDir,
			created: [...(res.created ?? [])],
		};
		if (res.warnings) out.warnings = [...res.warnings];
		return out;
	},
};

export const wailsEnvToolsAdapter: EnvToolsAdapter = {
	diff: async (envA, envB) => {
		const res = await AppService.EnvDiff(envA, envB);
		if (!res) throw new Error("diff failed");
		return { envA: res.envA, envB: res.envB, diffs: res.diffs ?? [] };
	},
	validate: async (name) => {
		const res = await AppService.EnvValidate(name);
		if (!res) throw new Error("validate failed");
		return { env: res.env, issues: res.issues ?? [] };
	},
	crossValidate: async () => {
		const gaps = await AppService.EnvCrossValidate();
		if (!gaps) throw new Error("cross-validation failed");
		return gaps.map((g) => ({
			key: g.key,
			presentIn: [...(g.presentIn ?? [])],
			missingIn: [...(g.missingIn ?? [])],
		}));
	},
};

export const wailsJwtAdapter: JwtAdapter = {
	decode: async (token) => {
		const res = await AppService.JwtDecode(token);
		if (!res) throw new Error("decode failed");
		// SAFETY: the backend emits exactly these four expiry statuses.
		const status = res.expiry?.status as JwtTokenView["expiry"]["status"];
		const e = res.expiry;
		return {
			header: res.header ?? [],
			payload: res.payload ?? [],
			signature: res.signature,
			alg: res.alg,
			expiry: {
				status,
				remaining: e?.remaining ?? 0,
				exp: e?.exp ?? undefined,
				nbf: e?.nbf ?? undefined,
				iat: e?.iat ?? undefined,
			},
		};
	},
};

export const wailsDiffAdapter: DiffAdapter = {
	specs: async (pathA, pathB) => {
		const res = await AppService.DiffSpecs(pathA, pathB);
		if (!res || !res.result) throw new Error("diff failed");
		return { result: toDiffResultView(res.result), breaking: res.breaking, addition: res.addition };
	},
	responses: async (idA, idB) => {
		const res = await AppService.DiffResponses(idA, idB);
		if (!res) throw new Error("diff failed");
		return {
			metaA: res.metaA ?? null,
			metaB: res.metaB ?? null,
			result: toDiffResultView(res.result ?? { hasChanges: false }),
		};
	},
};

export const wailsMockAdapter: MockAdapter = {
	start: async ({ specPath, port, delayMs, failEvery, routes }) => {
		const res: MockStatus | null = await AppService.MockStart({
			specPath,
			port,
			delayMs,
			failEvery,
			routes,
		});
		if (!res) throw new Error("mock start failed");
		return res;
	},
	stop: async () => {
		await AppService.MockStop();
	},
	status: async () => {
		const res = await AppService.MockStatusSnapshot();
		if (!res) return { running: false };
		return {
			running: res.running,
			url: res.url,
			port: res.port,
			error: res.error,
		};
	},
};

export const wailsRealtimeAdapter: RealtimeAdapter = {
	open: async ({ sessionId, kind, url, headers }) => {
		await AppService.RealtimeOpen({
			sessionId,
			kind,
			url,
			headers,
		});
	},
	send: async (sessionId, data) => {
		await AppService.RealtimeSend(sessionId, data);
	},
	sendBinary: async (sessionId, base64) => {
		await AppService.RealtimeSendBinary(sessionId, base64);
	},
	close: async (sessionId) => {
		await AppService.RealtimeClose(sessionId);
	},
	subscribe: (sessionId, onFrame) => {
		const off = Events.On(`reqly.realtime.${sessionId}`, (e: { data?: RealtimeFrameView }) => {
			if (e?.data) onFrame(e.data);
		});
		return () => off();
	},
};

type WailsExportResult = NonNullable<Awaited<ReturnType<typeof AppService.Export>>>;

export const wailsExportAdapter: ExportAdapter = {
	run: async (input: {
		format: ExportFormat;
		collection?: string;
		outName?: string;
	}) => {
		const { format, collection, outName } = input;
		const res: WailsExportResult | null = await AppService.Export({
			format,
			collection,
			outName,
		});
		if (!res) throw new Error("export failed");
		// SAFETY: the backend only emits the four ExportRequest format values.
		return {
			format: res.format as ExportFormat,
			path: res.path,
			requestCount: res.requestCount,
			entryCount: res.entryCount,
		};
	},
};

/**
 * Wires the Go core behind the shared request, auth, and environment stores.
 * Called once from the host entry point, before the React tree mounts.
 */
export function initRequestBridge(): void {
	useRequestStore.getState().setSender(wailsSender);
	useRequestStore.getState().setCancelSender(async (sendId) => {
		await AppService.CancelSend(sendId);
	});
	useAuthStore.getState().setAdapter(wailsAuthAdapter);
	useWorkspaceStore.getState().setEnvAdapter(wailsEnvAdapter);
	useWorkspaceStore.getState().setWorkspaceAdapter(wailsCollectionsAdapter);
	useHistoryStore.getState().setAdapter(wailsHistoryAdapter);
	useImportStore.getState().setAdapter(wailsImportAdapter);
	useExportStore.getState().setAdapter(wailsExportAdapter);
	useRealtimeStore.getState().setAdapter(wailsRealtimeAdapter);
	useMockStore.getState().setAdapter(wailsMockAdapter);
	setDiffBridge(wailsDiffAdapter);
	setJwtBridge(wailsJwtAdapter);
	setGqlBridge(wailsGqlAdapter);
	setRunnerBridge(wailsRunnerAdapter);
	setOpenapiBridge(wailsOpenapiAdapter);
	setEnvToolsBridge(wailsEnvToolsAdapter);
	useWorkspaceBootstrapStore.getState().setAdapter(wailsWorkspaceBootstrapAdapter);

	Events.On("reqly.golog", (e: { data?: { level?: string; message?: string } }) => {
		const payload = e.data ?? {};
		addGoLog(String(payload.level ?? "LOG"), String(payload.message ?? ""));
	});
}
