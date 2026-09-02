import { useEffect, useState } from "react";
import { Shield, ShieldCheck, ShieldAlert, Users, FileCheck, Play, Copy } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Textarea } from "#components/ui/textarea";
import { Alert, AlertDescription } from "#components/ui/alert";
import { cn } from "#lib/utils";
import { copyText } from "#lib/response";
import { getPolicyBridge } from "#lib/policy";
import { getRBACBridge, ALL_PERMISSIONS, canRole } from "#lib/rbac";
import type { Policy } from "#lib/policy";
import type { RBAC } from "#lib/rbac";

function PolicyYamlPreview({ policy }: { policy: Policy }) {
	const yaml = [
		`requireAudit: ${policy.requireAudit}`,
		`requireAuth: ${policy.requireAuth ?? false}`,
		`allowCustomThemes: ${policy.allowCustomThemes ?? true}`,
		`maxWorkflowSteps: ${policy.maxWorkflowSteps}`,
		`allowedActions: ${policy.allowedActions?.length ? `\n  - ${policy.allowedActions.join("\n  - ")}` : "[]"}`,
	].join("\n");
	return (
		<pre className="overflow-auto whitespace-pre-wrap rounded border bg-muted/40 p-2.5 font-mono text-[11px] leading-relaxed">{yaml}</pre>
	);
}

