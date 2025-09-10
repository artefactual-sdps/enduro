import { URL, fileURLToPath } from "node:url";

import vue from "@vitejs/plugin-vue";
import Icons from "unplugin-icons/vite";
import VueRouter from "unplugin-vue-router/vite";
import { defineConfig, loadEnv } from "vite";
import csp from "vite-plugin-csp-guard";
import vueDevTools from "vite-plugin-vue-devtools";

// Load environment variables from .env* files and the current process.
// The variables from the current process take precedence.
const fileEnv = loadEnv(process.env.NODE_ENV || "", process.cwd(), "");
const env = { ...fileEnv, ...process.env };

const cspPolicy: Record<string, string[]> = {
  "default-src": ["'self'"],
  "script-src-attr": ["'none'"],
  "img-src": ["'self'", "data:"],
  "frame-src": ["'none'"],
  "object-src": ["'none'"],
  "base-uri": ["'self'"],
  "form-action": ["'self'"],
};

if (env.VITE_OIDC_AUTHORITY) {
  cspPolicy["connect-src"] = ["'self'", env.VITE_OIDC_AUTHORITY];
}

if (env.VITE_INSTITUTION_LOGO) {
  cspPolicy["img-src"].push(env.VITE_INSTITUTION_LOGO);
}

if (env.NODE_ENV == "development") {
  cspPolicy["style-src-elem"] = ["'self'", "'unsafe-inline'"];
  cspPolicy["script-src-elem"] = ["'self'"]; // Needed to avoid SRI in dev.
  cspPolicy["frame-src"] = ["'self'"];
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    VueRouter({
      routesFolder: "src/pages",
    }),
    vue({}),
    vueDevTools(),
    Icons({ compiler: "vue3" }),
    // Include CSP plugin if not explicitly disabled.
    ...(env.ENDURO_DISABLE_CSP_META?.trim().toLowerCase() === "true"
      ? []
      : [
          csp({
            dev: { run: true },
            // Can't use SRI for local scripts/styles for now because the final
            // assets may be updated to replace/inject environment variables.
            // build: { sri: true },
            policy: cspPolicy,
            override: true,
          }),
        ]),
  ],
  server: {
    host: "127.0.0.1",
    port: 80,
    strictPort: true,
    proxy: {
      "/api": {
        target: env.ENDURO_API_ADDRESS
          ? "http://" + env.ENDURO_API_ADDRESS
          : "http://127.0.0.1:9000",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
      "^/api/ingest/monitor": {
        target: env.ENDURO_API_ADDRESS
          ? "http://" + env.ENDURO_API_ADDRESS
          : "http://127.0.0.1:9000",
        changeOrigin: false,
        rewrite: (path) => path.replace(/^\/api/, ""),
        ws: true,
      },
      "^/api/storage/monitor": {
        target: env.ENDURO_API_ADDRESS
          ? "http://" + env.ENDURO_API_ADDRESS
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
