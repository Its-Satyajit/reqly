export type SchemaNode = { id: string; label: string; x: number; y: number };
export type SchemaEdge = { from: string; to: string };

export const SCHEMA_NODES: SchemaNode[] = [
  { id: "User", label: "User", x: 200, y: 40 },
  { id: "Address", label: "Address", x: 60, y: 140 },
  { id: "Order", label: "Order", x: 200, y: 140 },
  { id: "Profile", label: "Profile", x: 340, y: 140 },
  { id: "Product", label: "Product", x: 200, y: 240 },
];

export const SCHEMA_EDGES: SchemaEdge[] = [
  { from: "User", to: "Address" },
  { from: "User", to: "Order" },
  { from: "User", to: "Profile" },
  { from: "Order", to: "Product" },
];

export function nodesForSpec(selectedId: string): SchemaNode[] {
  if (selectedId.startsWith("components")) return SCHEMA_NODES;
  return SCHEMA_NODES.slice(0, 1);
}

export function edgesForNodes(nodes: SchemaNode[]): SchemaEdge[] {
  const ids = new Set(nodes.map((n) => n.id));
  return SCHEMA_EDGES.filter((e) => ids.has(e.from) && ids.has(e.to));
}

export function nodesForOpenApiContent(content: string): SchemaNode[] {
  if (!content || !content.includes("components:")) return SCHEMA_NODES.slice(0, 1);
  // M53 T1: reuse static nodes for now; full YAML ref parsing deferred to follow-up
  return SCHEMA_NODES;
}

export const ZOOM_MIN = 0.5;
export const ZOOM_MAX = 2;
export const NODE_WIDTH = 88;
export const NODE_HEIGHT = 36;
export const NODE_HALF_WIDTH = NODE_WIDTH / 2;

export function clampZoom(z: number): number {
  return Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, z));
}

export function getNodeAt(nodes: SchemaNode[], x: number, y: number): SchemaNode | undefined {
  for (const n of nodes) {
    if (x >= n.x - NODE_HALF_WIDTH && x <= n.x + NODE_HALF_WIDTH && y >= n.y && y <= n.y + NODE_HEIGHT) return n;
  }
  return undefined;
}

export function getTransform(zoom: number, panX: number, panY: number): string {
  const z = clampZoom(zoom);
  return `translate(${panX} ${panY}) scale(${z})`;
}
