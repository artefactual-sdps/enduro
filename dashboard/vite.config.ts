import { URL, fileURLToPath } from "node:url";

import vue from "@vitejs/plugin-vue";
import Icons from "unplugin-icons/vite";
import VueRouter from "unplugin-vue-router/vite";
import { defineConfig } from "vite";
import vueDevTools from "vite-plugin-vue-devtools";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    VueRouter({
      routesFolder: "src/pages",
    }),
    vue({}),
    vueDevTools(),
    Icons({ compiler: "vue3" }),
  ],
  server: {
    host: "127.0.0.1",
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
      "^/api/ingest/monitor": {
        target: process.env.ENDURO_API_ADDRESS
          ? "http://" + process.env.ENDURO_API_ADDRESS
          : "http://127.0.0.1:9000",
        changeOrigin: false,
        rewrite: (path) => path.replace(/^\/api/, ""),
        ws: true,
      },
      "^/api/storage/monitor": {
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
        // Bootstrap v5.3 doesn't support the SASS modern API
        // (https://github.com/twbs/bootstrap/issues/40962),
        // so we need to use legacy mode. This has been removed
        // in Vite v7, which may require to upgrade Bootstrap v6
        // (https://github.com/twbs/bootstrap/pull/41512).
        api: "legacy",
        additionalData: `@import "src/styles/bootstrap-base.scss";`,
        silenceDeprecations: [
          "color-functions",
          "global-builtin",
          "import",
          "legacy-js-api",
        ],
      },
    },
  },
});
