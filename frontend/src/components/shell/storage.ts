/** Shared localStorage adapter for react-resizable-panels default layouts. */
export const shellStorage: Pick<Storage, "getItem" | "setItem"> = {
	getItem: (key) => window.localStorage.getItem(key),
	setItem: (key, value) => window.localStorage.setItem(key, value),
};
