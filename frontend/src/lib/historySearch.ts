import Fuse from "fuse.js";
import type { HistoryEntry } from "./history";

const FUSE_OPTIONS = {
	keys: [
		{ name: "url", weight: 0.5 },
		{ name: "requestPath", weight: 0.4 },
		{ name: "method", weight: 0.1 },
	],
	threshold: 0.35,
	ignoreLocation: true,
	minMatchCharLength: 2,
};

/** searchHistory fuzzy-matches user text over recent history entries with
 * Fuse.js, so punctuation (`todo/`, `list-todos`) and typos behave like
 * users expect instead of like query syntax. */
export function searchHistory(entries: HistoryEntry[], query: string): HistoryEntry[] {
	const fuse = new Fuse(entries, FUSE_OPTIONS);
	return fuse.search(query).map((hit) => hit.item);
}
