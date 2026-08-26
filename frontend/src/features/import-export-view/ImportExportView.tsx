import { useRef, useState } from "react";
import { ViewShell } from "../../components/shell/ViewLayout";
import { ArrowLeft, ArrowRight, Download, FileUp } from "lucide-react";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { Textarea } from "#components/ui/textarea";
import { cn } from "#lib/utils";
import {
	formatLabel,
	IMPORT_FORMAT_OPTIONS,
} from "#lib/import";
import {
	EXPORT_FORMAT_OPTIONS,
	exportFormatLabel,
	type ExportFormat,
} from "#lib/export";
import { useExportStore } from "#stores/useExportStore";
import { useImportStore } from "#stores/useImportStore";
import { useWorkspaceStore } from "#stores/useWorkspaceStore";
import { ImportReportView } from "../import-dialog/ImportReportView";

const STAGE_STEPS = [
	{ id: "input", label: "1 Payload" },
	{ id: "preview", label: "2 Preview" },
	{ id: "results", label: "3 Report" },
] as const;

const FORMAT_HINTS: [string, string][] = [
	["cURL", "terminal command curl -X POST …"],
	["OpenAPI 3.x", ".yaml .yml .json — openapi: 3.0.x"],
	["HAR 1.2", '{"log":{"version":"1.2"}}'],
	["Postman v2.1", '{"info":{"schema":"…/v2.1.0"}}'],
	["Insomnia v4/v5", '{"_type":"export…"} | yaml doc'],
	["Bruno", "collection dir .bru — bruno collection items tree"],
];

function StageChip({
	label,
	state,
}: {
	label: string;
	state: "active" | "done" | "todo";
}) {
	return (
		<span
			className={cn(
				"rounded-full border px-3 py-0.5 font-data text-2xs uppercase tracking-widest",
				state === "active"
					? "border-primary/50 bg-primary/10 text-primary"
					: state === "done"
						? "border-status-ok/40 text-status-ok"
						: "border-border text-muted-foreground",
			)}
		>
			{label}
		</span>
	);
}

/** ImportCard is the full-page import flow: stepper, drop zone, paste area,
 * format hint, and preview/report stages — same store as the dialog. */
function ImportCard({ onImported }: { onImported?: () => void }) {
	const setOpen = useImportStore((s) => s.setOpen);
	const stage = useImportStore((s) => s.stage);
	const content = useImportStore((s) => s.content);
	const filename = useImportStore((s) => s.filename);
	const formatHint = useImportStore((s) => s.formatHint);
	const detected = useImportStore((s) => s.detected);
	const outcome = useImportStore((s) => s.outcome);
	const busy = useImportStore((s) => s.busy);
	const error = useImportStore((s) => s.error);
	const setContent = useImportStore((s) => s.setContent);
	const setFormatHint = useImportStore((s) => s.setFormatHint);
	const runPreview = useImportStore((s) => s.runPreview);
	const commit = useImportStore((s) => s.commit);
	const back = useImportStore((s) => s.back);
	const fileInputRef = useRef<HTMLInputElement>(null);
	const [dragging, setDragging] = useState(false);

	const stageIndex = STAGE_STEPS.findIndex((s) => s.id === stage);

	const readFile = async (file: File) => {
		setContent(await file.text(), file.name);
	};

	return (
		<div className="flex flex-col gap-3 rounded-xl border border-border bg-card p-4">
			<h2 className="flex items-center gap-2 text-lg font-medium text-foreground">
				<FileUp className="size-4" aria-hidden />
				Import
			</h2>
			<div className="flex items-center gap-1.5">
				{STAGE_STEPS.map((s, i) => (
					<StageChip
						key={s.id}
						label={s.label}
						state={i === stageIndex ? "active" : i < stageIndex ? "done" : "todo"}
					/>
				))}
			</div>

			{error ? (
				<Alert variant="destructive">
					<AlertDescription>{error}</AlertDescription>
				</Alert>
			) : null}

			{stage === "input" ? (
				<>
					<button
						type="button"
						onClick={() => fileInputRef.current?.click()}
						onDragOver={(e) => {
							e.preventDefault();
							setDragging(true);
						}}
						onDragLeave={() => setDragging(false)}
						onDrop={(e) => {
							e.preventDefault();
							setDragging(false);
							const file = e.dataTransfer.files[0];
							if (file) void readFile(file);
						}}
						className={cn(
							"flex h-48 flex-col items-center justify-center gap-1.5 rounded-xl border-2 border-dashed transition-colors",
							dragging ? "border-primary bg-primary/5" : "border-border hover:bg-accent/40",
						)}
					>
						<FileUp className="size-8 text-muted-foreground/50" aria-hidden />
						<p className="text-xs text-foreground">
							<span className="font-medium">Drop a file</span> or click to browse
						</p>
						<p className="text-xs text-muted-foreground">
							cURL command, OpenAPI 3.x, HAR 1.2, Postman v2.1, Insomnia v4/v5,
							Bruno collection
						</p>
					</button>
					<input
						ref={fileInputRef}
						type="file"
						accept=".json,.yaml,.yml,.har,.txt"
						className="hidden"
						onChange={(e) => {
							const file = e.target.files?.[0];
							if (file) void readFile(file);
						}}
					/>
					<Textarea
						value={content}
						onChange={(e) => setContent(e.target.value, filename)}
						rows={4}
						spellCheck={false}
						aria-label="Paste import payload"
						placeholder="…or paste a cURL command / JSON / YAML here"
						className="resize-y font-mono text-xs"
					/>
					<div className="flex flex-wrap items-center gap-2">
						{detected && !detected.ok ? (
							<Badge variant="ghost">unknown format — pick manually</Badge>
						) : detected ? (
							<Badge variant="secondary" className="text-status-ok">
								{formatLabel(detected.format)}
							</Badge>
						) : null}
						<select
							value={formatHint}
							onChange={(e) => setFormatHint(e.target.value)}
							aria-label="Import format"
							className="h-7 rounded-md border border-border bg-transparent px-2 text-xs"
						>
							{IMPORT_FORMAT_OPTIONS.map((o) => (
								<option key={o.value} value={o.value}>
									{o.label}
								</option>
							))}
						</select>
						<Button
							size="sm"
							variant="destructive"
							disabled={busy || content.trim() === ""}
							onClick={() => void runPreview()}
						>
							{busy ? <Spinner data-icon="inline-start" /> : <ArrowRight data-icon="inline-start" />}
							Continue
						</Button>
					</div>
					<dl className="flex flex-col gap-0.5 text-xs">
						{FORMAT_HINTS.map(([name, hint]) => (
							<div key={name} className="flex gap-2">
								<dt className="w-28 shrink-0 font-semibold text-foreground">{name}</dt>
								<dd className="truncate font-mono text-muted-foreground">{hint}</dd>
							</div>
						))}
					</dl>
				</>
			) : stage === "preview" ? (
				<>
					<p className="text-xs text-muted-foreground">
						{detected ? `Detected ${formatLabel(detected.format)}` : "Preview"} —
						confirm to write files into the workspace.
					</p>
					<div className="flex items-center gap-2">
						<Button variant="outline" size="sm" onClick={back}>
							<ArrowLeft data-icon="inline-start" />
							Back
						</Button>
						<Button
							size="sm"
							variant="destructive"
							disabled={busy}
							onClick={() => {
								void commit().then(() => onImported?.());
							}}
						>
							{busy ? <Spinner data-icon="inline-start" /> : <Download data-icon="inline-start" />}
							Import
						</Button>
					</div>
				</>
			) : (
				<div className="flex flex-col gap-2">
					{outcome?.report ? <ImportReportView report={outcome.report} /> : null}
					<Button variant="outline" size="sm" className="self-start" onClick={() => setOpen(false)}>
						<ArrowLeft data-icon="inline-start" />
						Start another import
					</Button>
				</div>
			)}
		</div>
	);
}

