// oxlint configuration for the Reqly monorepo.
//
// oxc's linter (oxlint) runs over the TypeScript/React sources in
// `frontend/` and `apps/desktop/frontend/`. The `anti-slop` plugin is
// vendored (not a fixed dependency) at `tools/oxlint/anti-slop/` per its
// upstream guidance — copy from https://github.com/dmmulroy/anti-slop, keep
// it in sync with `oxlint`/`@oxlint/plugins`, and adjust the rule set to
// this repo's standards.
//
// Ignore patterns preserve existing exclusions and keep agent-tooling
// directories plus generated/vendored trees out of linting — deliberately
// not every dot-directory.

import { defineConfig } from "oxlint";

export default defineConfig({
	ignorePatterns: [
		".agent/**",
		".agents/**",
		".claude/**",
		".codex/**",
		".continue/**",
		".cursor/**",
		".gemini/**",
		".opencode/**",
		".pi/**",
		".roo/**",
		".windsurf/**",
		"tools/oxlint/anti-slop/**",
		"**/node_modules/**",
		"**/dist/**",
		"**/bindings/**",
	],
	jsPlugins: [
		{ name: "anti-slop", specifier: "./tools/oxlint/anti-slop/index.ts" },
	],
	rules: {
		// Anti-slop rules enforced as errors. These are the actionable
		// patterns the team wants to reject outright.
		"anti-slop/no-chained-type-assertions": "error",
		"anti-slop/no-conditional-empty-object-spread": "error",
		"anti-slop/no-known-value-widening": "error",
		"anti-slop/no-module-mocking": "error",
		"anti-slop/no-object-parameters": "error",
		"anti-slop/no-reflect-apply": "error",
		"anti-slop/no-reflect-get": "error",
		"anti-slop/no-shape-in-symbol-names": "error",
		"anti-slop/no-unknown-returns": "error",
		"anti-slop/no-unknown-type-aliases": "error",
		"anti-slop/no-widen-then-assert": "error",
		// The four rules below are deliberately relaxed to warnings: Reqly's
		// JSON tree, JSONPath, response-parsing and Wails-DTO code operates on
		// data that just crossed an I/O boundary, where ad hoc `typeof` guards
		// and `as` casts are the accepted boundary-parsing style. They stay
		// visible so each site can be revisited and re-promoted to error.
		"anti-slop/no-runtime-typeof": ["warn", { allowInTypeGuards: true }],
		"anti-slop/no-unknown-parameters": "warn",
		"anti-slop/no-unsafe-dictionary-type": "warn",
		"anti-slop/require-safety-comment-for-type-assertion": "warn",
	},
});
