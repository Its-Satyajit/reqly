/* eslint-disable react/no-children-prop */
import { useEffect } from "react";
import { useForm } from "@tanstack/react-form";
import * as z from "zod";
import { Button } from "#components/ui/button";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#components/ui/select";
import { Field, FieldError, FieldGroup, FieldLabel } from "#components/ui/field";
import { useAuthStore } from "#stores/useAuthStore";
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

const formSchema = z.object({
	flow: z.enum(["authorization_code", "device_code"]),
	configText: z.string().superRefine((v, ctx) => {
		try {
			parseConfig(v);
		} catch (err) {
			ctx.addIssue({
				code: "custom",
				message: err instanceof Error ? err.message : String(err),
			});
		}
	}),
});

type AuthFormValues = {
	flow: "authorization_code" | "device_code";
	configText: string;
};

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
		validators: { onSubmit: formSchema },
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
				onSubmit={(e) => {
					e.preventDefault();
					void form.handleSubmit();
				}}
			>
				<FieldGroup>
					<form.Field
						name="flow"
						children={(field) => (
							<Field>
								<FieldLabel htmlFor="auth-flow">Flow</FieldLabel>
								<div id="auth-flow">
									<Select
										items={[
											{
												value: "authorization_code",
												label: "Browser (auth code + PKCE)",
											},
											{ value: "device_code", label: "Device code" },
										]}
										value={field.state.value}
										onValueChange={(v) => {
											if (v !== null)
												field.handleChange(
													// SAFETY: options are constrained to
													// authorization_code | device_code
													v as typeof field.state.value,
												);
										}}
									>
										<SelectTrigger aria-label="OAuth flow" className="h-7 w-auto gap-1 rounded-md px-2 text-xs">
											<SelectValue />
										</SelectTrigger>
										<SelectContent className="max-h-72 min-w-(--anchor-width)">
											{[
												{
													value: "authorization_code",
													label: "Browser (auth code + PKCE)",
												},
												{ value: "device_code", label: "Device code" },
											].map((option) => (
												<SelectItem key={option.value} value={option.value} className="text-xs">
													{option.label}
												</SelectItem>
											))}
										</SelectContent>
									</Select>
								</div>
							</Field>
						)}
					/>

					<form.Field
						name="configText"
						validators={{
						onChange: ({ value }) => {
							try {
								parseConfig(value);
								return undefined;
							} catch (err) {
								return {
									message:
										err instanceof Error ? err.message : String(err),
								};
							}
						},
						}}
						children={(field) => {
							const isInvalid =
								field.state.meta.isTouched && !field.state.meta.isValid;
							return (
								<Field data-invalid={isInvalid}>
									<FieldLabel htmlFor="auth-config">
										OAuth 2.0 config (JSON)
									</FieldLabel>
									<textarea
										id="auth-config"
										name={field.name}
										value={field.state.value}
										onBlur={field.handleBlur}
										onChange={(e) => field.handleChange(e.target.value)}
										rows={9}
										spellCheck={false}
										aria-invalid={isInvalid}
										className="rounded-md border border-input bg-background p-2 font-mono text-xs text-foreground aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20"
									/>
									{isInvalid && <FieldError errors={field.state.meta.errors} />}
								</Field>
							);
						}}
					/>
				</FieldGroup>

				{error ? (
					<p className="mt-2 text-xs text-destructive">{error}</p>
				) : null}

				<div className="mt-3 flex gap-2">
					<form.Subscribe
						selector={(s) => ({
							canSubmit: s.canSubmit,
							isSubmitting: s.isSubmitting,
							flow: s.values.flow,
						})}
					>
						{({ canSubmit, isSubmitting, flow }: { canSubmit: boolean; isSubmitting: boolean; flow: "authorization_code" | "device_code" }) => (
							<Button
								type="submit"
								disabled={loading || !canSubmit || isSubmitting}
								className="flex-1"
							>
								{flow === "device_code"
									? "Log in with device code"
									: "Log in"}
							</Button>
						)}
					</form.Subscribe>
					<Button
						type="button"
						variant="outline"
						disabled={!canLogout}
						onClick={() => void logout()}
					>
						Log out
					</Button>
				</div>
			</form>
		</div>
	);
}
