import { useEffect, useMemo, useState } from "react";
import { Button } from "../../components";
import { useWorkspaceStore } from "../../stores";

const inputClass =
	"rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus-visible:border-ring";

interface SecretRow {
	name: string;
	value: string;
	touched: boolean;
	revealed: boolean;
	removing: boolean;
}

/**
 * SecretsEditor edits an environment's secrets without ever reading their
 * stored values back. Each existing secret is shown by name with a masked
 * placeholder; typing a new value masks it by default (per-row reveal
 * toggle) and marks it as a change. Save sends only the changed values plus
 * the names to remove — untouched secrets keep their on-disk values.
 */
export function SecretsEditor({
	envName,
	secretNames,
	variableNames,
}: {
	envName: string;
	secretNames: string[];
	variableNames: string[];
}) {
	const envAdapter = useWorkspaceStore((s) => s.envAdapter);
	const refreshEnvironments = useWorkspaceStore((s) => s.refreshEnvironments);
	const setEditorDirty = useWorkspaceStore((s) => s.setEditorDirty);

	const [rows, setRows] = useState<SecretRow[]>(() =>
		secretNames.map((name) => ({
			name,
			value: "",
			touched: false,
			revealed: false,
			removing: false,
		})),
	);
	const [newName, setNewName] = useState("");
	const [saving, setSaving] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const dirty = useMemo(
		() => rows.some((r) => r.touched || r.removing) || newName.trim() !== "",
		[rows, newName],
	);

	useEffect(() => {
		setEditorDirty(`secrets:${envName}`, dirty);
		return () => setEditorDirty(`secrets:${envName}`, false);
	}, [dirty, envName, setEditorDirty]);

	const duplicateName = useMemo(() => {
		const seen = new Set([...secretNames, ...variableNames]);
		return seen.has(newName.trim()) ? newName.trim() : null;
	}, [secretNames, variableNames, newName]);

	const setRow = (index: number, patch: Partial<SecretRow>) =>
		setRows((rs) => rs.map((r, i) => (i === index ? { ...r, ...patch } : r)));

	const onSave = async () => {
		setError(null);
		const values: Record<string, string> = {};
		const remove: string[] = [];
		for (const row of rows) {
			if (row.removing) {
				remove.push(row.name);
			} else if (row.touched) {
				if (row.value === "") {
					setError(
						`Secret "${row.name}" needs a value before it can be saved.`,
					);
					return;
				}
				values[row.name] = row.value;
			}
		}
		if (Object.keys(values).length === 0 && remove.length === 0) {
			return;
		}
		setSaving(true);
		try {
			await envAdapter.updateSecrets(envName, values, remove);
			await refreshEnvironments();
			setEditorDirty(`secrets:${envName}`, false);
		} catch (err) {
			setError(err instanceof Error ? err.message : String(err));
		} finally {
			setSaving(false);
		}
	};

	return (
		<div className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
			<div className="flex items-center justify-between">
				<p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
					Secrets
				</p>
				<span className="text-[10px] text-muted-foreground">
					values are never shown — only changed ones are written
				</span>
			</div>

			{rows.length === 0 && (
				<p className="text-xs text-muted-foreground">
					No secrets yet. Add one below to keep a value out of plain text.
				</p>
			)}

			<div className="flex flex-col gap-1.5">
				{rows.map((row, i) => (
					<div key={row.name} className="flex items-center gap-1.5">
						<span className="w-40 shrink-0 truncate font-mono text-xs">
							{row.name}
						</span>
						<input
							type={row.revealed ? "text" : "password"}
							value={row.value}
							onChange={(e) =>
								setRow(i, {
									value: e.target.value,
									touched: e.target.value !== "",
								})
							}
							disabled={row.removing}
							placeholder={
								row.removing ? "will be removed" : "unchanged — type to replace"
							}
							spellCheck={false}
							aria-label={`new value for secret ${row.name}`}
							className={`${inputClass} flex-1 font-mono`}
						/>
						<Button
							variant="ghost"
							size="icon-sm"
							aria-label={
								row.revealed
									? `mask secret ${row.name}`
									: `reveal secret ${row.name}`
							}
							disabled={row.removing || !row.touched}
							onClick={() => setRow(i, { revealed: !row.revealed })}
						>
							{row.revealed ? "hide" : "show"}
						</Button>
						<Button
							variant="ghost"
							size="icon-sm"
							aria-label={
								row.removing
									? `keep secret ${row.name}`
									: `remove secret ${row.name}`
							}
							onClick={() =>
								setRow(i, {
									removing: !row.removing,
									value: "",
									touched: false,
								})
							}
						>
							{row.removing ? "undo" : "remove"}
						</Button>
					</div>
				))}
				<div className="flex items-center gap-1.5">
					<input
						value={newName}
						onChange={(e) => setNewName(e.target.value)}
						placeholder="new secret name"
						spellCheck={false}
						className={`${inputClass} w-40 shrink-0 font-mono`}
					/>
					<Button
						variant="outline"
						size="sm"
						onClick={() => {
							if (!newName.trim() || duplicateName) return;
							setRows((rs) => [
								...rs,
								{
									name: newName.trim(),
									value: "",
									touched: false,
									revealed: false,
									removing: false,
								},
							]);
							setNewName("");
						}}
						disabled={!newName.trim() || !!duplicateName}
					>
						Add secret
					</Button>
					{duplicateName && (
						<span className="text-xs text-destructive">already exists</span>
					)}
				</div>
			</div>

			{error && <p className="text-xs text-destructive">{error}</p>}

			<div className="flex items-center justify-end">
				<Button
					size="sm"
					onClick={() => void onSave()}
					disabled={saving || !dirty}
				>
					Save secrets
				</Button>
			</div>
		</div>
	);
}
