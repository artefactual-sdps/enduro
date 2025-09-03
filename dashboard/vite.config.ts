import { URL, fileURLToPath } from "node:url";

import vue from "@vitejs/plugin-vue";
import Icons from "unplugin-icons/vite";
import VueRouter from "unplugin-vue-router/vite";
import { defineConfig } from "vite";
import csp from "vite-plugin-csp-guard";
import vueDevTools from "vite-plugin-vue-devtools";

const apiHost = process.env.ENDURO_API_ADDRESS || "127.0.0.1:9000";
const apiProtocol = process.env.ENDURO_API_PROTOCOL || "http";
const wsProtocol = apiProtocol === "https" ? "wss" : "ws";
const apiUrl = `${apiProtocol}://${apiHost}`;
const wsUrl = `${wsProtocol}://${apiHost}`;

// Base CSP policy.
const basePolicy = {
  "default-src": ["'self'"],
  "style-src": ["'self'", "'unsafe-inline'"],
  "font-src": ["'self'", "data:"],
  "img-src": ["'self'", "data:"],
  "connect-src": [
    "'self'",
    apiUrl,
    `${wsUrl}/ingest/monitor`,
    `${wsUrl}/storage/monitor`,
  ],
  "frame-src": ["'none'"],
  "object-src": ["'none'"],
  "base-uri": ["'self'"],
};

// Development CSP policy.
const devPolicy = {
  ...basePolicy,
  "script-src": ["'self'", "'unsafe-eval'", "'unsafe-inline'"],
};

// Production CSP policy.
const prodPolicy = {
  ...basePolicy,
  "script-src": ["'self'"],
  "form-action": ["'self'"],
  "frame-ancestors": ["'self'"],
  "upgrade-insecure-requests": [],
};

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    VueRouter({
      routesFolder: "src/pages",
    }),
    vue({}),
    vueDevTools(),
    Icons({ compiler: "vue3" }),
    csp({
      dev: {
        run: true,
        policy: devPolicy,
      },
      policy: prodPolicy,
      build: {
        sri: true,
      },
    }),
  ],
  server: {
    host: "127.0.0.1",
    port: 80,
    strictPort: true,
    proxy: {
      "/api": {
        target: apiUrl,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
      "^/api/ingest/monitor": {
        target: apiUrl,
        changeOrigin: false,
        rewrite: (path) => path.replace(/^\/api/, ""),
        ws: true,
      },
      "^/api/storage/monitor": {
        target: apiUrl,
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
