/** Canonical workspace view registry — single source for navigation surfaces
 * (sidebar, statusbar, command palette). Ids mirror `WorkspaceView`. */
export const WORKSPACE_VIEWS = [
	{ id: "overview", label: "Overview" },
	{ id: "requests", label: "Requests" },
	{ id: "tests", label: "Tests" },
	{ id: "realtime", label: "Realtime" },
	{ id: "oauth", label: "OAuth tokens" },
	{ id: "git", label: "Git" },
	{ id: "environments", label: "Environments" },
	{ id: "history", label: "History" },
	{ id: "mocks", label: "Mock servers" },
	{ id: "diff", label: "API diff" },
	{ id: "jwt", label: "JWT inspector" },
	{ id: "graphql", label: "GraphQL browser" },
	{ id: "runners", label: "Runners" },
	{ id: "explorer", label: "OpenAPI explorer" },
	{ id: "docs", label: "Docs generator" },
	{ id: "grpc", label: "gRPC client" },
	{ id: "settings", label: "Settings" },
] as const;

export type WorkspaceViewId = (typeof WORKSPACE_VIEWS)[number]["id"];

export function workspaceViewLabel(id: string): string {
	return WORKSPACE_VIEWS.find((v) => v.id === id)?.label ?? id;
}
