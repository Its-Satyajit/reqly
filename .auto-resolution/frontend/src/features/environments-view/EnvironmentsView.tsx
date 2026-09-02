import { EnvToolsPanel } from "../env-tools/EnvToolsPanel";
import { useEffect, useState } from "react";
import { useForm } from "@tanstack/react-form";
import { Button } from "../../components/ui/button";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "../../components/ui/alert-dialog";
import { useWorkspaceStore } from "../../stores/useWorkspaceStore";
import { cn } from "#lib/utils";
import { notifySuccess } from "../../lib/notify";
import { inputClass } from "../../lib/ui";
import { EnvironmentEditor } from "./EnvironmentEditor";
import { SecretsEditor } from "./SecretsEditor";

interface CreateEnvForm {
	name: string;
	description: string;
}

/**
 * EnvironmentsView manages the workspace's environments: list, create, set
 * the active one, and edit existing environments (description + variables)
 * through an in-memory editor with explicit Save.
 */
export function EnvironmentsView() {
	const environments = useWorkspaceStore((s) => s.environments);
	const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);
	const environmentsError = useWorkspaceStore((s) => s.environmentsError);
	const envAdapter = useWorkspaceStore((s) => s.envAdapter);
	const refreshEnvironments = useWorkspaceStore((s) => s.refreshEnvironments);

	const [createError, setCreateError] = useState<string | null>(null);
	const [setActiveError, setSetActiveError] = useState<string | null>(null);
	const [editingName, setEditingName] = useState<string | null>(null);
	const [deletingName, setDeletingName] = useState<string | null>(null);

	const createForm = useForm({
		defaultValues: { name: "", description: "" } satisfies CreateEnvForm,
		onSubmit: async ({ value }) => {
			setCreateError(null);
			try {
				await envAdapter.create(value.name.trim(), value.description.trim(), {});
				notifySuccess(`Environment "${value.name.trim()}" created`);
				createForm.reset();
				await refreshEnvironments();
			} catch (err) {
				setCreateError(err instanceof Error ? err.message : String(err));
			}
		},
	});

	useEffect(() => {
		void refreshEnvironments();
	}, [refreshEnvironments]);

	const editing = environments.find((e) => e.name === editingName) ?? null;

	const onEdit = (name: string) => {
		if (editingName === name) {
			setEditingName(null);
			return;
		}
		setEditingName(name);
	};

	const onSetActive = async (name: string) => {
		setSetActiveError(null);
		try {
			await envAdapter.setActive(name);
			await refreshEnvironments();
		} catch (err) {
			setSetActiveError(err instanceof Error ? err.message : String(err));
		}
	};

	const onDelete = async (name: string) => {
		setSetActiveError(null);
		try {
			await envAdapter.delete(name);
			if (editingName === name) setEditingName(null);
			await refreshEnvironments();
			setDeletingName(null);
		} catch (err) {
			setSetActiveError(err instanceof Error ? err.message : String(err));
			setDeletingName(null);
		}
	};

	return (
		<div className="mx-auto flex w-full max-w-3xl flex-col gap-5 p-6">
			<div>
				<h2 className="text-sm font-semibold">Environments</h2>
				<p className="text-xs text-muted-foreground">
					Named sets of variables (and secrets) applied to requests. The active
					environment is shared with the CLI.
				</p>
			</div>

			<form
				onSubmit={(e) => {
					e.preventDefault();
					void createForm.handleSubmit();
				}}
				className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4"
			>
				<p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
					New environment
				</p>
				<div className="flex flex-col gap-2 sm:flex-row">
					<createForm.Field
						name="name"
						validators={{
							onChange: ({ value }) =>
								value.trim()
									? undefined
									: 'Name the environment — e.g. "dev" or "staging".',
						}}
					>
						{(field) => (
							<span className="flex flex-col gap-1">
								<input
									name={field.name}
									value={field.state.value}
									onBlur={field.handleBlur}
									onChange={(e) => field.handleChange(e.target.value)}
									placeholder="name (e.g. dev)"
									aria-label="Environment name"
									spellCheck={false}
									className={`${inputClass} font-mono`}
								/>
								{field.state.meta.errors.length > 0 ? (
									<span className="text-xs text-destructive">
										{field.state.meta.errors[0]}
									</span>
								) : null}
							</span>
						)}
					</createForm.Field>
					<createForm.Field name="description">
						{(field) => (
							<input
								name={field.name}
								value={field.state.value}
								onBlur={field.handleBlur}
								onChange={(e) => field.handleChange(e.target.value)}
								placeholder="description (optional)"
								aria-label="Environment description"
								className={`${inputClass} flex-1`}
							/>
						)}
					</createForm.Field>
					<createForm.Subscribe
						selector={(state) => ({
							canSubmit: state.canSubmit,
							isSubmitting: state.isSubmitting,
						})}
					>
						{(sub: { canSubmit: boolean; isSubmitting: boolean }) => (
							<Button
								type="submit"
								disabled={!sub.canSubmit || sub.isSubmitting}
							>
								{sub.isSubmitting ? "Creating…" : "Create"}
							</Button>
						)}
					</createForm.Subscribe>
				</div>
				{createError && (
					<p className="text-xs text-destructive">{createError}</p>
				)}
			</form>

			{setActiveError && (
				<p className="text-xs text-destructive">{setActiveError}</p>
			)}

			{editing && (
				<div className="flex flex-col gap-3">
					<EnvironmentEditor
						key={`env-${editing.name}`}
						env={editing}
						onCancel={() => setEditingName(null)}
					/>
					<SecretsEditor
						key={`secrets-${editing.name}`}
						envName={editing.name}
						secretNames={editing.secrets}
						variableNames={Object.keys(editing.variables)}
					/>
				</div>
			)}

			{environmentsError ? (
				<p className="rounded-md border border-border bg-card p-3 text-xs text-destructive">
					{environmentsError}
				</p>
			) : environments.length === 0 ? (
				<div className="rounded-md border border-dashed border-border bg-card p-4 text-center">
					<p className="text-xs text-muted-foreground">
						No environments yet. Create one above to start modeling your
						targets.
					</p>
				</div>
			) : (
				<>
					{(() => {
						const active = environments.find(
							(e) => e.id === activeEnvironmentId,
						);
						if (
							active &&
							Object.keys(active.variables).length === 0 &&
							active.secrets.length === 0
						) {
							return (
								<p className="rounded-md border border-status-warn/40 bg-status-warn/10 p-3 text-xs text-status-warn">
									The active environment has no variables or secrets — requests
									resolve no values from it.
								</p>
							);
						}
						return null;
					})()}
					<ul className="flex flex-col gap-1.5">
						{environments.map((env) => {
							const active = env.id === activeEnvironmentId;
							const varCount = Object.keys(env.variables).length;
							const secretCount = env.secrets.length;
							return (
								<li
									key={env.id}
									className={cn(
										"flex items-center justify-between gap-3 rounded border bg-card/60 px-3 py-2 transition-colors select-none",
										active ? "border-primary/50 bg-primary/[0.02]" : "border-border hover:border-border/80",
									)}
								>
									<button
										type="button"
										onClick={() => onEdit(env.name)}
										className="flex min-w-0 flex-1 flex-col gap-0.5 text-left"
									>
										<div className="flex items-center gap-2">
											<span className="font-mono text-xs font-semibold text-foreground">{env.name}</span>
											{active && (
												<span className="rounded border border-primary/30 bg-primary/10 px-1.5 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-wider text-primary">
													active
												</span>
											)}
											<span className="font-mono text-[11px] tabular-nums text-muted-foreground">
												{varCount} var{varCount === 1 ? "" : "s"}
												{secretCount > 0 ? ` · ${secretCount} secret${secretCount === 1 ? "" : "s"}` : ""}
											</span>
										</div>
										{env.description && (
											<p className="truncate text-xs text-muted-foreground">
												{env.description}
											</p>
										)}
									</button>
									<div className="flex items-center gap-1.5">
										{!active && (
											<Button
												size="xs"
												variant="outline"
												onClick={() => void onSetActive(env.name)}
												className="font-mono text-[11px]"
											>
												Set active
											</Button>
										)}
										<Button
											size="xs"
											variant="ghost"
											onClick={() => onEdit(env.name)}
											className="font-mono text-[11px]"
										>
											{editingName === env.name ? "Editing" : "Edit"}
										</Button>
										<Button
											size="xs"
											variant="ghost"
											onClick={() => setDeletingName(env.name)}
											aria-label={`Delete ${env.name}`}
											className="text-destructive hover:bg-destructive/10"
										>
											Delete
										</Button>
									</div>
								</li>
							);
						})}
					</ul>
				</>
			)}

			<AlertDialog
				open={deletingName != null}
				onOpenChange={(open) => {
					if (!open) setDeletingName(null);
				}}
			>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>
							Delete environment "{deletingName}"?
						</AlertDialogTitle>
						<AlertDialogDescription>
							This removes its file and its secrets from disk. Requests that
							use these variables will stop resolving them.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Cancel</AlertDialogCancel>
						<AlertDialogAction
							className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
							onClick={() => void onDelete(deletingName ?? "")}
						>
							Delete
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
			<EnvToolsPanel />
		</div>
	);
}
