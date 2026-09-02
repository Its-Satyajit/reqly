export interface PluginView {
	name: string;
	version: string;
	capabilities: string[];
	valid: boolean;
	error?: string;
	dir: string;
}

export interface PluginAdapter {
	list(): Promise<PluginView[]>;
	validate(name: string): Promise<PluginView>;
}

export const fallbackPluginAdapter: PluginAdapter = {
	async list() {
		throw new Error("plugin is not available in this build");
	},
	async validate() {
		throw new Error("plugin is not available in this build");
	},
};

let pluginBridge: PluginAdapter | null = null;
export function setPluginBridge(a: PluginAdapter): void {
	pluginBridge = a;
}
export function getPluginBridge(): PluginAdapter {
	return pluginBridge ?? fallbackPluginAdapter;
}
