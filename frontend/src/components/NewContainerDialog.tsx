import { useState } from "react";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#components/ui/dialog";
import { Button } from "#components/ui/button";
import { inputClass } from "#lib/ui";

export interface NewContainerDialogProps {
	/** Dialog headline, e.g. "New collection" or "New folder in payments". */
	title: string;
	description: string;
	/** Creates the container. Returns an error message, or null on success. */
	onCreate: (name: string) => Promise<string | null>;
	onClose: () => void;
}

/** Runs a create call, mapping failures to a display message. Module-level
 * because React Compiler cannot handle try/catch, and owns the busy flag so
 * its reset sits in a finally. */
async function runCreate(
	create: (name: string) => Promise<unknown>,
	name: string,
	setBusy: (busy: boolean) => void,
): Promise<string | null> {
	setBusy(true);
	try {
		await create(name);
		return null;
	} catch (err) {
		return err instanceof Error ? err.message : String(err);
	} finally {
		setBusy(false);
	}
}

/** Name prompt for scaffolding a collection or folder descriptor on disk.
 * Mounted only while open (parent conditionally renders it), so state starts
 * fresh on every open without reset effects. */
export function NewContainerDialog({
	title,
	description,
	onCreate,
	onClose,
}: NewContainerDialogProps) {
	const [name, setName] = useState("");
	const [error, setError] = useState<string | null>(null);
	const [busy, setBusy] = useState(false);

	const canCreate = name.trim() !== "" && !busy;

	const submit = async () => {
		if (!canCreate) return;
		const createError = await runCreate(onCreate, name.trim(), setBusy);
		if (createError !== null) {
			setError(createError);
			return;
		}
		onClose();
	};

	return (
		<Dialog open onOpenChange={(next) => !next && onClose()}>
			<DialogContent className="sm:max-w-sm">
				<DialogHeader>
					<DialogTitle>{title}</DialogTitle>
					<DialogDescription>{description}</DialogDescription>
				</DialogHeader>
				<form
					onSubmit={(e) => {
						e.preventDefault();
						void submit();
					}}
					className="flex flex-col gap-2"
				>
					<input
						value={name}
						onChange={(e) => setName(e.target.value)}
						placeholder="name (e.g. payments)"
						aria-label="Container name"
						spellCheck={false}
						className={inputClass}
					/>
					{error && <p className="text-xs text-destructive">{error}</p>}
					<DialogFooter>
						<Button type="button" variant="ghost" size="sm" onClick={onClose}>
							Cancel
						</Button>
						<Button type="submit" size="sm" disabled={!canCreate}>
							{busy ? "Creating…" : "Create"}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
