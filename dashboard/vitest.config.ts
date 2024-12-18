import { fileURLToPath } from "node:url";

import { configDefaults, defineConfig, mergeConfig } from "vitest/config";

import viteConfig from "./vite.config";

export default mergeConfig(
  viteConfig,
  defineConfig({
    test: {
      environment: "happy-dom",
      exclude: [...configDefaults.exclude, "e2e/*"],
      root: fileURLToPath(new URL("./", import.meta.url)),
      restoreMocks: true,
      sequence: {
        shuffle: true,
      },
      coverage: {
        exclude: ["src/openapi-generator/**"],
      },
      globalSetup: "./src/test-globals.ts",
      // Needed by vue-testing-library.
      globals: true,
    },
  }),
);
