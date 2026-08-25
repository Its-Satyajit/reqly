import { useEffect, useRef } from "react";
import type { Extension } from "@codemirror/state";

export type EditorLanguage =
	| "json"
	| "javascript"
	| "xml"
	| "yaml"
	| "markdown"
	| "graphql"
	| "text";

interface CodeMirrorModules {
	EditorState: typeof import("@codemirror/state").EditorState;
	EditorView: typeof import("@codemirror/view").EditorView;
	oneDark: typeof import("@codemirror/theme-one-dark").oneDark;
	basicSetup: typeof import("codemirror").basicSetup;
	languageExtensions: Record<EditorLanguage, Extension>;
}

// The CodeMirror stack (~400KB) is loaded lazily on first editor mount so the
// app shell paints without it.
let cmPromise: Promise<CodeMirrorModules> | null = null;

function loadCodeMirror(): Promise<CodeMirrorModules> {
	cmPromise ??= Promise.all([
		import("@codemirror/lang-javascript"),
		import("@codemirror/lang-json"),
		import("@codemirror/lang-markdown"),
		import("@codemirror/lang-xml"),
		import("@codemirror/lang-yaml"),
		import("@codemirror/state"),
		import("@codemirror/theme-one-dark"),
		import("@codemirror/view"),
		import("codemirror"),
	]).then(
		([
			{ javascript },
			{ json },
			{ markdown },
			{ xml },
			{ yaml },
			{ EditorState },
			{ oneDark },
			{ EditorView },
			{ basicSetup },
		]) => ({
			EditorState,
			EditorView,
			oneDark,
			basicSetup,
			languageExtensions: {
				json: json(),
				javascript: javascript(),
				xml: xml(),
				yaml: yaml(),
				markdown: markdown(),
				graphql: [],
				text: [],
			},
		}),
	);
	return cmPromise;
}

export interface CodeMirrorEditorProps {
	value?: string;
	language?: EditorLanguage;
	theme?: "light" | "dark";
	readOnly?: boolean;
	onChange?: (value: string) => void;
	className?: string;
}

export function CodeMirrorEditor({
	value = "",
	language = "text",
	theme = "dark",
	readOnly = false,
	onChange,
	className,
}: CodeMirrorEditorProps) {
	const containerRef = useRef<HTMLDivElement>(null);
	// SAFETY: viewRef holds the live EditorView instance created in the mount effect below.
	const viewRef = useRef<import("@codemirror/view").EditorView | null>(null);

	useEffect(() => {
		let disposed = false;
		void loadCodeMirror().then((cm) => {
			if (disposed || !containerRef.current) return;

			const state = cm.EditorState.create({
				doc: value,
				extensions: [
					cm.basicSetup,
					cm.languageExtensions[language],
					readOnly ? cm.EditorState.readOnly.of(true) : [],
					theme === "dark" ? cm.oneDark : [],
					cm.EditorView.updateListener.of((update) => {
						if (update.docChanged) onChange?.(update.state.doc.toString());
					}),
				],
			});
			const view = new cm.EditorView({
				state,
				parent: containerRef.current,
			});
			viewRef.current = view;
		});

		return () => {
			disposed = true;
			viewRef.current?.destroy();
			viewRef.current = null;
		};
		// Intentionally ignore `value` and `onChange`: the editor is built once.
		// External value changes are applied by the effect below.
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [language, theme, readOnly]);

	useEffect(() => {
		const view = viewRef.current;
		if (!view) return;
		const current = view.state.doc.toString();
		if (current !== value) {
			view.dispatch({
				changes: { from: 0, to: current.length, insert: value },
			});
		}
	}, [value]);

	return <div ref={containerRef} className={className} />;
}
