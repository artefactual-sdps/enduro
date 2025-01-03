import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import Icons from "unplugin-icons/vite";
import VueRouter from "unplugin-vue-router/vite";
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
        // TODO: remove this line once bootstrap v5.3.4 is released.
        silenceDeprecations: ["mixed-decls"],
      },
    },
  },
});
