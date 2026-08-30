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

export function clampZoom(z: number): number {
  return Math.min(2, Math.max(0.5, z));
}

export function getNodeAt(nodes: SchemaNode[], x: number, y: number): SchemaNode | undefined {
  // Node rect: width 88, height 36, anchored at x-44,y
  for (const n of nodes) {
    if (x >= n.x - 44 && x <= n.x + 44 && y >= n.y && y <= n.y + 36) return n;
  }
  return undefined;
}

export function getTransform(zoom: number, panX: number, panY: number): string {
  const z = clampZoom(zoom);
  return `translate(${panX} ${panY}) scale(${z})`;
}
