export const BOTTOM_PANELS = ["console", "network", "tests", "variables", "cookies"] as const;
export type BottomPanelId = (typeof BOTTOM_PANELS)[number];

export function isBottomPanelId(v: string): v is BottomPanelId {
  return (BOTTOM_PANELS as readonly string[]).includes(v);
}

export function nextPanel(current: BottomPanelId | null): BottomPanelId | null {
  if (current === null) return "console";
  const idx = BOTTOM_PANELS.indexOf(current);
  if (idx === BOTTOM_PANELS.length - 1) return null;
  return BOTTOM_PANELS[idx + 1];
}
