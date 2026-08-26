/* eslint-disable react/no-children-prop */
import { useForm } from "@tanstack/react-form";
import * as z from "zod";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#components/ui/dialog";
import { Button } from "#components/ui/button";
import { Field, FieldError, FieldGroup, FieldLabel } from "#components/ui/field";
import { Input } from "#components/ui/input";

const formSchema = z.object({
	name: z.string().trim().min(1, "Name is required."),
});

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
		validators: { onSubmit: formSchema },
		onSubmit: async ({ value }) => {
			const createError = await onCreate(value.name.trim());
			if (createError !== null) {
				form.setFieldMeta("name", (meta) => ({
					...meta,
					errors: [{ message: createError }],
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
					<FieldGroup>
						<form.Field
							name="name"
							children={(field) => {
								const isInvalid =
									field.state.meta.isTouched && !field.state.meta.isValid;
								return (
									<Field data-invalid={isInvalid}>
										<FieldLabel htmlFor="container-name">Name</FieldLabel>
										<Input
											id="container-name"
											name={field.name}
											value={field.state.value}
											onBlur={field.handleBlur}
											onChange={(e) => field.handleChange(e.target.value)}
											placeholder="name (e.g. payments)"
											spellCheck={false}
											aria-invalid={isInvalid}
										/>
										{isInvalid && (
											<FieldError errors={field.state.meta.errors} />
										)}
									</Field>
								);
							}}
						/>
					</FieldGroup>
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
