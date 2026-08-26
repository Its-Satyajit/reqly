import { useEffect, useState } from "react";
import { Button } from "../../components/ui/button";
import type { Environment } from "../../stores/useWorkspaceStore";
import { useWorkspaceStore } from "../../stores/useWorkspaceStore";
import type { EnvAdapter } from "../../lib/env";
import { inputClass } from "../../lib/ui";

interface VariableRow {
	key: string;
	value: string;
}

/** Persists the editor draft. Kept at module level (outside the compiled
 * component) because React Compiler cannot handle try/catch. */
async function persistEnvironment(
	envAdapter: EnvAdapter,
	refresh: () => Promise<void>,
	setSaving: (saving: boolean) => void,
	name: string,
	description: string,
	variables: Record<string, string>,
): Promise<string | null> {
	setSaving(true);
	try {
		await envAdapter.update(name, description, variables);
		await refresh();
		return null;
	} catch (err) {
		return err instanceof Error ? err.message : String(err);
	} finally {
		setSaving(false);
	}
}

/**
 * EnvironmentEditor edits a single environment's description and variables in
 * an in-memory draft. The file system is the source of truth: nothing is
 * written until Save, and unsaved edits are flagged on the store so sidebar
 * navigation confirms before discarding them. Secret values are never shown
 * or edited here (separate surface).
 */
export function EnvironmentEditor({
	env,
	onCancel,
}: {
	env: Environment;
	onCancel: () => void;
}) {
	const envAdapter = useWorkspaceStore((s) => s.envAdapter);
	const refreshEnvironments = useWorkspaceStore((s) => s.refreshEnvironments);
	const setEditorDirty = useWorkspaceStore((s) => s.setEditorDirty);

	const [description, setDescription] = useState(env.description);
	const [rows, setRows] = useState<VariableRow[]>(() =>
		Object.entries(env.variables).map(([key, value]) => ({ key, value })),
	);
	const [saving, setSaving] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const dirty = (() => {
		if (description !== env.description) return true;
		const current = new Map(Object.entries(env.variables));
		if (rows.length !== current.size) return true;
		for (const row of rows) {
			if (current.get(row.key) !== row.value) return true;
			current.delete(row.key);
		}
		return current.size > 0;
	})();

	useEffect(() => {
		setEditorDirty(`vars:${env.name}`, dirty);
		return () => setEditorDirty(`vars:${env.name}`, false);
	}, [dirty, env.name, setEditorDirty]);

	const duplicateKey = (() => {
		const seen = new Map<string, number>();
		for (const row of rows) {
			const key = row.key.trim();
			if (!key) continue;
			const n = seen.get(key) ?? 0;
			seen.set(key, n + 1);
			if (seen.get(key) === 2) return key;
		}
		return null;
	})();

	// Matches the CLI's secret-exposure heuristic (key/token/secret/password/
	// credential): variables whose names look like secrets are a warning, not a
	// hard error — the user may intend them as secrets or plain variables.
	const secretLikeWarnings = (() => {
		const pattern = /(key|token|secret|password|credential)/i;
		const seen = new Set<string>();
		const warnings: string[] = [];
		for (const row of rows) {
			const key = row.key.trim();
			if (!key || seen.has(key)) continue;
			seen.add(key);
			if (pattern.test(key)) warnings.push(key);
		}
		return warnings;
	})();

	const setRow = (index: number, patch: Partial<VariableRow>) =>
		setRows((rs) => rs.map((r, i) => (i === index ? { ...r, ...patch } : r)));

	const onSave = async () => {
		setError(null);
		if (duplicateKey) {
			setError(`Variable "${duplicateKey}" is defined more than once.`);
			return;
		}
		const variables: Record<string, string> = {};
		for (const row of rows) {
			if (row.key.trim()) variables[row.key.trim()] = row.value;
		}
		const saveError = await persistEnvironment(
			envAdapter,
			refreshEnvironments,
			setSaving,
			env.name,
			description.trim(),
			variables,
		);
		if (saveError !== null) {
			setError(saveError);
			return;
		}
		setEditorDirty(`vars:${env.name}`, false);
		onCancel();
	};

	return (
		<div className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
			<div className="flex items-center justify-between">
				<p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
					Edit{" "}
					<span className="font-mono normal-case text-foreground">
						{env.name}
					</span>
				</p>
				<span className="text-2xs text-muted-foreground">
					{dirty ? "unsaved changes" : "saved"}
				</span>
			</div>

			<div className="flex flex-col gap-1">
				<label
					htmlFor="env-description"
					className="text-xs text-muted-foreground"
				>
					Description
				</label>
				<input
					id="env-description"
					value={description}
					onChange={(e) => setDescription(e.target.value)}
					placeholder="what this environment is for"
					className={inputClass}
				/>
			</div>

			<div className="flex flex-col gap-1">
				<p className="text-xs text-muted-foreground">Variables</p>
				{rows.length === 0 && (
					<p className="text-xs text-muted-foreground">
						No variables yet — add one below.
					</p>
				)}
				<div className="flex flex-col gap-1.5">
					{rows.map((row, i) => (
						// Rows are anonymous key/value drafts (often blank) with no stable
						// identity — positional keys are the only correct choice.
						// react-doctor-disable-next-line react-doctor/no-array-index-as-key
						<div key={i} className="flex items-center gap-1.5">
							<input
								value={row.key}
								onChange={(e) => setRow(i, { key: e.target.value })}
								placeholder="KEY"
								spellCheck={false}
								aria-label={`variable ${i + 1} name`}
								className={`${inputClass} w-40 shrink-0 font-mono`}
							/>
							<input
								value={row.value}
								onChange={(e) => setRow(i, { value: e.target.value })}
								placeholder="value"
								spellCheck={false}
								aria-label={`variable ${i + 1} value`}
								className={`${inputClass} flex-1 font-mono`}
							/>
							<Button
								variant="ghost"
								size="icon-sm"
								aria-label={`remove variable ${i + 1}`}
								onClick={() => setRows((rs) => rs.filter((_, j) => j !== i))}
							>
								×
							</Button>
						</div>
					))}
				</div>
				<div>
					<Button
						variant="outline"
						size="sm"
						onClick={() => setRows((rs) => [...rs, { key: "", value: "" }])}
					>
						Add variable
					</Button>
				</div>
				{secretLikeWarnings.length > 0 && (
					<p className="text-xs text-status-warn">
						{secretLikeWarnings.join(", ")} look
						{secretLikeWarnings.length === 1 ? "s" : ""} like a secret —
						consider moving it to the Secrets section.
					</p>
				)}
			</div>

			{error && <p className="text-xs text-destructive">{error}</p>}

			<div className="flex items-center justify-end gap-2">
				<Button variant="ghost" size="sm" onClick={onCancel} disabled={saving}>
					Cancel
				</Button>
				<Button
					size="sm"
					onClick={() => void onSave()}
					disabled={saving || !dirty || !!duplicateKey}
				>
					Save changes
				</Button>
			</div>
		</div>
	);
}
