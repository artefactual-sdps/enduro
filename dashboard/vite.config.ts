import vue from "@vitejs/plugin-vue";
import * as path from "path";
import { defineConfig } from "vite";
import Pages from "vite-plugin-pages";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue({ reactivityTransform: true }), Pages()],
  server: {
    proxy: {
      "/api": {
        target: process.env.ENDURO_API_ADDRESS || "http://127.0.0.1:9000",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
});
