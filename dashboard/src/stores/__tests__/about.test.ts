import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, client } from "@/client";
import { useAboutStore } from "@/stores/about";

vi.mock("@/client");

describe("useAboutStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  describe("state", () => {
    it("initializes with correct default values", () => {
      const aboutStore = useAboutStore();

      expect(aboutStore.loaded).toBe(false);
      expect(aboutStore.childWorkflows).toEqual([]);
      expect(aboutStore.preservationSystem).toBe("a3m");
      expect(aboutStore.uploadMaxSize).toBe(0);
      expect(aboutStore.version).toBe("");
    });
  });

  describe("getters", () => {
    describe("formattedUploadMaxSize", () => {
      it("formats bytes correctly", () => {
        const aboutStore = useAboutStore();
        aboutStore.uploadMaxSize = 512;
        expect(aboutStore.formattedUploadMaxSize).toBe("512 bytes");
      });

      it("formats KiB correctly", () => {
        const aboutStore = useAboutStore();
        aboutStore.uploadMaxSize = 1536; // 1.5 KiB
        expect(aboutStore.formattedUploadMaxSize).toBe("1.50 KiB");
      });

      it("formats MiB correctly", () => {
        const aboutStore = useAboutStore();
        aboutStore.uploadMaxSize = 1572864; // 1.5 MiB
        expect(aboutStore.formattedUploadMaxSize).toBe("1.50 MiB");
      });

      it("formats GiB correctly", () => {
        const aboutStore = useAboutStore();
        aboutStore.uploadMaxSize = 1610612736; // 1.5 GiB
        expect(aboutStore.formattedUploadMaxSize).toBe("1.50 GiB");
      });

      it("formats TiB correctly", () => {
        const aboutStore = useAboutStore();
        aboutStore.uploadMaxSize = 1.5 * 1024 ** 4; // 1.5 TiB
        expect(aboutStore.formattedUploadMaxSize).toBe("1.50 TiB");
      });

      it("handles edge case at 1024 bytes", () => {
        const aboutStore = useAboutStore();
        aboutStore.uploadMaxSize = 1024;
        expect(aboutStore.formattedUploadMaxSize).toBe("1.00 KiB");
      });

      it("handles edge case at 1 MiB", () => {
        const aboutStore = useAboutStore();
        aboutStore.uploadMaxSize = 1024 * 1024;
        expect(aboutStore.formattedUploadMaxSize).toBe("1.00 MiB");
      });

      it("handles edge case at 1 GiB", () => {
        const aboutStore = useAboutStore();
        aboutStore.uploadMaxSize = 1024 ** 3;
        expect(aboutStore.formattedUploadMaxSize).toBe("1.00 GiB");
      });

      it("handles edge case at 1 TiB", () => {
        const aboutStore = useAboutStore();
        aboutStore.uploadMaxSize = 1024 ** 4;
        expect(aboutStore.formattedUploadMaxSize).toBe("1.00 TiB");
      });
    });

    describe("formattedVersion", () => {
      it("formats version with v prefix", () => {
        const aboutStore = useAboutStore();
        aboutStore.version = "1.2.3";
        expect(aboutStore.formattedVersion).toBe("v1.2.3");
      });

      it("handles empty version", () => {
        const aboutStore = useAboutStore();
        aboutStore.version = "";
        expect(aboutStore.formattedVersion).toBe("unknown");
      });
    });
  });

  describe("actions", () => {
    describe("fetch", () => {
      it("fetches and updates state successfully", async () => {
        const mockResponse: api.EnduroAbout = {
          childWorkflows: [
            {
              type: "preprocessing",
              taskQueue: "pp-queue",
              workflowName: "pp-workflow",
            },
            {
              type: "poststorage",
              taskQueue: "ps-queue",
              workflowName: "ps-workflow",
            },
          ],
          preservationSystem: "archivematica",
          uploadMaxSize: 1073741824, // 1 GiB
          version: "1.0.0",
        };

        client.about.aboutAbout = vi.fn().mockResolvedValue(mockResponse);

        const aboutStore = useAboutStore();
        await aboutStore.fetch();

        expect(aboutStore.loaded).toBe(true);
        expect(aboutStore.childWorkflows).toEqual([
          {
            type: "preprocessing",
            taskQueue: "pp-queue",
            workflowName: "pp-workflow",
          },
          {
            type: "poststorage",
            taskQueue: "ps-queue",
            workflowName: "ps-workflow",
          },
        ]);
        expect(aboutStore.preservationSystem).toBe("archivematica");
        expect(aboutStore.uploadMaxSize).toBe(1073741824);
        expect(aboutStore.version).toBe("1.0.0");
      });

      it("handles fetch errors", async () => {
        const consoleErrorSpy = vi
          .spyOn(console, "error")
          .mockImplementation(() => {});
        const error = new Error("Network error");

        client.about.aboutAbout = vi.fn().mockRejectedValue(error);

        const aboutStore = useAboutStore();
        await aboutStore.fetch();

        expect(consoleErrorSpy).toHaveBeenCalledWith(
          "Error fetching about data:",
          "Network error",
        );
        expect(aboutStore.loaded).toBe(false);

        consoleErrorSpy.mockRestore();
      });
    });

    describe("load", () => {
      it("calls fetch when not loaded", async () => {
        const aboutStore = useAboutStore();
        const fetchSpy = vi.spyOn(aboutStore, "fetch").mockResolvedValue();

        await aboutStore.load();

        expect(fetchSpy).toHaveBeenCalledOnce();
      });

      it("does not call fetch when already loaded", async () => {
        const aboutStore = useAboutStore();
        aboutStore.loaded = true;
        const fetchSpy = vi.spyOn(aboutStore, "fetch").mockResolvedValue();

        await aboutStore.load();

        expect(fetchSpy).not.toHaveBeenCalled();
      });

      it("returns fetch promise when not loaded", async () => {
        const mockResponse: api.EnduroAbout = {
          childWorkflows: [],
          preservationSystem: "a3m",
          uploadMaxSize: 0,
          version: "1.0.0",
        };

        client.about.aboutAbout = vi.fn().mockResolvedValue(mockResponse);

        const aboutStore = useAboutStore();
        const result = await aboutStore.load();

        expect(aboutStore.loaded).toBe(true);
        expect(result).toBeUndefined();
      });
    });
  });

  describe("integration", () => {
    it("formats upload max size and version after successful fetch", async () => {
      const mockResponse: api.EnduroAbout = {
        childWorkflows: [],
        preservationSystem: "a3m",
        uploadMaxSize: 2147483648, // 2 GiB
        version: "2.0.0",
      };

      client.about.aboutAbout = vi.fn().mockResolvedValue(mockResponse);

      const aboutStore = useAboutStore();
      await aboutStore.fetch();

      expect(aboutStore.formattedUploadMaxSize).toBe("2.00 GiB");
      expect(aboutStore.formattedVersion).toBe("v2.0.0");
    });
  });
});
