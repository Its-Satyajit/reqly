import Fuse, { type IFuseOptions } from "fuse.js";

/** fuseSearch runs a fuzzy query over items with the given options and
 * returns matched items in relevance order. Pure — usable outside React. */
export function fuseSearch<T>(items: T[], query: string, options: IFuseOptions<T>): T[] {
	if (query.trim() === "") return [];
	return new Fuse(items, options)
		.search(query)
		.map((hit) => hit.item);
}