/** ExportCard is the WHAT/FORMAT/OPTIONS export form — same store as the
 * dialog; output lands in .reqly/exports. */
function ExportCard() {
	const format = useExportStore((s) => s.format);
	const collection = useExportStore((s) => s.collection);
	const outName = useExportStore((s) => s.outName);
	const outcome = useExportStore((s) => s.outcome);
	const busy = useExportStore((s) => s.busy);
	const error = useExportStore((s) => s.error);
	const setFormat = useExportStore((s) => s.setFormat);
	const setCollection = useExportStore((s) => s.setCollection);
	const setOutName = useExportStore((s) => s.setOutName);
	const run = useExportStore((s) => s.run);
	const tree = useWorkspaceStore((s) => s.workspaceTree);

	return (
		<div className="flex flex-col gap-3 rounded-xl border border-border bg-card p-4">
			<h2 className="flex items-center gap-2 text-sm font-semibold text-foreground">
				<Download className="size-4" aria-hidden />
				Export
			</h2>
			{error ? (
				<Alert variant="destructive">
					<AlertDescription>{error}</AlertDescription>
				</Alert>
			) : null}
			<div className="flex flex-col gap-1">
				<p className="font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
					What
				</p>
				<select
					value={collection}
					onChange={(e) => setCollection(e.target.value)}
					aria-label="Export scope"
					className="h-8 rounded-md border border-border bg-transparent px-2 text-xs"
				>
					<option value="">Whole workspace</option>
					{(tree?.collections ?? []).map((c) => (
						<option key={c.path} value={c.name}>
							{c.name}
						</option>
					))}
				</select>
			</div>
			<div className="flex flex-col gap-1">
				<p className="font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
					Format
				</p>
				<select
					value={format}
					onChange={(e) => {
						// SAFETY: options come from EXPORT_FORMAT_OPTIONS, all valid ExportFormats.
						setFormat(e.target.value as ExportFormat);
					}}
					aria-label="Export format"
					className="h-8 rounded-md border border-border bg-transparent px-2 text-xs"
				>
					{EXPORT_FORMAT_OPTIONS.map((o) => (
						<option key={o.value} value={o.value}>
							{o.label}
						</option>
					))}
				</select>
			</div>
			<div className="flex flex-col gap-1">
				<p className="font-data text-2xs font-medium uppercase tracking-widest text-muted-foreground">
					Output file name (optional)
				</p>
				<Input
					value={outName}
					onChange={(e) => setOutName(e.target.value)}
					placeholder="defaults to a timestamped name"
					spellCheck={false}
					className="font-mono text-xs"
					aria-label="Output file name"
				/>
			</div>
			<div className="flex items-center gap-2">
				<Button
					size="sm"
					variant="destructive"
					disabled={busy}
					onClick={() => void run()}
				>
					{busy ? <Spinner data-icon="inline-start" /> : <Download data-icon="inline-start" />}
					Export {exportFormatLabel(format)}
				</Button>
				{outcome ? (
					<span className="text-xs text-status-ok">
						Wrote {outcome.path} — files stay local.
					</span>
				) : null}
			</div>
		</div>
	);
}

/** ImportExportView is the G-17.4.10 full-page import/export surface
 * (reference 14-import-export.png), sharing stores with the quick dialogs. */
export function ImportExportView() {
	const refreshWorkspace = useWorkspaceStore((s) => s.refreshWorkspace);
	return (
		<ViewShell label="Import and export">
			<ImportCard onImported={() => void refreshWorkspace()} />
			<ExportCard />
		</ViewShell>
	);
}
