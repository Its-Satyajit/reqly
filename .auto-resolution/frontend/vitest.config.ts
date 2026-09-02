import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

const root = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
	test: {
		environment: "jsdom",
		include: ["src/**/*.test.ts", "src/**/*.test.tsx"],
	},
	resolve: {
		alias: {
			"@": path.resolve(root, "src"),
			"#lib": path.resolve(root, "src/lib"),
			"#components": path.resolve(root, "src/components"),
			"#hooks": path.resolve(root, "src/hooks"),
			"#stores": path.resolve(root, "src/stores/index.ts"),
			"#stores/": `${path.resolve(root, "src/stores")}/`,
		},
	},
});
