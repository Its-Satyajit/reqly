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
import { CompactSelect } from "#components/CompactSelect";
import { inputClass } from "#lib/ui";

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

interface SettingsFormValues {
	timeoutText: string;
	redirects: RedirectValue;
}

/** Per-request send overrides (timeout, redirect following), applied to the
 * tab's draft and persisted with the request file on save. Mounted only while
 * open, so local state reseeds from props each time. */
export function RequestSettingsDialog({
	settings,
	onApply,
	onClose,
}: RequestSettingsDialogProps) {
	const form = useForm({
		// SAFETY: shape matches SettingsFormValues; assertion keeps the field
		// types from collapsing to literal strings.
		defaultValues: {
			timeoutText:
				settings.timeout !== undefined ? String(settings.timeout) : "",
			redirects: redirectFor(settings.followRedirects),
		} as SettingsFormValues,
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
					<form.Field
						name="timeoutText"
						validators={{
							onChange: ({ value }) =>
								value.trim() !== "" &&
								(!/^\d+$/.test(value.trim()) || Number(value) <= 0)
									? "Timeout must be a positive number of milliseconds."
									: undefined,
						}}
					>
						{(field) => (
							<label className="flex flex-col gap-1 text-xs">
								<span className="font-medium text-foreground">
									Timeout (ms)
								</span>
								<input
									name={field.name}
									value={field.state.value}
									onBlur={field.handleBlur}
									onChange={(e) => field.handleChange(e.target.value)}
									placeholder="Default"
									inputMode="numeric"
									aria-invalid={field.state.meta.errors.length > 0}
									className={inputClass}
								/>
								{field.state.meta.errors.length > 0 ? (
									<span className="text-status-warn">
										{field.state.meta.errors[0]}
									</span>
								) : null}
							</label>
						)}
					</form.Field>
					<form.Field name="redirects">
						{(field) => (
							<label className="flex flex-col gap-1 text-xs">
								<span className="font-medium text-foreground">Redirects</span>
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
							</label>
						)}
					</form.Field>
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
