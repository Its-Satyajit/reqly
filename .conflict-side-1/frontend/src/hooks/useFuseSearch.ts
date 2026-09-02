import { useMemo } from "react";
import type { IFuseOptions } from "fuse.js";
import { fuseSearch } from "./fuseSearch";

/** useFuseSearch fuzzy-matches query over items with Fuse.js, recomputing
 * only when inputs change. Returns [] for blank queries. */
export function useFuseSearch<T>(
	items: T[],
	query: string,
	options: IFuseOptions<T>,
): T[] {
	return useMemo(
		() => fuseSearch(items, query, options),
		[items, query, options],
	);
}
