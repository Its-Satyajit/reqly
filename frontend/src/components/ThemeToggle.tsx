import { Monitor, Moon, Sun } from "lucide-react";

import { useThemeStore } from "#stores";

import { Button } from "./ui/button";

export function ThemeToggle() {
  const theme = useThemeStore((s) => s.theme);
  const cycleTheme = useThemeStore((s) => s.cycleTheme);

  const label =
    theme === "atlas-light"
      ? "Switch to dark mode"
      : theme === "atlas-dark"
        ? "Switch to system mode"
        : "Switch to light mode";

  return (
    <Button variant="ghost" size="icon-sm" onClick={cycleTheme} aria-label={label} title={label}>
      {theme === "atlas-light" ? (
        <Moon />
      ) : theme === "atlas-dark" ? (
        <Monitor />
      ) : (
        <Sun />
      )}
    </Button>
  );
}
