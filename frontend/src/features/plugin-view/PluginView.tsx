import { useEffect, useState } from "react";
import { Puzzle, CheckCircle2, XCircle, RefreshCw } from "lucide-react";
import { PageHeader } from "#components/PageHeader";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Alert, AlertDescription } from "#components/ui/alert";
import { getPluginBridge } from "#lib/plugin";
import type { PluginView } from "#lib/plugin";

export function PluginView() {
	const [plugins, setPlugins] = useState<PluginView[]>([]);
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const load = async () => {
		setBusy(true);
		setError(null);
		try {
			const list = await getPluginBridge().list();
			setPlugins(list ?? []);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setPlugins([
					{ name: "example", version: "0.1.0", capabilities: ["transform", "auth"], valid: true, dir: "plugins/example" },
					{ name: "broken", version: "0.1.0", capabilities: [], valid: false, error: "missing manifest.json", dir: "plugins/broken" },
				]);
			} else setError(msg);
		} finally {
			setBusy(false);
		}
	};

	useEffect(() => {
		void load();
	}, []);

	const validate = async (name: string) => {
		setError(null);
		try {
			const v = await getPluginBridge().validate(name);
			setPlugins((prev) => prev.map((p) => (p.name === name ? v : p)));
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("not available in this build")) {
				setPlugins((prev) => prev.map((p) => (p.name === name ? { ...p, valid: !p.valid } : p)));
			} else setError(msg);
		}
	};

	return (
		<section className="flex min-h-0 flex-col gap-4 p-4" aria-label="Plugins">
			<PageHeader
				icon={Puzzle}
				title="Plugins"
				description="Local Goja plugins — manifest.json + plugin.js in plugins/<name>/, capabilities as badges."
				actions={
					<Button size="sm" variant="outline" onClick={() => void load()} disabled={busy} className="gap-1.5">
						<RefreshCw className="size-3.5" /> Reload
					</Button>
				}
			/>

			{error && <Alert variant="destructive" className="py-2"><AlertDescription className="font-mono text-xs">{error}</AlertDescription></Alert>}

			<div className="overflow-auto rounded border border-border bg-card">
				<table className="w-full border-collapse text-xs">
					<thead className="border-b border-border bg-muted/30">
						<tr className="text-left font-mono text-[11px] text-muted-foreground">
							<th className="px-3 py-1.5 font-semibold">Plugin</th>
							<th className="px-3 py-1.5 font-semibold">Version</th>
							<th className="px-3 py-1.5 font-semibold">Capabilities</th>
							<th className="px-3 py-1.5 font-semibold">Status</th>
							<th className="px-3 py-1.5 font-semibold">Dir</th>
							<th className="px-3 py-1.5 font-semibold text-right">Check</th>
						</tr>
					</thead>
					<tbody>
						{plugins.length === 0 ? (
							<tr>
								<td colSpan={6} className="px-3 py-8 text-center font-mono text-xs text-muted-foreground">No plugins found — create <code className="rounded bg-muted px-1">plugins/&lt;name&gt;/manifest.json</code> + <code className="rounded bg-muted px-1">plugin.js</code>.</td>
							</tr>
						) : (
							plugins.map((p) => (
								<tr key={p.name} className="border-b border-border/40 last:border-0">
									<td className="px-3 py-1.5 font-mono font-medium">{p.name}</td>
									<td className="px-3 py-1.5 font-mono text-muted-foreground">{p.version || "—"}</td>
									<td className="px-3 py-1.5">
										<div className="flex flex-wrap gap-1">
											{(p.capabilities ?? []).length === 0 ? (
												<span className="font-mono text-[11px] text-muted-foreground">—</span>
											) : (
												p.capabilities.map((c) => (
													<Badge key={c} variant="outline" className="font-mono text-[10px]">
														{c}
													</Badge>
												))
											)}
										</div>
									</td>
									<td className="px-3 py-1.5">
										{p.valid ? (
											<span className="inline-flex items-center gap-1 rounded border border-status-ok/20 bg-status-ok/10 px-1.5 py-0.5 font-mono text-[10px] text-status-ok">
												<CheckCircle2 className="size-3" /> valid
											</span>
										) : (
											<span className="inline-flex items-center gap-1 rounded border border-status-error/20 bg-status-error/10 px-1.5 py-0.5 font-mono text-[10px] text-status-error" title={p.error}>
												<XCircle className="size-3" /> invalid
											</span>
										)}
									</td>
									<td className="max-w-[20ch] truncate px-3 py-1.5 font-mono text-[11px] text-muted-foreground" title={p.dir}>
										{p.dir}
									</td>
									<td className="px-3 py-1.5 text-right">
										<Button size="xs" variant="ghost" onClick={() => void validate(p.name)} className="h-6 text-xs">
											Validate
										</Button>
									</td>
								</tr>
							))
						)}
					</tbody>
				</table>
			</div>

			<p className="font-mono text-[10px] text-muted-foreground">Plugins are local, Goja-compiled, capabilities advertised in manifest — no network, no store.</p>
		</section>
	);
}
