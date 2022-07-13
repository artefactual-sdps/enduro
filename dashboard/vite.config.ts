import vue from "@vitejs/plugin-vue";
import { resolve } from "path";
import Icons from "unplugin-icons/vite";
import Pages from "vite-plugin-pages";
import { configDefaults, defineConfig } from "vitest/config";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue({ reactivityTransform: true }),
    Pages(),
    Icons({ compiler: "vue3" }),
  ],
  server: {
    proxy: {
      "/api": {
        target: process.env.ENDURO_API_ADDRESS || "http://127.0.0.1:9000",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
      "^/api/package/monitor": {
        target: process.env.ENDURO_API_ADDRESS || "http://127.0.0.1:9000",
        changeOrigin: false,
        rewrite: (path) => path.replace(/^\/api/, ""),
        ws: true,
      },
    },
  },
  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
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
      exclude: ["src/openapi-generator/**"],
    },
  },
});
