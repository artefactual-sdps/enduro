import { fileURLToPath } from "node:url";
import { mergeConfig, defineConfig, configDefaults } from "vitest/config";
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
      // Needed by vue-testing-library.
      globals: true,
    },
  }),
);
