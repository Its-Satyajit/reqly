import { useEffect, useState } from "react";
import { Fingerprint, ShieldCheck, Users, UserPlus, Mail } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Alert, AlertDescription } from "#components/ui/alert";
import { getSSOBridge, getSCIMBridge } from "#lib/sso";
import type { SCIMUser } from "#lib/sso";

export function SsoScimView() {
	const [issuer, setIssuer] = useState("https://example.okta.com");
	const [clientId, setClientId] = useState("0oa1…");
	const [token, setToken] = useState("");
	const [secret, setSecret] = useState("");
	const [ssoBusy, setSsoBusy] = useState(false);
	const [ssoResult, setSsoResult] = useState<string | null>(null);
	const [ssoError, setSsoError] = useState<string | null>(null);

	const [scimUserName, setScimUserName] = useState("");
	const [scimEmail, setScimEmail] = useState("");
	const [scimBusy, setScimBusy] = useState(false);
	const [scimError, setScimError] = useState<string | null>(null);
	const [scimOk, setScimOk] = useState<string | null>(null);
	const [users, setUsers] = useState<SCIMUser[]>([]);

	const loadUsers = async () => {
		try {
			const list = await getSCIMBridge().listUsers();
			setUsers(list ?? []);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setUsers([
					{ id: "user-1", userName: "alice", email: "alice@example.com", active: true },
					{ id: "user-2", userName: "bob", email: "bob@example.com", active: true },
				]);
			} else setScimError(msg);
		}
	};

	useEffect(() => {
		void loadUsers();
	}, []);

	const validate = async () => {
		if (!issuer.trim() || !clientId.trim() || !token.trim()) {
			setSsoError("issuer, clientId, and token are required");
			return;
		}
		setSsoBusy(true);
		setSsoError(null);
		setSsoResult(null);
		try {
			await getSSOBridge().validate(issuer.trim(), clientId.trim(), token.trim(), secret);
			setSsoResult("valid — issuer matches iss claim and HMAC verified");
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				// Mock: decode token and check issuer substring
				try {
					const payload = JSON.parse(atob(token.trim().split(".")[1] ?? ""));
					if (payload.iss === issuer.trim()) setSsoResult("valid (mock — iss matches)");
					else setSsoError(`issuer mismatch: got ${JSON.stringify(payload.iss)} want ${JSON.stringify(issuer.trim())} (mock)`);
				} catch {
					setSsoError("invalid token — not a JWT (mock)");
				}
			} else setSsoError(msg);
		} finally {
			setSsoBusy(false);
		}
	};

	const createUser = async () => {
		if (!scimUserName.trim()) {
			setScimError("userName is required");
			return;
		}
		setScimBusy(true);
		setScimError(null);
		setScimOk(null);
		try {
			const u = await getSCIMBridge().createUser(scimUserName.trim(), scimEmail.trim());
			setUsers((prev) => [...prev, u]);
			setScimOk(`created ${u.userName} (${u.id})`);
			setScimUserName("");
			setScimEmail("");
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				const mock: SCIMUser = { id: `user-${users.length + 1}`, userName: scimUserName.trim(), email: scimEmail.trim() || undefined, active: true };
				setUsers((prev) => [...prev, mock]);
				setScimOk(`created ${mock.userName} (mock)`);
				setScimUserName("");
				setScimEmail("");
			} else setScimError(msg);
		} finally {
			setScimBusy(false);
		}
	};

	return (
		<section className="flex min-h-0 flex-col gap-4 p-4" aria-label="SSO and SCIM">
			<PageHeader icon={Fingerprint} title="SSO & SCIM" description="Local OIDC validation (HMAC) and SCIM users — in-memory, zero telemetry, per-workspace." />

			<div className="grid gap-4 lg:grid-cols-2">
				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<ShieldCheck className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">SSO Validate</h3>
						<Badge variant="outline" className="ml-auto font-mono text-[10px]">OIDC · iss + HMAC</Badge>
					</div>

					<div className="flex flex-col gap-2 px-3 pb-3">
						<label className="flex flex-col gap-1">
							<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Issuer</span>
							<Input value={issuer} onChange={(e) => setIssuer(e.target.value)} placeholder="https://example.okta.com" className="font-mono text-xs" />
						</label>
						<label className="flex flex-col gap-1">
							<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Client ID</span>
							<Input value={clientId} onChange={(e) => setClientId(e.target.value)} placeholder="0oa..." className="font-mono text-xs" />
						</label>
						<label className="flex flex-col gap-1">
							<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">ID Token (JWT)</span>
							<Input value={token} onChange={(e) => setToken(e.target.value)} placeholder="eyJ..." spellCheck={false} className="font-mono text-xs" />
						</label>
						<label className="flex flex-col gap-1">
							<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">HMAC Secret (HS256)</span>
							<Input value={secret} onChange={(e) => setSecret(e.target.value)} type="password" placeholder="only for HS* in this demo" className="font-mono text-xs" />
						</label>

						{ssoError && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{ssoError}</AlertDescription></Alert>}
						{ssoResult && <p className="rounded border border-status-ok/20 bg-status-ok/10 px-2 py-1.5 font-mono text-xs text-status-ok">✓ {ssoResult}</p>}

						<Button size="sm" onClick={() => void validate()} disabled={ssoBusy} className="w-fit">
							{ssoBusy ? "Validating…" : "Validate"}
						</Button>

						<p className="font-mono text-[10px] text-muted-foreground">For M73: HMAC via <code className="rounded bg-muted px-1">internal/jwt.Verify</code>; RS256/JWKS deferred. Issuer must equal token <code className="rounded bg-muted px-1">iss</code>.</p>
					</div>
				</div>

				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<Users className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">SCIM Users</h3>
						<Badge variant="outline" className="ml-auto font-mono text-[10px]">{users.length} users · in-memory</Badge>
					</div>

					<div className="flex flex-col gap-3 px-3 pb-3">
						<div className="grid grid-cols-[1fr_1fr_auto] gap-2">
							<Input value={scimUserName} onChange={(e) => setScimUserName(e.target.value)} placeholder="userName (required)" className="font-mono text-xs" />
							<Input value={scimEmail} onChange={(e) => setScimEmail(e.target.value)} placeholder="email (optional)" className="font-mono text-xs" />
							<Button size="sm" onClick={() => void createUser()} disabled={scimBusy} className="gap-1.5">
								<UserPlus className="size-3.5" /> Create
							</Button>
						</div>

						{scimError && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{scimError}</AlertDescription></Alert>}
						{scimOk && <p className="font-mono text-xs text-status-ok">✓ {scimOk}</p>}

						<div className="overflow-auto rounded border border-border">
							<table className="w-full border-collapse text-xs">
								<thead className="border-b border-border bg-muted/30">
									<tr className="text-left font-mono text-[11px] text-muted-foreground">
										<th className="px-2.5 py-1.5 font-semibold">User</th>
										<th className="px-2.5 py-1.5 font-semibold">Email</th>
										<th className="px-2.5 py-1.5 font-semibold">ID</th>
										<th className="px-2.5 py-1.5 font-semibold">Active</th>
									</tr>
								</thead>
								<tbody>
									{users.length === 0 ? (
										<tr>
											<td colSpan={4} className="px-2.5 py-6 text-center font-mono text-xs text-muted-foreground">No users — create one above. In-memory, per-process, zero persistence.</td>
										</tr>
									) : (
										users.map((u) => (
											<tr key={u.id} className="border-b border-border/40 last:border-0">
												<td className="px-2.5 py-1.5 font-mono font-medium">{u.userName}</td>
												<td className="px-2.5 py-1.5 font-mono text-muted-foreground">
													{u.email ? (
														<span className="inline-flex items-center gap-1">
															<Mail className="size-3" aria-hidden />
															{u.email}
														</span>
													) : (
														"—"
													)}
												</td>
												<td className="px-2.5 py-1.5 font-mono text-[11px] text-muted-foreground">{u.id}</td>
												<td className="px-2.5 py-1.5">
													<span className={`inline-flex rounded border px-1.5 py-0.5 font-mono text-[10px] ${u.active ? "border-status-ok/20 bg-status-ok/10 text-status-ok" : "border-border bg-muted text-muted-foreground"}`}>{u.active ? "active" : "inactive"}</span>
												</td>
											</tr>
										))
									)}
								</tbody>
							</table>
						</div>

						<p className="font-mono text-[10px] text-muted-foreground">SCIM Store is local & in-memory for M73, mirroring audit/policy — later SQLite. UserName must be unique.</p>
					</div>
				</div>
			</div>
		</section>
	);
}
