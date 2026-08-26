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
import { CompactSelect } from "#components/CompactSelect";
import { Field, FieldError, FieldGroup, FieldLabel } from "#components/ui/field";
import { Input } from "#components/ui/input";

export interface RequestSettings {
	timeout?: number;
	followRedirects?: boolean;
}

export interface RequestSettingsDialogProps {
	settings: RequestSettings;
	onApply: (settings: RequestSettings) => void;
	onClose: () => void;
}

/** Redirect select values: "default" (engine behavior), "on", "off". */
type RedirectValue = "default" | "on" | "off";

const redirectOptions = [
	{ value: "default", label: "Default (follow)" },
	{ value: "on", label: "Follow redirects" },
	{ value: "off", label: "Don't follow redirects" },
];

const redirectFor = (v: boolean | undefined): RedirectValue =>
	v === undefined ? "default" : v ? "on" : "off";

const valueFor = (v: RedirectValue): boolean | undefined =>
	v === "default" ? undefined : v === "on";

const formSchema = z.object({
	timeoutText: z
		.string()
		.refine(
			(v) => v.trim() === "" || (/^\d+$/.test(v.trim()) && Number(v.trim()) > 0),
			"Timeout must be a positive number of milliseconds.",
		),
	redirects: z.enum(["default", "on", "off"]),
});

/** Per-request send overrides (timeout, redirect following), applied to the
 * tab's draft and persisted with the request file on save. Mounted only while
 * open, so local state reseeds from props each time. */
export function RequestSettingsDialog({
	settings,
	onApply,
	onClose,
}: RequestSettingsDialogProps) {
	const form = useForm({
		defaultValues: {
			timeoutText:
				settings.timeout !== undefined ? String(settings.timeout) : "",
			redirects: redirectFor(settings.followRedirects),
		},
		validators: { onSubmit: formSchema },
		onSubmit: ({ value }) => {
			const trimmed = value.timeoutText.trim();
			onApply({
				timeout: trimmed === "" ? undefined : Number(trimmed),
				// SAFETY: redirects comes from the form typed as RedirectValue;
				// the assertion only restores the literal union for valueFor.
				followRedirects: valueFor(value.redirects as RedirectValue),
			});
			onClose();
		},
	});

	return (
		<Dialog open onOpenChange={(next) => !next && onClose()}>
			<DialogContent className="sm:max-w-sm">
				<DialogHeader>
					<DialogTitle>Request settings</DialogTitle>
					<DialogDescription>
						Per-request overrides; saved with the request file.
					</DialogDescription>
				</DialogHeader>
				<form
					onSubmit={(e) => {
						e.preventDefault();
						void form.handleSubmit();
					}}
					className="flex flex-col gap-3"
				>
					<FieldGroup>
						<form.Field
							name="timeoutText"
							children={(field) => {
								const isInvalid =
									field.state.meta.isTouched && !field.state.meta.isValid;
								return (
									<Field data-invalid={isInvalid}>
										<FieldLabel htmlFor="request-timeout">
											Timeout (ms)
										</FieldLabel>
										<Input
											id="request-timeout"
											name={field.name}
											value={field.state.value}
											onBlur={field.handleBlur}
											onChange={(e) => field.handleChange(e.target.value)}
											placeholder="Default"
											inputMode="numeric"
											aria-invalid={isInvalid}
										/>
										{isInvalid && (
											<FieldError errors={field.state.meta.errors} />
										)}
									</Field>
								);
							}}
						/>
						<form.Field
							name="redirects"
							children={(field) => (
								<Field>
									<FieldLabel htmlFor="request-redirects">Redirects</FieldLabel>
									<div id="request-redirects">
										<CompactSelect
											value={field.state.value}
											onChange={(v) =>
												field.handleChange(
													// SAFETY: options are the three RedirectValue
													// literals rendered above, so v is always one.
													v as RedirectValue,
												)
											}
											options={redirectOptions}
											ariaLabel="Follow redirects"
										/>
									</div>
								</Field>
							)}
						/>
					</FieldGroup>
					<DialogFooter>
						<Button type="button" variant="outline" onClick={onClose}>
							Cancel
						</Button>
						<Button type="submit">Apply</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