export function PolicyRbacView() {
	const [policy, setPolicy] = useState<Policy | null>(null);
	const [policyRaw, setPolicyRaw] = useState("");
	const [policyError, setPolicyError] = useState<string | null>(null);
	const [policyOk, setPolicyOk] = useState<string | null>(null);
	const [policyValid, setPolicyValid] = useState<boolean | null>(null);
	const [enforceAction, setEnforceAction] = useState("request.send");
	const [enforceResource, setEnforceResource] = useState("collections/api");
	const [enforceResult, setEnforceResult] = useState<string | null>(null);
	const [enforceError, setEnforceError] = useState<string | null>(null);

	const [rbac, setRbac] = useState<RBAC | null>(null);
	const [rbacError, setRbacError] = useState<string | null>(null);
	const [rbacCheckUser, setRbacCheckUser] = useState("alice");
	const [rbacCheckAction, setRbacCheckAction] = useState("workflow.run");
	const [rbacCheckResource, setRbacCheckResource] = useState("collections/api");
	const [rbacCheckResult, setRbacCheckResult] = useState<string | null>(null);
	const [rbacCheckError, setRbacCheckError] = useState<string | null>(null);

	const load = async () => {
		setPolicyError(null);
		setRbacError(null);
		try {
			const p = await getPolicyBridge().get();
			setPolicy(p);
			setPolicyRaw(`requireAudit: ${p.requireAudit}\nrequireAuth: ${p.requireAuth ?? false}\nallowCustomThemes: ${p.allowCustomThemes ?? true}\nmaxWorkflowSteps: ${p.maxWorkflowSteps}\nallowedActions: ${p.allowedActions?.length ? p.allowedActions.join(", ") : ""}`);
			setPolicyValid(null);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				const mock: Policy = { requireAudit: false, maxWorkflowSteps: 0, allowedActions: [], requireAuth: false, allowCustomThemes: true };
				setPolicy(mock);
				setPolicyRaw(`requireAudit: false\nrequireAuth: false\nallowCustomThemes: true\nmaxWorkflowSteps: 0\nallowedActions: `);
			} else setPolicyError(msg);
		}
		try {
			const r = await getRBACBridge().get();
			setRbac(r);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setRbac({
					roles: {
						admin: { name: "admin", permissions: ["*"] },
						editor: { name: "editor", permissions: ["request.send", "workflow.run", "collection.run"] },
						viewer: { name: "viewer", permissions: ["request.send"] },
					},
					userRoles: { alice: "editor", bob: "viewer" },
				});
			} else setRbacError(msg);
		}
	};

	useEffect(() => {
		void load();
	}, []);

	const savePolicy = async () => {
		if (!policy) return;
		setPolicyError(null);
		setPolicyOk(null);
		// Parse raw textarea as simple key: value lines for demo
		const next: Policy = { ...policy };
		for (const line of policyRaw.split("\n")) {
			const [k, ...rest] = line.split(":");
			if (!k) continue;
			const v = rest.join(":").trim();
			if (k.trim() === "requireAudit") next.requireAudit = v === "true";
			if (k.trim() === "requireAuth") next.requireAuth = v === "true";
			if (k.trim() === "allowCustomThemes") next.allowCustomThemes = v === "true";
			if (k.trim() === "maxWorkflowSteps") next.maxWorkflowSteps = Number(v) || 0;
			if (k.trim() === "allowedActions") next.allowedActions = v ? v.split(",").map((s) => s.trim()).filter(Boolean) : [];
		}
		try {
			await getPolicyBridge().save(next);
			setPolicy(next);
			setPolicyOk("policy saved to .reqly/policy.yaml (0600)");
			setPolicyValid(true);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setPolicy(next);
				setPolicyOk("policy saved (mock — no workspace in dev)");
				setPolicyValid(true);
			} else {
				setPolicyError(msg);
				setPolicyValid(false);
			}
		}
	};

	const runEnforce = async () => {
		setEnforceResult(null);
		setEnforceError(null);
		try {
			await getPolicyBridge().enforce(enforceAction.trim(), enforceResource.trim());
			setEnforceResult(`allowed — action ${JSON.stringify(enforceAction.trim())} on ${JSON.stringify(enforceResource.trim())}`);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				// Mock: allow all except when allowedActions is set and doesn't include action
				if (policy?.allowedActions?.length && !policy.allowedActions.includes(enforceAction.trim()) && !policy.allowedActions.includes("*")) {
					setEnforceError(`denied — action ${JSON.stringify(enforceAction.trim())} not in allowedActions`);
				} else setEnforceResult("allowed (mock)");
			} else setEnforceError(msg);
		}
	};

	const runRbacCheck = async () => {
		setRbacCheckResult(null);
		setRbacCheckError(null);
		try {
			await getRBACBridge().check(rbacCheckUser.trim(), rbacCheckAction.trim(), rbacCheckResource.trim());
			setRbacCheckResult(`allowed — ${rbacCheckUser.trim()} can ${rbacCheckAction.trim()} on ${rbacCheckResource.trim()}`);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				if (!rbac) return;
				const roleName = rbac.userRoles[rbacCheckUser.trim()] ?? "viewer";
				const role = rbac.roles[roleName];
				if (role && canRole(role, rbacCheckAction.trim())) setRbacCheckResult(`allowed (mock) — ${roleName} can ${rbacCheckAction.trim()}`);
				else setRbacCheckError(`denied (mock) — ${rbacCheckUser.trim()} (${roleName}) cannot ${rbacCheckAction.trim()}`);
			} else setRbacCheckError(msg);
		}
	};

	return (
		<section className="flex min-h-0 flex-col gap-4 p-4" aria-label="Policy and RBAC">
			<PageHeader icon={Shield} title="Policy & RBAC" description="Local-only governance — YAML policy (0600) and role matrix, Git-native, zero telemetry." />

			<div className="grid gap-4 lg:grid-cols-2">
				{/* Policy */}
				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<FileCheck className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">Policy</h3>
						<span className="font-mono text-[11px] text-muted-foreground">.reqly/policy.yaml</span>
						<span className="ml-auto inline-flex items-center gap-1">
							{policyValid === true && <Badge variant="outline" className="gap-1 border-status-ok/20 bg-status-ok/10 text-status-ok"><ShieldCheck className="size-3" /> valid</Badge>}
							{policyValid === false && <Badge variant="destructive" className="gap-1"><ShieldAlert className="size-3" /> invalid</Badge>}
							<Button variant="ghost" size="xs" onClick={() => void copyText(policyRaw)} aria-label="Copy policy"><Copy className="size-3.5" /></Button>
						</span>
					</div>

					<div className="flex flex-col gap-2 px-3 pb-3">
						{policyError && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{policyError}</AlertDescription></Alert>}
						{policyOk && <p className="font-mono text-xs text-status-ok">✓ {policyOk}</p>}

						<label className="flex flex-col gap-1">
							<span className="font-mono text-[11px] font-medium tracking-wide text-muted-foreground">Policy (YAML)</span>
							<Textarea value={policyRaw} onChange={(e) => setPolicyRaw(e.target.value)} rows={7} spellCheck={false} className="font-mono text-xs" placeholder="requireAudit: false&#10;maxWorkflowSteps: 0&#10;allowedActions: " />
						</label>

						{policy && <PolicyYamlPreview policy={policy} />}

						<div className="flex gap-1.5">
							<Button size="sm" onClick={() => void savePolicy()}>Save (0600)</Button>
							<Button variant="outline" size="sm" onClick={() => void load()}>Reload</Button>
							<span className="ml-auto font-mono text-[10px] text-muted-foreground self-center">FileStore 0600 · gitignored .reqly/</span>
						</div>

						<div className="rounded border border-border/60 bg-muted/20 p-2.5">
							<p className="font-mono text-[11px] font-semibold tracking-wide text-muted-foreground">Enforce dry-run</p>
							<div className="mt-2 grid grid-cols-2 gap-2">
								<Input value={enforceAction} onChange={(e) => setEnforceAction(e.target.value)} placeholder="action — e.g. workflow.run" className="font-mono text-xs" />
								<Input value={enforceResource} onChange={(e) => setEnforceResource(e.target.value)} placeholder="resource — e.g. collections/api" className="font-mono text-xs" />
							</div>
							<div className="mt-2 flex items-center gap-2">
								<Button size="sm" variant="outline" onClick={() => void runEnforce()} className="gap-1.5"><Play className="size-3.5" /> Enforce</Button>
								{enforceResult && <span className="font-mono text-xs text-status-ok">✓ {enforceResult}</span>}
								{enforceError && <span className="font-mono text-xs text-status-error">{enforceError}</span>}
							</div>
						</div>
					</div>
				</div>

				{/* RBAC */}
				<div className="flex flex-col gap-3 rounded border border-border bg-card">
					<div className="flex items-center gap-2 border-b border-border px-3 py-2.5">
						<Users className="size-4 text-muted-foreground" aria-hidden />
						<h3 className="text-sm font-semibold">RBAC</h3>
						<span className="font-mono text-[11px] text-muted-foreground">.reqly/rbac.yaml</span>
						<Badge variant="outline" className="ml-auto font-mono text-[10px]">{rbac ? Object.keys(rbac.roles).length : 0} roles</Badge>
					</div>

					<div className="flex flex-col gap-3 px-3 pb-3">
						{rbacError && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{rbacError}</AlertDescription></Alert>}

						{rbac ? (
							<>
								<div className="overflow-auto rounded border border-border">
									<table className="w-full border-collapse text-xs">
										<thead>
											<tr className="border-b border-border bg-muted/30">
												<th className="px-2 py-1.5 text-left font-mono text-[11px] font-semibold text-muted-foreground">Role → Permission</th>
												{ALL_PERMISSIONS.slice(0, 5).map((perm) => (
													<th key={perm} className="px-2 py-1.5 text-center font-mono text-[10px] font-medium text-muted-foreground">{perm.split(".")[1] ?? perm}</th>
												))}
											</tr>
										</thead>
										<tbody>
											{Object.values(rbac.roles).map((role) => (
												<tr key={role.name} className="border-b border-border/60 last:border-0">
													<td className="px-2 py-1.5">
														<span className={cn("inline-flex items-center gap-1 rounded border px-1.5 py-0.5 font-mono text-[10px] font-medium", role.name === "admin" ? "bg-status-warn/10 text-status-warn border-status-warn/20" : role.name === "editor" ? "bg-status-redirect/10 text-status-redirect border-status-redirect/20" : "bg-muted text-muted-foreground")}>{role.name}</span>
													</td>
													{ALL_PERMISSIONS.slice(0, 5).map((perm) => (
														<td key={perm} className="px-2 py-1.5 text-center">
															<span className={cn("inline-block size-2 rounded-full", canRole(role, perm) || canRole(role, "*") ? "bg-status-ok" : "bg-border")} aria-hidden />
														</td>
													))}
												</tr>
											))}
										</tbody>
									</table>
									<p className="border-t border-border bg-muted/20 px-2 py-1 font-mono text-[10px] text-muted-foreground">● has permission · ○ denied — matrix encodes authority, not decoration</p>
								</div>

								<div className="rounded border border-border/60 bg-muted/20 p-2.5">
									<p className="font-mono text-[11px] font-semibold tracking-wide text-muted-foreground">User → Role</p>
									<div className="mt-1.5 flex flex-col gap-1">
										{Object.entries(rbac.userRoles).length === 0 ? (
											<p className="font-mono text-xs text-muted-foreground">No explicit assignments — unknown users default to viewer (if present).</p>
										) : (
											Object.entries(rbac.userRoles).map(([user, role]) => (
												<div key={user} className="flex items-center gap-2 font-mono text-xs">
													<span className="min-w-20 truncate font-medium">{user}</span>
													<span className="text-muted-foreground">→</span>
													<Badge variant="outline" className="font-mono text-[10px]">{role}</Badge>
												</div>
											))
										)}
									</div>
								</div>

								<div className="rounded border border-border/60 bg-muted/20 p-2.5">
									<p className="font-mono text-[11px] font-semibold tracking-wide text-muted-foreground">Check (dry-run)</p>
									<div className="mt-2 grid grid-cols-3 gap-2">
										<Input value={rbacCheckUser} onChange={(e) => setRbacCheckUser(e.target.value)} placeholder="user" className="font-mono text-xs" />
										<Input value={rbacCheckAction} onChange={(e) => setRbacCheckAction(e.target.value)} placeholder="action" className="font-mono text-xs" />
										<Input value={rbacCheckResource} onChange={(e) => setRbacCheckResource(e.target.value)} placeholder="resource" className="font-mono text-xs" />
									</div>
									<div className="mt-2 flex items-center gap-2">
										<Button size="sm" variant="outline" onClick={() => void runRbacCheck()} className="gap-1.5"><Play className="size-3.5" /> Check</Button>
										{rbacCheckResult && <span className="font-mono text-xs text-status-ok">✓ {rbacCheckResult}</span>}
										{rbacCheckError && <span className="font-mono text-xs text-status-error">{rbacCheckError}</span>}
									</div>
								</div>
							</>
						) : (
							<p className="font-mono text-xs text-muted-foreground">Loading RBAC…</p>
						)}
					</div>
				</div>
			</div>

			<p className="font-mono text-[10px] text-muted-foreground">Policy & RBAC are local files (0600) under `.reqly/`, Git-native when committed. Enforce is pure — no network, no telemetry.</p>
		</section>
	);
}
