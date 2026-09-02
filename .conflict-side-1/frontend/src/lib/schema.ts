export type Violation = {
	path: string;
	message: string;
};

export interface SchemaValidateResult {
	valid: boolean;
	violations: Violation[];
}

export interface SchemaAdapter {
	validate(schemaJson: string, instanceJson: string, draft: string): Promise<SchemaValidateResult>;
	inspect(schemaJson: string): Promise<string>;
	generate(schemaJson: string, seed: number): Promise<string>;
}

export const fallbackSchemaAdapter: SchemaAdapter = {
	async validate() {
		throw new Error("schema is not available in this build");
	},
	async inspect() {
		throw new Error("schema is not available in this build");
	},
	async generate() {
		throw new Error("schema is not available in this build");
	},
};

let schemaBridge: SchemaAdapter | null = null;
export function setSchemaBridge(a: SchemaAdapter): void {
	schemaBridge = a;
}
export function getSchemaBridge(): SchemaAdapter {
	return schemaBridge ?? fallbackSchemaAdapter;
}
