import { Monitor, Moon, Sun } from "lucide-react";

import { useThemeStore } from "#stores";

import { Button } from "./ui/button";

export function ThemeToggle() {
  const theme = useThemeStore((s) => s.theme);
  const cycleTheme = useThemeStore((s) => s.cycleTheme);

  const isLight = theme === "atlas-light";
  const isSystem = theme === "system";

  const label = isLight
    ? "Switch to dark mode"
    : isSystem
      ? "Switch to light mode"
      : "Switch to system mode";

  return (
    <Button variant="ghost" size="icon-sm" onClick={cycleTheme} aria-label={label} title={label}>
      {isLight ? <Moon /> : isSystem ? <Sun /> : <Monitor />}
    </Button>
  );
}
