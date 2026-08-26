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
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#components/ui/select";
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

/** Per-request send overrides (timeout, redirect following), applied to the
 * tab's draft and persisted with the request file on save. Mounted only while
 * open, so local state reseeds from props each time. */
export function RequestSettingsDialog({
	settings,
	onApply,
	onClose,
}: RequestSettingsDialogProps) {
	const [timeoutText, setTimeoutText] = useState(() =>
		settings.timeout !== undefined ? String(settings.timeout) : "",
	);
	const [redirects, setRedirects] = useState<RedirectValue>(() =>
		redirectFor(settings.followRedirects),
	);

	const timeoutError =
		timeoutText.trim() !== "" &&
		(!/^\d+$/.test(timeoutText.trim()) || Number(timeoutText) <= 0)
			? "Timeout must be a positive number of milliseconds."
			: null;

	const apply = () => {
		if (timeoutError) return;
		const trimmed = timeoutText.trim();
		onApply({
			timeout: trimmed === "" ? undefined : Number(trimmed),
			followRedirects: valueFor(redirects),
		});
		onClose();
	};

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
						apply();
					}}
					className="flex flex-col gap-3"
				>
					<label className="flex flex-col gap-1 text-xs">
						<span className="font-medium text-foreground">Timeout (ms)</span>
						<input
							value={timeoutText}
							onChange={(e) => setTimeoutText(e.target.value)}
							placeholder="Default"
							inputMode="numeric"
							aria-invalid={Boolean(timeoutError)}
							className={inputClass}
						/>
						{timeoutError && (
							<span className="text-status-warn">{timeoutError}</span>
						)}
					</label>
					<label className="flex flex-col gap-1 text-xs">
						<span className="font-medium text-foreground">Redirects</span>
						<Select
							items={redirectOptions}
							value={redirects}
							onValueChange={(v) => {
								if (v === "default" || v === "on" || v === "off")
									setRedirects(v);
							}}
						>
							<SelectTrigger aria-label="Follow redirects" className="h-7 w-full text-xs">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								{redirectOptions.map((o) => (
									<SelectItem key={o.value} value={o.value} className="text-xs">
										{o.label}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</label>
					<DialogFooter>
						<Button type="button" variant="outline" onClick={onClose}>
							Cancel
						</Button>
						<Button type="submit" disabled={Boolean(timeoutError)}>
							Apply
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
