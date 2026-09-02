export interface WorkflowReportView {
	workflowName: string;
	passed: boolean;
	duration: string;
	steps: { name: string; passed: boolean; requestError?: string }[];
	extractedVars: Record<string, string>;
}

export interface WorkflowAdapter {
	run(yaml: string): Promise<WorkflowReportView>;
}

export const fallbackWorkflowAdapter: WorkflowAdapter = {
	async run() {
		throw new Error("workflow is not available in this build");
	},
};

let workflowBridge: WorkflowAdapter | null = null;
export function setWorkflowBridge(a: WorkflowAdapter): void {
	workflowBridge = a;
}
export function getWorkflowBridge(): WorkflowAdapter {
	return workflowBridge ?? fallbackWorkflowAdapter;
}
