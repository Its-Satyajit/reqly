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

// AI AGENTS: DO NOT MODIFY THIS FILE.
// This configuration is intentionally strict.
// Do not add, remove, weaken, downgrade, disable, or bypass rules.
// Do not change category severity from "error".
// If lint fails, fix the source code instead of weakening this configuration.
// Changes to this file require explicit human approval.

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
		// Vendored third-party import-test fixtures and spec sources.
		"internal/importer/testdata/**",
		// Design-experiment scratch area (plain JS demos, not shipped code).
		"ui-demos/**",
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
		// All anti-slop rules enforced as errors — 0 warnings tolerated.
		// Boundary parsing uses `typeGuards.ts` (isRecord/isString) + explicit domain types;
		// every `as` carries `// SAFETY:` (see `frontend/src/lib/typeGuards.ts`).
		"anti-slop/no-runtime-typeof": ["error", { allowInTypeGuards: true }],
		"anti-slop/no-unknown-parameters": "error",
		"anti-slop/no-unsafe-dictionary-type": "error",
		"anti-slop/require-safety-comment-for-type-assertion": "error",
	},
});
