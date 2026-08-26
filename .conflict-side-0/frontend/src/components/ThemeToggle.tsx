import { Moon, Sun } from 'lucide-react'

import { useThemeStore } from '#stores'

import { Button } from './ui/button'

export function ThemeToggle() {
  const preference = useThemeStore((s) => s.preference)
  const appearance = useThemeStore((s) => s.appearance)
  const toggleTheme = useThemeStore((s) => s.toggleTheme)

  const label =
    appearance === 'dark'
      ? `Switch to light mode (currently ${preference})`
      : `Switch to dark mode (currently ${preference})`

  return (
    <Button variant="ghost" size="icon-sm" onClick={toggleTheme} aria-label={label} title={label}>
      {appearance === 'dark' ? <Sun /> : <Moon />}
    </Button>
  )
}
