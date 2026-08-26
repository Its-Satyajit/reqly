import { useEffect, useState } from "react";
import { useForm } from "@tanstack/react-form";
import { Check, Copy, Plus, Trash2 } from "lucide-react";
import { Button } from "../../components/ui/button";
import { cn } from "../../lib/utils";
import { notifySuccess } from "../../lib/notify";
import { inputClass } from "../../lib/ui";
import { DYNAMIC_TAGS } from "../../lib/tags";
import type { EnvAdapter } from "../../lib/env";
import { useWorkspaceStore } from "../../stores/useWorkspaceStore";
import { EnvironmentEditor } from "./EnvironmentEditor";
import { SecretsEditor } from "./SecretsEditor";
import { Tabs, TabsList, TabsTrigger } from "#components/ui/tabs";

const DOT_COLORS = ["bg-status-ok", "bg-status-warn", "bg-status-error", "bg-status-info"];

/** copyTag copies a {{$tag}} chip to the clipboard. Module-level because
 * React Compiler cannot handle try/catch. */
async function copyTag(tag: string): Promise<boolean> {
	try {
		await navigator.clipboard.writeText(`{{$${tag}}}`);
		return true;
	} catch {
		return false;
	}
}

/** deleteEnvironment removes an environment via the adapter. Module-level
 * (outside the compiled component) because React Compiler cannot handle
 * try/catch. */
async function deleteEnvironment(
	envAdapter: EnvAdapter,
	refresh: () => Promise<void>,
	name: string,
): Promise<string | null> {
	try {
		await envAdapter.delete(name);
		await refresh();
		return null;
	} catch (err) {
		return err instanceof Error ? err.message : String(err);
	}
}

type EnvTab = "variables" | "secrets";

interface CreateEnvForm {
	name: string;
	description: string;
}

type CreateEnvFieldsProps = {
	onCreated: (name: string) => void;
	onError: (message: string) => void;
};

/** CreateEnvFields is the self-contained inline new-environment form. */
function CreateEnvFields({ onCreated, onError }: CreateEnvFieldsProps) {
	const envAdapter = useWorkspaceStore((s) => s.envAdapter);
	const form = useForm({
		defaultValues: { name: "", description: "" } satisfies CreateEnvForm,
		onSubmit: async ({ value }) => {
			const name = value.name.trim();
			if (name === "") return;
			try {
				await envAdapter.create(name, value.description.trim(), {});
				notifySuccess(`Environment "${name}" created`);
				form.reset();
				onCreated(name);
			} catch (err) {
				onError(err instanceof Error ? err.message : String(err));
			}
		},
	});
	return (
		<form
			onSubmit={(e) => {
				e.preventDefault();
				void form.handleSubmit();
			}}
			className="flex flex-col gap-1 pb-1"
		>
			<form.Field
				name="name"
				validators={{
					onChange: ({ value }) =>
						value.trim() ? undefined : 'Name the environment — e.g. "staging".',
				}}
			>
				{(field) => (
					<span className="flex flex-col gap-1">
						<input
							name={field.name}
							value={field.state.value}
							onBlur={field.handleBlur}
							onChange={(e) => field.handleChange(e.target.value)}
							placeholder="name (e.g. staging)"
							aria-label="Environment name"
							spellCheck={false}
							className={`${inputClass} font-mono text-xs`}
						/>
						{field.state.meta.errors.length > 0 ? (
							<span className="text-[11px] text-status-error">
								{field.state.meta.errors[0]}
							</span>
						) : null}
					</span>
				)}
			</form.Field>
			<form.Field name="description">
				{(field) => (
					<input
						name={field.name}
						value={field.state.value}
						onBlur={field.handleBlur}
						onChange={(e) => field.handleChange(e.target.value)}
						placeholder="description (optional)"
						aria-label="Environment description"
						className={`${inputClass} text-xs`}
					/>
				)}
			</form.Field>
			<form.Subscribe
				selector={(state) => ({
					canSubmit: state.canSubmit,
					isSubmitting: state.isSubmitting,
				})}
			>
				{(sub: { canSubmit: boolean; isSubmitting: boolean }) => (
					<Button size="xs" type="submit" disabled={!sub.canSubmit || sub.isSubmitting}>
						{sub.isSubmitting ? "Creating…" : "Create"}
					</Button>
				)}
			</form.Subscribe>
		</form>
	);
}

