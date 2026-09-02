import { useEffect, useState } from "react";
import { Users, UserPlus, UserMinus, Server, Copy, Shield } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { CompactSelect } from "#components/CompactSelect";
import { Alert, AlertDescription } from "#components/ui/alert";
import { cn } from "#lib/utils";
import { copyText } from "#lib/response";
import { getCollabBridge, COLLAB_ROLES, roleBadgeClass } from "#lib/collab";
import type { Collaborator } from "#lib/collab";

function formatTime(iso: string): string {
	try {
		const d = new Date(iso);
		return d.toLocaleString(undefined, { month: "short", day: "2-digit", hour: "2-digit", minute: "2-digit" });
	} catch {
		return iso;
	}
}

export function CollabView() {
	const [collabs, setCollabs] = useState<Collaborator[]>([]);
	const [user, setUser] = useState("");
	const [role, setRole] = useState("viewer");
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [ok, setOk] = useState<string | null>(null);

	const [servePort, setServePort] = useState("8080");
	const [serveUrl, setServeUrl] = useState<string | null>(null);
	const [serveBusy, setServeBusy] = useState(false);
	const [serveError, setServeError] = useState<string | null>(null);

	const load = async () => {
		try {
			const list = await getCollabBridge().list();
			setCollabs(list ?? []);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setCollabs([
					{ user: "alice", role: "admin", addedAt: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString() },
					{ user: "bob", role: "editor", addedAt: new Date(Date.now() - 1000 * 60 * 30).toISOString() },
				]);
			} else setError(msg);
		}
	};

	useEffect(() => {
		void load();
	}, []);

	const add = async () => {
		if (!user.trim()) {
			setError("user is required");
			return;
		}
		setBusy(true);
		setError(null);
		setOk(null);
		try {
			await getCollabBridge().add(user.trim(), role);
			setOk(`added ${user.trim()} as ${role}`);
			setUser("");
			await load();
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setCollabs((prev) => {
					const exists = prev.find((c) => c.user === user.trim());
					if (exists) return prev.map((c) => (c.user === user.trim() ? { ...c, role } : c));
					return [...prev, { user: user.trim(), role, addedAt: new Date().toISOString() }];
				});
				setOk(`added ${user.trim()} as ${role} (mock — .reqly/collab.yaml 0600)`);
				setUser("");
			} else setError(msg);
		} finally {
			setBusy(false);
		}
	};

	const remove = async (target: string) => {
		setError(null);
		setOk(null);
		try {
			await getCollabBridge().remove(target);
			setOk(`removed ${target}`);
			await load();
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setCollabs((prev) => prev.filter((c) => c.user !== target));
				setOk(`removed ${target} (mock)`);
			} else setError(msg);
		}
	};

	const serve = async () => {
		const port = Number(servePort);
		if (!Number.isFinite(port) || port <= 0 || port > 65535) {
			setServeError("port must be 1–65535");
			return;
		}
		setServeBusy(true);
		setServeError(null);
		setServeUrl(null);
		try {
			const url = await getCollabBridge().serve(port);
			setServeUrl(url);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setServeUrl(`http://localhost:${port} (mock — no server in dev)`);
			} else setServeError(msg);
		} finally {
			setServeBusy(false);
		}
	};

	return (
		<section className="flex min-h-0 flex-col gap-4 p-4" aria-label="Collaboration">
			<PageHeader icon={Users} title="Collaboration" description="Git-native shared workspaces — collaborators in .reqly/collab.yaml (0600), committed to Git for sharing." />

			<div className="grid gap-4 lg:grid-cols-[1.4fr_0.9fr]">
				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<Shield className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">Roster</h3>
						<Badge variant="outline" className="ml-auto font-mono text-[10px]">{collabs.length} members</Badge>
					</div>

					<div className="flex flex-col gap-3 px-3 pb-3">
						<div className="grid grid-cols-[1fr_170px_auto] gap-2">
							<Input value={user} onChange={(e) => setUser(e.target.value)} placeholder="user — e.g. alice" className="font-mono text-xs" />
							<CompactSelect value={role} onChange={setRole} options={COLLAB_ROLES.map((r) => ({ value: r.value, label: r.label }))} ariaLabel="Role" />
							<Button size="sm" onClick={() => void add()} disabled={busy} className="gap-1.5">
								<UserPlus className="size-3.5" /> Add
							</Button>
						</div>

						{error && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{error}</AlertDescription></Alert>}
						{ok && <p className="font-mono text-xs text-status-ok">✓ {ok}</p>}

						<div className="overflow-auto rounded border border-border">
							<table className="w-full border-collapse text-xs">
								<thead className="border-b border-border bg-muted/30">
									<tr className="text-left font-mono text-[11px] text-muted-foreground">
										<th className="px-2.5 py-1.5 font-semibold">User</th>
										<th className="px-2.5 py-1.5 font-semibold">Role</th>
										<th className="px-2.5 py-1.5 font-semibold">Added</th>
										<th className="px-2.5 py-1.5 font-semibold text-right">Action</th>
									</tr>
								</thead>
								<tbody>
									{collabs.length === 0 ? (
										<tr>
											<td colSpan={4} className="px-2.5 py-8 text-center font-mono text-xs text-muted-foreground">No collaborators yet — add one above. Stored in .reqly/collab.yaml, 0600, Git-native.</td>
										</tr>
									) : (
										collabs.map((c) => (
											<tr key={c.user} className="border-b border-border/40 last:border-0">
												<td className="px-2.5 py-1.5 font-mono font-medium">{c.user}</td>
												<td className="px-2.5 py-1.5">
													<span className={cn("inline-flex rounded border px-1.5 py-0.5 font-mono text-[10px] font-medium", roleBadgeClass(c.role))}>{c.role}</span>
												</td>
												<td className="px-2.5 py-1.5 font-mono text-[11px] text-muted-foreground">{formatTime(c.addedAt)}</td>
												<td className="px-2.5 py-1.5 text-right">
													<Button variant="ghost" size="xs" onClick={() => void remove(c.user)} className="gap-1 text-status-error hover:text-status-error">
														<UserMinus className="size-3.5" /> Remove
													</Button>
												</td>
											</tr>
										))
									)}
								</tbody>
							</table>
						</div>

						<p className="font-mono text-[10px] text-muted-foreground">Roles: viewer (send & export), editor (runs & imports), admin (all incl. policy/rbac). File is 0600, committed for sharing.</p>
					</div>
				</div>

				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<Server className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">Serve</h3>
						<Badge variant="outline" className="ml-auto font-mono text-[10px]">local · {serveUrl ? "running" : "idle"}</Badge>
					</div>

					<div className="flex flex-col gap-3 px-3 pb-3">
						<p className="font-mono text-xs text-muted-foreground">Start a self-hosted collaboration endpoint for this workspace — local-only, zero telemetry, URL is shareable on your LAN.</p>

						<label className="flex flex-col gap-1">
							<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Port</span>
							<div className="flex gap-2">
								<Input value={servePort} onChange={(e) => setServePort(e.target.value)} placeholder="8080" className="font-mono text-xs" inputMode="numeric" />
								<Button size="sm" onClick={() => void serve()} disabled={serveBusy} className="shrink-0 gap-1.5">
									<Server className="size-3.5" /> Serve
								</Button>
							</div>
						</label>

						{serveError && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{serveError}</AlertDescription></Alert>}

						{serveUrl ? (
							<div className="flex items-center gap-2 rounded border border-status-ok/20 bg-status-ok/10 px-2.5 py-2">
								<span className="size-2 rounded-full bg-status-ok animate-pulse" aria-hidden />
								<span className="truncate font-mono text-xs text-status-ok">{serveUrl}</span>
								<Button variant="ghost" size="xs" onClick={() => void copyText(serveUrl)} className="ml-auto gap-1">
									<Copy className="size-3.5" /> Copy
								</Button>
							</div>
						) : (
							<p className="rounded border border-dashed border-border bg-muted/20 px-2.5 py-2 font-mono text-xs text-muted-foreground">No server running — choose a port and serve. In dev, this is mocked.</p>
						)}

						<div className="rounded border border-border/60 bg-muted/20 p-2.5">
							<p className="font-mono text-[11px] font-semibold tracking-wide text-muted-foreground">How sharing works</p>
							<ul className="mt-1.5 list-disc space-y-0.5 pl-4 font-mono text-[11px] text-muted-foreground">
								<li>Edit <code className="rounded bg-muted px-1">.reqly/collab.yaml</code> is committed — teammates see the roster on pull.</li>
								<li>Serve is ephemeral — not persisted, per-process, local LAN only.</li>
								<li>All files remain Git-native; no cloud, no account.</li>
							</ul>
						</div>
					</div>
				</div>
			</div>
		</section>
	);
}
