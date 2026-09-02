export interface ChangelogItem {
	type: string;
	path: string;
	summary: string;
	severity: string;
}

export interface Changelog {
	suggested_semver: string;
	breaking: ChangelogItem[];
	additions: ChangelogItem[];
	info: ChangelogItem[];
}

export interface ChangelogResultView {
	changelog: Changelog;
	markdown: string;
	json: string;
}

export interface ChangelogAdapter {
	generate(oldPath: string, newPath: string, format: string, failOnBreaking: boolean): Promise<ChangelogResultView>;
}

export const fallbackChangelogAdapter: ChangelogAdapter = {
	async generate() {
		throw new Error("changelog is not available in this build");
	},
};

let changelogBridge: ChangelogAdapter | null = null;
export function setChangelogBridge(a: ChangelogAdapter): void {
	changelogBridge = a;
}
export function getChangelogBridge(): ChangelogAdapter {
	return changelogBridge ?? fallbackChangelogAdapter;
}
