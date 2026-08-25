import type { IFuseOptions } from "fuse.js";
import { fuseSearch } from "./fuseSearch";

/** useFuseSearch fuzzy-matches query over items with Fuse.js. Returns [] for
 * blank queries. React Compiler memoizes the computation automatically. */
export function useFuseSearch<T>(
	items: T[],
	query: string,
	options: IFuseOptions<T>,
): T[] {
	return fuseSearch(items, query, options);
}
