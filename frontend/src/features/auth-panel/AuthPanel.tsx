import { useEffect, useState } from "react";
import { Button } from "../../components";
import { useAuthStore } from "../../stores";

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

function parseConfig(raw: string): Record<string, string> {
	const parsed: unknown = JSON.parse(raw);
	if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
		throw new Error("Config must be a JSON object");
	}
	const out: Record<string, string> = {};
	for (const [k, v] of Object.entries(parsed as Record<string, unknown>)) {
		out[k] = typeof v === "string" ? v : String(v);
	}
	return out;
}

function formatExpiry(expiry: string): string {
	if (!expiry) return "—";
	return new Date(expiry).toLocaleString();
}

export function AuthPanel() {
	const { status, loading, error, refresh, login, logout, clearError } =
		useAuthStore();
	const [flow, setFlow] = useState<"authorization_code" | "device_code">(
		"authorization_code",
	);
	const [configText, setConfigText] = useState(DEFAULT_CONFIG);
	const [configError, setConfigError] = useState<string | null>(null);

	useEffect(() => {
		void refresh();
	}, [refresh]);

	const onLogin = async () => {
		setConfigError(null);
		let config: Record<string, string>;
		try {
			config = parseConfig(configText);
		} catch (err) {
			setConfigError(err instanceof Error ? err.message : String(err));
			return;
		}
		clearError();
		await login(config, flow);
	};

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
					{tokens.map((tok, i) => (
						<li
							key={`${tok.endpoint}-${i}`}
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
					void onLogin();
				}}
			>
				<label className="flex flex-col gap-1">
					<span className="text-xs text-muted-foreground">Flow</span>
					<select
						value={flow}
						onChange={(e) =>
							setFlow(e.target.value as "authorization_code" | "device_code")
						}
						className="rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground"
					>
						<option value="authorization_code">
							Browser (authorization code + PKCE)
						</option>
						<option value="device_code">
							Device code (approve on any device)
						</option>
					</select>
				</label>

				<label className="flex flex-col gap-1">
					<span className="text-xs text-muted-foreground">
						OAuth 2.0 config (JSON)
					</span>
					<textarea
						value={configText}
						onChange={(e) => setConfigText(e.target.value)}
						rows={9}
						spellCheck={false}
						className="rounded-md border border-input bg-background p-2 font-mono text-xs text-foreground"
					/>
				</label>

				{configError ? (
					<p className="text-xs text-destructive">{configError}</p>
				) : null}
				{error ? <p className="text-xs text-destructive">{error}</p> : null}

				<div className="flex gap-2">
					<Button type="submit" disabled={loading} className="flex-1">
						{flow === "device_code" ? "Log in with device code" : "Log in"}
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
			</form>
		</div>
	);
}
