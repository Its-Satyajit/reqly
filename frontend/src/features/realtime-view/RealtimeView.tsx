import { Radio } from "lucide-react";
import { useRealtimeStore } from "#stores/useRealtimeStore";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { ViewShell } from "../../components/shell/ViewLayout";

/** RealtimeView is the G-17.4.15 full-page home for WebSocket/SSE client
 * tabs, replacing the sidebar REALTIME strip (user directive). Opening a
 * kind lands you in the requests view with a live tab. */
export function RealtimeView() {
	const openTab = useWorkspaceStore((s) => s.openTab);
	const setActiveView = useWorkspaceStore((s) => s.setActiveView);
	const newTab = useRealtimeStore((s) => s.newTab);

	const openRealtimeTab = (kind: "ws" | "sse") => {
		const id = `realtime-${kind}-${Date.now()}`;
		newTab(id, kind);
		openTab({ id, title: kind === "ws" ? "WebSocket" : "SSE", kind: "realtime" });
		setActiveView("requests");
	};

	return (
		<ViewShell label="Realtime" className="overflow-y-auto">
			<h2 className="text-lg font-medium">Realtime</h2>
			<div className="grid max-w-xl grid-cols-2 gap-3">
				<button
					type="button"
					onClick={() => openRealtimeTab("ws")}
					className="flex flex-col items-center gap-2 rounded-xl border border-border bg-card px-4 py-6 transition-colors hover:bg-accent"
				>
					<span className="font-data text-sm font-semibold text-foreground">WS</span>
					<span className="text-xs text-muted-foreground">
						Open a WebSocket client tab — send text/binary frames, watch the log.
					</span>
				</button>
				<button
					type="button"
					onClick={() => openRealtimeTab("sse")}
					className="flex flex-col items-center gap-2 rounded-xl border border-border bg-card px-4 py-6 transition-colors hover:bg-accent"
				>
					<span className="font-data text-sm font-semibold text-foreground">SSE</span>
					<span className="text-xs text-muted-foreground">
						Open a server-sent-events tab — filter, pause, and inspect the stream.
					</span>
				</button>
			</div>
			<p className="flex max-w-xl items-center gap-1.5 text-xs text-muted-foreground/70">
				<Radio className="size-3" aria-hidden />
				Connections run through the desktop bridge; frames stay local.
			</p>
		</ViewShell>
	);
}
