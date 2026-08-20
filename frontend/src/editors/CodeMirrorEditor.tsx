import { javascript } from "@codemirror/lang-javascript";
import { json } from "@codemirror/lang-json";
import { markdown } from "@codemirror/lang-markdown";
import { xml } from "@codemirror/lang-xml";
import { yaml } from "@codemirror/lang-yaml";
import { EditorState, type Extension } from "@codemirror/state";
import { oneDark } from "@codemirror/theme-one-dark";
import { EditorView } from "@codemirror/view";
import { basicSetup } from "codemirror";
import { useEffect, useRef } from "react";

export type EditorLanguage =
	| "json"
	| "javascript"
	| "xml"
	| "yaml"
	| "markdown"
	| "text";

const languageExtensions = {
	json: json(),
	javascript: javascript(),
	xml: xml(),
	yaml: yaml(),
	markdown: markdown(),
	text: [],
} satisfies Record<EditorLanguage, Extension>;

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
	const viewRef = useRef<EditorView | null>(null);

	useEffect(() => {
		if (!containerRef.current) return;

		const state = EditorState.create({
			doc: value,
			extensions: [
				basicSetup,
				languageExtensions[language],
				readOnly ? EditorState.readOnly.of(true) : [],
				theme === "dark" ? oneDark : [],
				EditorView.updateListener.of((update) => {
					if (update.docChanged) onChange?.(update.state.doc.toString());
				}),
			],
		});
		const view = new EditorView({ state, parent: containerRef.current });
		viewRef.current = view;

		return () => {
			view.destroy();
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
