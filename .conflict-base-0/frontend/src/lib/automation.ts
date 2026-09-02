export interface AutomationReportView {
	workflowName: string;
	passed: boolean;
	duration: string;
	steps: { name: string; passed: boolean; requestError?: string }[];
	extractedVars: Record<string, string>;
}

export interface AutomationAdapter {
	run(yaml: string): Promise<AutomationReportView>;
}

export const fallbackAutomationAdapter: AutomationAdapter = {
	async run() {
		throw new Error("automation is not available in this build");
	},
};

let automationBridge: AutomationAdapter | null = null;
export function setAutomationBridge(a: AutomationAdapter): void {
	automationBridge = a;
}
export function getAutomationBridge(): AutomationAdapter {
	return automationBridge ?? fallbackAutomationAdapter;
}
