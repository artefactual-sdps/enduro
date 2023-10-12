import vue from "@vitejs/plugin-vue";
import ReactivityTransform from "@vue-macros/reactivity-transform/vite";
import { fileURLToPath, URL } from "node:url";
import Icons from "unplugin-icons/vite";
import Pages from "vite-plugin-pages";
import { defineConfig } from "vitest/config";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue({}),
    ReactivityTransform(),
    Pages(),
    Icons({ compiler: "vue3" }),
  ],
  // Use esbuild deps optimization at build time.
  // https://vitejs.dev/guide/migration.html#using-esbuild-deps-optimization-at-build-time
  build: {
    commonjsOptions: {
      include: [],
    },
  },
  optimizeDeps: {
    disabled: false,
  },
  server: {
    port: 80,
    strictPort: true,
    proxy: {
      "/api": {
        target: process.env.ENDURO_API_ADDRESS
          ? "http://" + process.env.ENDURO_API_ADDRESS
          : "http://127.0.0.1:9000",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
      "^/api/package/monitor": {
        target: process.env.ENDURO_API_ADDRESS
          ? "http://" + process.env.ENDURO_API_ADDRESS
          : "http://127.0.0.1:9000",
        changeOrigin: false,
        rewrite: (path) => path.replace(/^\/api/, ""),
        ws: true,
      },
    },
  },
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@import "src/styles/bootstrap-base.scss";`,
      },
    },
  },
  test: {
    environment: "happy-dom",
    restoreMocks: true,
    sequence: {
      shuffle: true,
    },
    coverage: {
      exclude: ["src/openapi-generator/**", "test/**"],
    },
  },
});
