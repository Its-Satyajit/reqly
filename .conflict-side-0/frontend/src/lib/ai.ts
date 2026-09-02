export interface AIAdapter {
	explain(responseJson: string): Promise<string>;
	generateTests(responseJson: string): Promise<string>;
	generateDocs(requestJson: string, responseJson: string): Promise<string>;
	diagnose(responseJson: string, errMsg: string): Promise<string>;
	explainSchema(schemaJson: string): Promise<string>;
}

export const fallbackAIAdapter: AIAdapter = {
	async explain() {
		throw new Error("ai is not available in this build");
	},
	async generateTests() {
		throw new Error("ai is not available in this build");
	},
	async generateDocs() {
		throw new Error("ai is not available in this build");
	},
	async diagnose() {
		throw new Error("ai is not available in this build");
	},
	async explainSchema() {
		throw new Error("ai is not available in this build");
	},
};

let aiBridge: AIAdapter | null = null;
export function setAIBridge(a: AIAdapter): void {
	aiBridge = a;
}
export function getAIBridge(): AIAdapter {
	return aiBridge ?? fallbackAIAdapter;
}
