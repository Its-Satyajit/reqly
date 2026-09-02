export const BOTTOM_PANELS = ["console", "network", "tests", "variables", "cookies", "devtools"] as const;
export type BottomPanelId = (typeof BOTTOM_PANELS)[number];

const PANEL_SET: ReadonlySet<string> = new Set(BOTTOM_PANELS);
export function isBottomPanelId(v: string): v is BottomPanelId {
  // SAFETY: PANEL_SET is constructed from BOTTOM_PANELS and checks string membership directly.
  return PANEL_SET.has(v);
}

export function nextPanel(current: BottomPanelId | null): BottomPanelId | null {
  if (current === null) return "console";
  const idx = BOTTOM_PANELS.indexOf(current);
  if (idx === BOTTOM_PANELS.length - 1) return null;
  return BOTTOM_PANELS[idx + 1];
}
