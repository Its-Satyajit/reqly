import type { IFuseOptions } from "fuse.js";
import type { HistoryEntry } from "./history";

export const HISTORY_FUSE_OPTIONS: IFuseOptions<HistoryEntry> = {
	keys: [
		{ name: "url", weight: 0.5 },
		{ name: "requestPath", weight: 0.4 },
		{ name: "method", weight: 0.1 },
	],
	threshold: 0.35,
	ignoreLocation: true,
	minMatchCharLength: 2,
};

