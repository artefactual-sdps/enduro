import DOMPurify from "dompurify";
import { acceptHMRUpdate, defineStore } from "pinia";

interface Route {
  name: string;
  path: string;
  url: string;
}

interface Manifest {
  homeUrl?: string;
  stylesheets?: string[];
  scripts?: string[];
  routes?: Route[];
}

export const useCustomStore = defineStore("custom", {
  state: () => ({
    loaded: false,
    manifest: null as Manifest | null,
    homeContent: null as string | null,
    homeLoading: false,
    homeError: null as string | null,
    stylesApplied: false,
    scriptsApplied: false,
    routeContents: new Map<string, string>(),
    routeErrors: new Map<string, string>(),
  }),
  actions: {
    async loadManifest() {
      if (this.loaded) return;

      const manifestUrl = import.meta.env.VITE_CUSTOM_MANIFEST.trim();
      if (!manifestUrl) {
        this.loaded = true;
        return;
      }

      try {
        const res = await fetch(manifestUrl);
        if (!res.ok) throw new Error(`Failed to load manifest: ${res.status}`);
        this.manifest = (await res.json()) as Manifest;
        await this.applyStyles();
        await this.applyScripts();
      } catch (e) {
        // TODO: notify the user about this error.
        console.error("Error loading customization manifest:", e);
        this.manifest = null;
      } finally {
        this.loaded = true;
      }
    },

    async applyStyles() {
      if (this.stylesApplied || !this.manifest?.stylesheets) return;

      const stylePromises = this.manifest.stylesheets.map((cssUrl) => {
        return new Promise<void>((resolve, reject) => {
          const link = document.createElement("link");
          link.rel = "stylesheet";
          link.href = cssUrl;
          link.setAttribute("data-custom-styles", "true");
          link.onload = () => resolve();
          link.onerror = () =>
            reject(new Error(`Failed to load stylesheet: ${cssUrl}`));
          document.head.appendChild(link);
        });
      });

      try {
        await Promise.all(stylePromises);
      } catch (e) {
        console.error("Error loading custom stylesheets:", e);
      }

      this.stylesApplied = true;
    },

    async applyScripts() {
      if (this.scriptsApplied || !this.manifest?.scripts) return;

      const scriptPromises = this.manifest.scripts.map((scriptUrl) => {
        return new Promise<void>((resolve, reject) => {
          const script = document.createElement("script");
          script.src = scriptUrl;
          script.setAttribute("data-custom-script", "true");
          script.onload = () => resolve();
          script.onerror = () =>
            reject(new Error(`Failed to load script: ${scriptUrl}`));
          document.head.appendChild(script);
        });
      });

      try {
        await Promise.all(scriptPromises);
      } catch (e) {
        console.error("Error loading custom scripts:", e);
      }

      this.scriptsApplied = true;
    },

    async loadHomeContent() {
      if (!this.loaded) await this.loadManifest();

      if (
        !this.manifest?.homeUrl ||
        this.homeLoading ||
        this.homeContent ||
        this.homeError
      )
        return;

      this.homeLoading = true;
      this.homeError = null;

      try {
        const res = await fetch(this.manifest.homeUrl);
        if (!res.ok) throw new Error(`Response status: ${res.status}`);
        const content = DOMPurify.sanitize(await res.text());
        if (!content) throw new Error("Sanitized content is empty.");
        this.homeContent = content;
      } catch (e) {
        console.error("Error loading custom home HTML:", e);
        this.homeError = "Failed to load custom home content.";
        this.homeContent = null;
      } finally {
        this.homeLoading = false;
      }
    },

    async loadRouteContent(routePath: string): Promise<string | null> {
      if (!this.loaded) await this.loadManifest();

      if (this.routeContents.has(routePath)) {
        return this.routeContents.get(routePath) || null;
      }
      if (this.routeErrors.has(routePath)) {
        return null;
      }

      const route = this.manifest?.routes?.find((r) => r.path === routePath);
      if (!route) {
        this.routeErrors.set(routePath, "Route not found in manifest");
        return null;
      }

      try {
        const res = await fetch(route.url);
        if (!res.ok) throw new Error(`Response status: ${res.status}`);
        const content = DOMPurify.sanitize(await res.text());
        if (!content) throw new Error("Sanitized content is empty.");
        this.routeContents.set(routePath, content);
        return content;
      } catch (e) {
        console.error(
          `Error loading custom route content for ${routePath}:`,
          e,
        );
        const errorMsg = `Failed to load content for route: ${routePath}`;
        this.routeErrors.set(routePath, errorMsg);
        return null;
      }
    },

    getRouteContent(routePath: string): string | null {
      return this.routeContents.get(routePath) || null;
    },

    getRouteError(routePath: string): string | null {
      return this.routeErrors.get(routePath) || null;
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useCustomStore, import.meta.hot));
}
