import { useEffect } from "react";
import { useForm } from "@tanstack/react-form";
import { Button } from "../../components/ui/button";
import { CompactSelect } from "#components/CompactSelect";
import { useAuthStore } from "../../stores/useAuthStore";
import { isRecord, isString, type JsonObject, type JsonValue } from "../../lib/typeGuards";

const DEFAULT_CONFIG = JSON.stringify(
	{
		authorization_url: "https://idp.example.com/authorize",
		token_url: "https://idp.example.com/token",
		client_id: "",
		client_secret: "",
	},
	null,
	2,
);

function parseConfig(raw: string) {
	// SAFETY: JSON config parsed at I/O boundary; shape validated via isRecord below
	const parsed = JSON.parse(raw) as JsonValue;
	if (!isRecord(parsed)) {
		throw new Error("Config must be a JSON object");
	}
	const out: Record<string, string> = {};
	// SAFETY: isRecord narrows parsed to JsonObject with JsonValue values
	for (const [k, v] of Object.entries(parsed as JsonObject)) {
		out[k] = isString(v) ? v : String(v);
	}
	return out;
}

function formatExpiry(expiry: string): string {
	if (!expiry) return "—";
	return new Date(expiry).toLocaleString();
}

interface AuthFormValues {
	flow: "authorization_code" | "device_code";
	configText: string;
}

export function AuthPanel() {
	const { status, loading, error, refresh, login, logout, clearError } =
		useAuthStore();

	const form = useForm({
		// SAFETY: literal matches AuthFormValues["flow"]; assertion widens the
		// default so the form field accepts both flow options.
		defaultValues: {
			flow: "authorization_code",
			configText: DEFAULT_CONFIG,
		} as AuthFormValues,
		onSubmit: async ({ value }) => {
			const config = parseConfig(value.configText);
			clearError();
			await login(config, value.flow);
		},
	});

	useEffect(() => {
		void refresh();
	}, [refresh]);

	const tokens = status?.tokens ?? [];
	const canLogout = tokens.length > 0 && !loading;

	return (
		<div id="auth-panel" className="flex flex-col gap-3 px-2 pt-4 scroll-mt-2">
			<p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
				OAuth tokens
			</p>

			{status ? (
				<p className="text-xs text-muted-foreground">Store: {status.backend}</p>
			) : null}

			{tokens.length === 0 ? (
				<p className="text-xs text-muted-foreground">
					No cached tokens. Log in to authenticate requests for this workspace.
				</p>
			) : (
				<ul className="flex flex-col gap-2">
					{tokens.map((tok) => (
						<li
							key={`${tok.endpoint}-${tok.expiry}-${tok.grantType}`}
							className="rounded-md border border-border p-2 text-xs"
						>
							<p className="truncate text-foreground" title={tok.endpoint}>
								{tok.endpoint}
							</p>
							<p className="mt-1 text-muted-foreground">
								{tok.grantType} · {tok.accessToken}
							</p>
							<p className="text-muted-foreground">
								Expires {formatExpiry(tok.expiry)} ·{" "}
								{tok.hasRefresh ? "refresh token" : "no refresh token"} ·{" "}
								<span
									className={
										tok.state === "expired" ? "text-destructive" : undefined
									}
								>
									{tok.state}
								</span>
							</p>
						</li>
					))}
				</ul>
			)}

			<form
				className="flex flex-col gap-2"
				onSubmit={(e) => {
					e.preventDefault();
					void form.handleSubmit();
				}}
			>
				<form.Field name="flow">
					{(field) => (
						<div className="flex flex-col gap-1">
							<span className="text-xs text-muted-foreground">Flow</span>
							<CompactSelect
								value={field.state.value}
								onChange={(v) =>
									field.handleChange(
										// SAFETY: options are constrained to
										// authorization_code | device_code
										v as AuthFormValues["flow"],
									)
								}
								ariaLabel="OAuth flow"
								options={[
									{
										value: "authorization_code",
										label: "Browser (auth code + PKCE)",
									},
									{ value: "device_code", label: "Device code" },
								]}
							/>
						</div>
					)}
				</form.Field>

				<form.Field
					name="configText"
					validators={{
						onChange: ({ value }) => {
							try {
								parseConfig(value);
								return undefined;
							} catch (err) {
								return err instanceof Error ? err.message : String(err);
							}
						},
					}}
				>
					{(field) => (
						<label className="flex flex-col gap-1">
							<span className="text-xs text-muted-foreground">
								OAuth 2.0 config (JSON)
							</span>
							<textarea
								name={field.name}
								value={field.state.value}
								onBlur={field.handleBlur}
								onChange={(e) => field.handleChange(e.target.value)}
								rows={9}
								spellCheck={false}
								className="rounded-md border border-input bg-background p-2 font-mono text-xs text-foreground"
							/>
						</label>
					)}
				</form.Field>

				<form.Subscribe
					selector={(state) => ({
						errors: state.fieldMeta.configText?.errors ?? [],
						canSubmit: state.canSubmit,
						isSubmitting: state.isSubmitting,
						flow: state.values.flow,
					})}
				>
					{({
						errors,
						canSubmit,
						isSubmitting,
						flow,
					}: {
						errors: unknown[];
						canSubmit: boolean;
						isSubmitting: boolean;
						flow: AuthFormValues["flow"];
					}) => (
						<>
							{errors.length > 0 ? (
								<p className="text-xs text-destructive">{String(errors[0])}</p>
							) : null}
							{error ? (
								<p className="text-xs text-destructive">{error}</p>
							) : null}

							<div className="flex gap-2">
								<Button type="submit" disabled={loading || !canSubmit || isSubmitting} className="flex-1">
									{flow === "device_code"
										? "Log in with device code"
										: "Log in"}
								</Button>
								<Button
									type="button"
									variant="outline"
									disabled={!canLogout}
									onClick={() => void logout()}
								>
									Log out
								</Button>
							</div>
						</>
					)}
				</form.Subscribe>
			</form>
		</div>
	);
}
