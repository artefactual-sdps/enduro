import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useCustomStore } from "../custom";

describe("custom store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("initializes with default state", () => {
    const store = useCustomStore();

    expect(store.manifest).toBeNull();
    expect(store.loaded).toBe(false);
    expect(store.homeContent).toBeNull();
    expect(store.homeLoading).toBe(false);
    expect(store.homeError).toBeNull();
    expect(store.stylesApplied).toBe(false);
  });

  it("loads manifest successfully", async () => {
    const mockManifest = { homeUrl: "/custom/home.html" };
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => mockManifest,
      } as Response),
    );
    vi.stubEnv("VITE_CUSTOM_MANIFEST", "/custom/manifest.json");

    const store = useCustomStore();
    await store.loadManifest();

    expect(store.manifest).toEqual(mockManifest);
    expect(store.loaded).toBe(true);
  });

  it("handles manifest load failure", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValueOnce({ ok: false, status: 404 } as Response),
    );
    vi.stubEnv("VITE_CUSTOM_MANIFEST", "/custom/manifest.json");

    const store = useCustomStore();
    await store.loadManifest();

    expect(store.manifest).toBeNull();
    expect(store.loaded).toBe(true);
  });

  it("does not load manifest if already loaded", async () => {
    vi.stubGlobal("fetch", vi.fn());
    vi.stubEnv("VITE_CUSTOM_MANIFEST", "/custom/manifest.json");

    const store = useCustomStore();
    store.loaded = true;
    await store.loadManifest();

    expect(fetch).not.toHaveBeenCalled();
  });

  it("does not load manifest if env var is not set", async () => {
    vi.stubGlobal("fetch", vi.fn());
    vi.stubEnv("VITE_CUSTOM_MANIFEST", "");

    const store = useCustomStore();
    await store.loadManifest();

    expect(fetch).not.toHaveBeenCalled();
    expect(store.loaded).toBe(true);
  });

  it("loads home content successfully", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValueOnce({
        ok: true,
        text: async () => "<h1>Custom Home</h1>",
      } as Response),
    );

    const store = useCustomStore();
    store.manifest = { homeUrl: "/custom/home.html" };
    store.loaded = true;
    await store.loadHomeContent();

    expect(store.homeContent).toBe("<h1>Custom Home</h1>");
    expect(store.homeLoading).toBe(false);
    expect(store.homeError).toBeNull();
  });

  it("sanitizes home content", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValueOnce({
        ok: true,
        text: async () => '<h1>Safe</h1><script>alert("xss")</script>',
      } as Response),
    );

    const store = useCustomStore();
    store.manifest = { homeUrl: "/custom/home.html" };
    store.loaded = true;
    await store.loadHomeContent();

    expect(store.homeContent).toBe("<h1>Safe</h1>");
    expect(store.homeContent).not.toContain("script");
  });

  it("handles home content load failure", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValueOnce({ ok: false, status: 404 } as Response),
    );

    const store = useCustomStore();
    store.manifest = { homeUrl: "/custom/home.html" };
    store.loaded = true;
    await store.loadHomeContent();

    expect(store.homeContent).toBeNull();
    expect(store.homeError).toBe("Failed to load custom home content.");
  });

  it("handles empty sanitized content", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValueOnce({
        ok: true,
        text: async () => '<script>alert("only malicious")</script>',
      } as Response),
    );

    const store = useCustomStore();
    store.manifest = { homeUrl: "/custom/home.html" };
    store.loaded = true;
    await store.loadHomeContent();

    expect(store.homeContent).toBeNull();
    expect(store.homeError).toBe("Failed to load custom home content.");
  });

  it("does not load home content if no custom home configured", async () => {
    vi.stubGlobal("fetch", vi.fn());

    const store = useCustomStore();
    store.manifest = {};
    store.loaded = true;
    await store.loadHomeContent();

    expect(fetch).not.toHaveBeenCalled();
  });

  it("does not load home content if already loaded", async () => {
    vi.stubGlobal("fetch", vi.fn());

    const store = useCustomStore();
    store.manifest = { homeUrl: "/custom/home.html" };
    store.loaded = true;
    store.homeContent = "<h1>Already loaded</h1>";
    await store.loadHomeContent();

    expect(fetch).not.toHaveBeenCalled();
  });

  it("does not load home content if there was a previous error", async () => {
    vi.stubGlobal("fetch", vi.fn());

    const store = useCustomStore();
    store.manifest = { homeUrl: "/custom/home.html" };
    store.loaded = true;
    store.homeError = "Previous error";
    await store.loadHomeContent();

    expect(fetch).not.toHaveBeenCalled();
  });

  describe("style customizations", () => {
    afterEach(() => {
      document.head
        .querySelectorAll('link[data-custom-styles="true"]')
        .forEach((link) => link.remove());
    });

    it("loads external CSS files from manifest", async () => {
      const mockManifest = {
        stylesheets: ["/custom/styles.css", "/custom/theme.css"],
      };
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValueOnce({
          ok: true,
          json: async () => mockManifest,
        } as Response),
      );
      vi.stubEnv("VITE_CUSTOM_MANIFEST", "/custom/manifest.json");

      const store = useCustomStore();
      await store.loadManifest();

      const linkElements = document.querySelectorAll(
        'link[data-custom-styles="true"]',
      );
      expect(linkElements.length).toBe(2);
      expect(linkElements[0]?.getAttribute("href")).toBe("/custom/styles.css");
      expect(linkElements[1]?.getAttribute("href")).toBe("/custom/theme.css");
      expect(store.stylesApplied).toBe(true);
    });

    it("does not apply styles if already applied", () => {
      const store = useCustomStore();
      store.manifest = { stylesheets: ["/custom/styles.css"] };
      store.stylesApplied = true;
      store.applyStyles();

      const linkElements = document.querySelectorAll(
        'link[data-custom-styles="true"]',
      );
      expect(linkElements.length).toBe(0);
    });

    it("does not apply styles if no stylesheets in manifest", async () => {
      const store = useCustomStore();
      store.manifest = {};
      store.applyStyles();

      const linkElements = document.querySelectorAll(
        'link[data-custom-styles="true"]',
      );
      expect(linkElements.length).toBe(0);
      expect(store.stylesApplied).toBe(false);
    });
  });

  describe("script customizations", () => {
    afterEach(() => {
      document.head
        .querySelectorAll('script[data-custom-script="true"]')
        .forEach((script) => script.remove());
    });

    it("loads external JS files from manifest", async () => {
      const mockManifest = {
        scripts: ["/custom/script1.js", "/custom/script2.js"],
      };
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValueOnce({
          ok: true,
          json: async () => mockManifest,
        } as Response),
      );
      vi.stubEnv("VITE_CUSTOM_MANIFEST", "/custom/manifest.json");

      const store = useCustomStore();
      await store.loadManifest();

      const scriptElements = document.querySelectorAll(
        'script[data-custom-script="true"]',
      );
      expect(scriptElements.length).toBe(2);
      expect(scriptElements[0]?.getAttribute("src")).toBe("/custom/script1.js");
      expect(scriptElements[1]?.getAttribute("src")).toBe("/custom/script2.js");
      expect(store.scriptsApplied).toBe(true);
    });

    it("does not apply scripts if already applied", () => {
      const store = useCustomStore();
      store.manifest = { scripts: ["/custom/script.js"] };
      store.scriptsApplied = true;
      store.applyScripts();

      const scriptElements = document.querySelectorAll(
        'script[data-custom-script="true"]',
      );
      expect(scriptElements.length).toBe(0);
    });

    it("does not apply scripts if no scripts in manifest", async () => {
      const store = useCustomStore();
      store.manifest = {};
      store.applyScripts();

      const scriptElements = document.querySelectorAll(
        'script[data-custom-script="true"]',
      );
      expect(scriptElements.length).toBe(0);
      expect(store.scriptsApplied).toBe(false);
    });
  });

  describe("custom routes", () => {
    it("loads custom route content successfully", async () => {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValueOnce({
          ok: true,
          text: async () => "<h1>About Page</h1>",
        } as Response),
      );

      const store = useCustomStore();
      store.manifest = {
        routes: [{ path: "/about", name: "About", url: "/custom/about.html" }],
      };
      store.loaded = true;

      const content = await store.loadRouteContent("/about");

      expect(content).toBe("<h1>About Page</h1>");
      expect(store.getRouteContent("/about")).toBe("<h1>About Page</h1>");
      expect(store.getRouteError("/about")).toBeNull();
    });

    it("sanitizes custom route content", async () => {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValueOnce({
          ok: true,
          text: async () => '<h1>Safe</h1><script>alert("xss")</script>',
        } as Response),
      );

      const store = useCustomStore();
      store.manifest = {
        routes: [
          {
            path: "/test",
            name: "Test",
            url: "/custom/test.html",
          },
        ],
      };
      store.loaded = true;

      const content = await store.loadRouteContent("/test");

      expect(content).toBe("<h1>Safe</h1>");
      expect(content).not.toContain("script");
    });

    it("handles route content load failure", async () => {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValueOnce({ ok: false, status: 404 } as Response),
      );

      const store = useCustomStore();
      store.manifest = {
        routes: [
          {
            path: "/notfound",
            name: "NotFound",
            url: "/custom/notfound.html",
          },
        ],
      };
      store.loaded = true;

      const content = await store.loadRouteContent("/notfound");

      expect(content).toBeNull();
      expect(store.getRouteError("/notfound")).toContain(
        "Failed to load content for route",
      );
    });

    it("returns null for missing route", async () => {
      vi.stubGlobal("fetch", vi.fn());

      const store = useCustomStore();
      store.manifest = { routes: [] };
      store.loaded = true;

      const content = await store.loadRouteContent("/missing");

      expect(content).toBeNull();
      expect(store.getRouteError("/missing")).toBe(
        "Route not found in manifest",
      );
      expect(fetch).not.toHaveBeenCalled();
    });

    it("caches loaded route content", async () => {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValueOnce({
          ok: true,
          text: async () => "<h1>Cached Page</h1>",
        } as Response),
      );

      const store = useCustomStore();
      store.manifest = {
        routes: [
          {
            path: "/cached",
            name: "Cached",
            url: "/custom/cached.html",
          },
        ],
      };
      store.loaded = true;

      const content1 = await store.loadRouteContent("/cached");
      const content2 = await store.loadRouteContent("/cached");

      expect(content1).toBe("<h1>Cached Page</h1>");
      expect(content2).toBe("<h1>Cached Page</h1>");
      expect(fetch).toHaveBeenCalledTimes(1);
    });

    it("does not retry loading after error", async () => {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValueOnce({ ok: false, status: 500 } as Response),
      );

      const store = useCustomStore();
      store.manifest = {
        routes: [
          {
            path: "/error",
            name: "Error",
            url: "/custom/error.html",
          },
        ],
      };
      store.loaded = true;

      const content1 = await store.loadRouteContent("/error");
      const content2 = await store.loadRouteContent("/error");

      expect(content1).toBeNull();
      expect(content2).toBeNull();
      expect(fetch).toHaveBeenCalledTimes(1);
    });
  });
});
