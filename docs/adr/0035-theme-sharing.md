# ADR 0035: Theme Sharing — Git-Native Shareable Themes (M67)

## Status
Accepted — grill Q1 (core + CLI + desktop, UI picker deferred)

## Context
ROADMAP P3 §58.2 Theme Marketplace requires custom themes that are Git-native, validated, and shareable without cloud. Tokens live in `frontend/src/index.css` (`--primary` etc.), currently hardcoded to atlas-light/dark.

## Decision
Ship `internal/theme` (M67) + `frontend/src/lib/themes.ts` extension: `Theme{id, label, appearance, tokens}` + `Validate` (kebab-case id, non-empty label, light/dark, token hex/hsl) + `Parse` (YAML via gopkg.in/yaml.v3, JSON via encoding/json) + `MarshalYAML/JSON` + `ToCSS` + `BuiltInThemes`/`IsBuiltIn`; CLI `reqly theme list/export/import` + desktop `AppService.ThemeList/Export/Import` + frontend `CustomTheme`, `validateCustomTheme`, `parseCustomTheme`, `customThemeToCSS` (jsdom, SAFETY comments, YamlResult).

## Consequences
Q1: Tokens are hex/hsl only; Tailwind `bg-primary` etc. remain via CSS vars, no runtime Tailwind generation.
Q2: UI picker is deferred — themes are import/export only, no live preview.
