export type SpecNode = { id: string; label: string; children?: SpecNode[] };

export const SPEC_SECTIONS: SpecNode[] = [
  { id: "info", label: "Info" },
  { id: "servers", label: "Servers" },
  {
    id: "paths",
    label: "Paths",
    children: [
      { id: "paths:/users", label: "/users" },
      { id: "paths:/products", label: "/products" },
      { id: "paths:/orders", label: "/orders" },
    ],
  },
  {
    id: "components",
    label: "Components",
    children: [
      { id: "components:schemas", label: "Schemas" },
      { id: "components:security", label: "Security" },
    ],
  },
];

export function flattenSpecTree(nodes: SpecNode[]): SpecNode[] {
  const out: SpecNode[] = [];
  const walk = (ns: SpecNode[]) => {
    for (const n of ns) {
      out.push(n);
      if (n.children) walk(n.children);
    }
  };
  walk(nodes);
  return out;
}
