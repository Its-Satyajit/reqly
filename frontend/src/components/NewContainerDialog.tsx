import { useForm } from "@tanstack/react-form";
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

/** Name prompt for scaffolding a collection or folder descriptor on disk.
 * Mounted only while open (parent conditionally renders it), so state starts
 * fresh on every open without reset effects. */
export function NewContainerDialog({
	title,
	description,
	onCreate,
	onClose,
}: NewContainerDialogProps) {
	const form = useForm({
		defaultValues: { name: "" },
		onSubmit: async ({ value }) => {
			const createError = await onCreate(value.name.trim());
			if (createError !== null) {
				form.setFieldMeta("name", (meta) => ({
					...meta,
					errors: [createError],
				}));
				return;
			}
			onClose();
		},
	});

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
						void form.handleSubmit();
					}}
					className="flex flex-col gap-2"
				>
					<form.Field
						name="name"
						validators={{
							onChange: ({ value }) =>
								value.trim() ? undefined : "Name is required.",
						}}
					>
						{(field) => (
							<span className="flex flex-col gap-2">
								<input
									name={field.name}
									value={field.state.value}
									onBlur={field.handleBlur}
									onChange={(e) => field.handleChange(e.target.value)}
									placeholder="name (e.g. payments)"
									aria-label="Container name"
									spellCheck={false}
									className={inputClass}
								/>
								{field.state.meta.errors.length > 0 ? (
									<p className="text-xs text-destructive">
										{field.state.meta.errors[0]}
									</p>
								) : null}
							</span>
						)}
					</form.Field>
					<form.Subscribe
						selector={(state) => ({
							canSubmit: state.canSubmit,
							isSubmitting: state.isSubmitting,
						})}
					>
						{({ canSubmit, isSubmitting }: { canSubmit: boolean; isSubmitting: boolean }) => (
							<DialogFooter>
								<Button
									type="button"
									variant="ghost"
									size="sm"
									onClick={onClose}
								>
									Cancel
								</Button>
								<Button type="submit" size="sm" disabled={!canSubmit}>
									{isSubmitting ? "Creating…" : "Create"}
								</Button>
							</DialogFooter>
						)}
					</form.Subscribe>
				</form>
			</DialogContent>
		</Dialog>
	);
}