const SHARED_SCOPES = [
	{ name: "Global", detail: "Every workspace, every request" },
	{ name: "Workspace", detail: "This repository only" },
	{ name: "Collection", detail: "Inherited by folders & requests" },
] as const;

/** EnvironmentsView is the G-17.4.6 environments surface: an env-list side
 * panel (vars/secrets counts, active pill) with shared-scope cards beside a
 * per-environment editor — Active toggle, Duplicate, Delete, Variables/
 * Secrets tabs, and click-to-copy dynamic tags. */
export function EnvironmentsView() {
	const environments = useWorkspaceStore((s) => s.environments);
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);
	const environmentsError = useWorkspaceStore((s) => s.environmentsError);
	const envAdapter = useWorkspaceStore((s) => s.envAdapter);
	const refreshEnvironments = useWorkspaceStore((s) => s.refreshEnvironments);

	const [selectedName, setSelectedName] = useState<string | null>(null);
	const [tab, setTab] = useState<EnvTab>("variables");
	const [creating, setCreating] = useState(false);
	const [actionError, setActionError] = useState<string | null>(null);
	const [copiedTag, setCopiedTag] = useState<string | null>(null);

	useEffect(() => {
		void refreshEnvironments();
	}, [refreshEnvironments]);

	const selected =
		environments.find((e) => e.name === selectedName) ??
		environments.find((e) => e.id === activeEnvironmentId) ??
		environments[0] ??
		null;
	const isActive = selected?.id === activeEnvironmentId;

	const duplicate = async (): Promise<void> => {
		if (!selected) return;
		setActionError(null);
		const copyName = `${selected.name}-copy`;
		try {
			await envAdapter.create(copyName, selected.description, selected.variables);
			notifySuccess(`Duplicated to "${copyName}"`);
			await refreshEnvironments();
			setSelectedName(copyName);
		} catch (err) {
			setActionError(err instanceof Error ? err.message : String(err));
		}
	};

	const activate = async (): Promise<void> => {
		if (!selected) return;
		setActionError(null);
		try {
			await envAdapter.setActive(selected.name);
			await refreshEnvironments();
		} catch (err) {
			setActionError(err instanceof Error ? err.message : String(err));
		}
	};

	const remove = async (): Promise<void> => {
		if (!selected) return;
		setActionError(null);
		const deleteError = await deleteEnvironment(envAdapter, refreshEnvironments, selected.name);
		if (deleteError !== null) {
			setActionError(deleteError);
		} else {
			setSelectedName(null);
		}
	};

	const copyTagChip = (tag: string): void => {
		void copyTag(tag).then((ok) => {
			if (!ok) return;
			setCopiedTag(tag);
			setTimeout(() => setCopiedTag(null), 1500);
		});
	};

	return (
		<div className="flex h-full min-h-0 gap-4 p-4" aria-label="Environments">
			<aside className="flex w-64 shrink-0 flex-col gap-3 overflow-y-auto">
				<div className="flex flex-col gap-1 rounded-xl border border-border bg-card p-3">
					<div className="flex items-center justify-between">
						<p className="font-data text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
							Environments
						</p>
						<Button
							size="icon"
							variant="ghost"
							aria-label="New environment"
							onClick={() => setCreating(!creating)}
						>
							<Plus className="size-3.5" />
						</Button>
					</div>
					{creating ? (
						<CreateEnvFields
							onCreated={(name) => {
								setCreating(false);
								setSelectedName(name);
							}}
							onError={setActionError}
						/>
					) : null}
					{environments.map((env, i) => {
						const envActive = env.id === activeEnvironmentId;
						return (
							<button
								key={env.id}
								type="button"
								onClick={() => setSelectedName(env.name)}
								className={cn(
									"flex flex-col gap-0.5 rounded-lg border px-2.5 py-2 text-left transition-colors",
									selected?.name === env.name
										? "border-primary/50 bg-primary/5"
										: "border-transparent hover:bg-accent",
								)}
							>
								<span className="flex items-center gap-1.5">
									<span
										aria-hidden
										className={cn("size-2 shrink-0 rounded-full", DOT_COLORS[i % DOT_COLORS.length])}
									/>
									<span className="truncate text-xs font-semibold text-foreground">{env.name}</span>
									{envActive ? (
										<span className="ml-auto shrink-0 rounded-full border border-status-ok/40 px-1.5 font-data text-[9px] text-status-ok">
											active
										</span>
									) : null}
								</span>
								<span className="pl-3.5 text-[11px] text-muted-foreground">
									{Object.keys(env.variables).length} vars · {env.secrets.length} secrets
								</span>
							</button>
						);
					})}
					{environments.length === 0 ? (
						<p className="pb-1 text-[11px] text-muted-foreground">
							No environments yet — hit + to create one.
						</p>
					) : null}
				</div>

				<div className="flex flex-col gap-1 rounded-xl border border-border bg-card p-3">
					<p className="font-data text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
						Shared scopes
					</p>
					{SHARED_SCOPES.map((s) => (
						<div key={s.name} className="rounded-lg border border-border bg-muted/20 px-2.5 py-1.5">
							<p className="text-xs font-semibold text-foreground">{s.name}</p>
							<p className="text-[11px] text-muted-foreground">{s.detail}</p>
						</div>
					))}
				</div>
			</aside>

			<section className="flex min-w-0 flex-1 flex-col gap-3 overflow-y-auto">
				{selected ? (
					<>
						<div className="flex flex-wrap items-center gap-2">
							<h2 className="text-lg font-semibold text-foreground">{selected.name}</h2>
							{isActive ? (
								<span className="rounded-full border border-status-ok/40 px-2 py-0.5 font-data text-[10px] text-status-ok">
									ok
								</span>
							) : null}
							<Button
								size="sm"
								variant={isActive ? "outline" : "destructive"}
								disabled={isActive}
								onClick={() => void activate()}
							>
								{isActive ? (
									<>
										<Check data-icon="inline-start" />
										Active
									</>
								) : (
									"Set active"
								)}
							</Button>
							<div className="ml-auto flex items-center gap-2">
								<Button variant="outline" size="sm" onClick={() => void duplicate()}>
									<Copy data-icon="inline-start" />
									Duplicate
								</Button>
								<Button variant="outline" size="sm" onClick={() => void remove()}>
									<Trash2 data-icon="inline-start" />
									Delete
								</Button>
							</div>
						</div>

						{actionError ? <p className="text-xs text-status-error">{actionError}</p> : null}
						{environmentsError ? (
							<p className="text-xs text-status-error">{environmentsError}</p>
						) : null}

						<Tabs
							value={tab}
							onValueChange={(v) => {
								// SAFETY: tab ids come from the local tab list above
								setTab(v as EnvTab)
							}}
						>
							<TabsList variant="line" aria-label="Environment tabs">
								<TabsTrigger
									value="variables"
									className={cn(
										"rounded-full px-3 py-1 text-xs",
										tab === "variables"
											? "bg-primary/15 font-medium text-primary"
											: "text-muted-foreground hover:text-foreground",
									)}
								>
									{`Variables (${Object.keys(selected.variables).length})`}
								</TabsTrigger>
								<TabsTrigger
									value="secrets"
									className={cn(
										"rounded-full px-3 py-1 text-xs",
										tab === "secrets"
											? "bg-primary/15 font-medium text-primary"
											: "text-muted-foreground hover:text-foreground",
									)}
								>
									{`Secrets (${selected.secrets.length})`}
								</TabsTrigger>
							</TabsList>
						</Tabs>

						{tab === "variables" ? (
							<EnvironmentEditor key={`env-${selected.name}`} env={selected} onCancel={() => undefined} />
						) : (
							<SecretsEditor
								key={`secrets-${selected.name}`}
								envName={selected.name}
								secretNames={selected.secrets}
								variableNames={Object.keys(selected.variables)}
							/>
						)}

						<div className="flex flex-col gap-1.5">
							<p className="font-data text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
								Dynamic tags
							</p>
							<div className="flex flex-wrap gap-1.5">
								{DYNAMIC_TAGS.map((tag) => (
									<button
										key={tag}
										type="button"
										onClick={() => copyTagChip(tag)}
										title="Click to copy"
										className="rounded-full border border-border bg-muted/30 px-2 py-0.5 font-data text-[11px] text-muted-foreground transition-colors hover:text-foreground"
									>
										{copiedTag === tag ? "copied!" : `{{$${tag}}}`}
									</button>
								))}
							</div>
							<p className="text-[11px] text-muted-foreground/70">click to copy</p>
						</div>
					</>
				) : (
					<div className="flex flex-1 items-center justify-center rounded-xl border border-border bg-card">
						<p className="text-xs text-muted-foreground">
							Create or pick an environment to edit its variables and secrets.
						</p>
					</div>
				)}
			</section>
		</div>
	);
}
