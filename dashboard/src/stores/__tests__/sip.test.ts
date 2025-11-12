import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { api, client } from "@/client";
import { ResponseError } from "@/openapi-generator";
import { useLayoutStore } from "@/stores/layout";
import { useSipStore } from "@/stores/sip";

vi.mock("@/client");

describe("useSipStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("isPending", () => {
    const sipStore = useSipStore();
    const now = new Date();

    expect(sipStore.isPending).toEqual(false);

    sipStore.$patch({
      current: {
        createdAt: now,
        uuid: "a499e8fc-7309-4e26-b39d-d8ab68466c27",
        status: api.EnduroIngestSipStatusEnum.Pending,
      },
    });
    expect(sipStore.isPending).toEqual(true);
  });

  it("getActionById finds actions", () => {
    const sipStore = useSipStore();
    const now = new Date();
    sipStore.$patch({
      currentWorkflows: {
        workflows: [
          {
            uuid: "workflow-uuid-1",
            type: api.EnduroIngestSipWorkflowTypeEnum.CreateAip,
            startedAt: now,
            status: api.EnduroIngestSipWorkflowStatusEnum.Done,
            temporalId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
          },
          {
            uuid: "workflow-uuid-2",
            type: api.EnduroIngestSipWorkflowTypeEnum.CreateAndReviewAip,
            startedAt: now,
            status: api.EnduroIngestSipWorkflowStatusEnum.Done,
            temporalId: "051cf998-6f87-4461-8091-8561ebf479c4",
          },
        ],
      },
    });

    expect(sipStore.getWorkflowById("workflow-uuid-1")).toEqual({
      uuid: "workflow-uuid-1",
      type: api.EnduroIngestSipWorkflowTypeEnum.CreateAip,
      startedAt: now,
      status: api.EnduroIngestSipWorkflowStatusEnum.Done,
      temporalId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
    });
    expect(sipStore.getWorkflowById("workflow-uuid-2")).toEqual({
      uuid: "workflow-uuid-2",
      type: api.EnduroIngestSipWorkflowTypeEnum.CreateAndReviewAip,
      startedAt: now,
      status: api.EnduroIngestSipWorkflowStatusEnum.Done,
      temporalId: "051cf998-6f87-4461-8091-8561ebf479c4",
    });
    expect(sipStore.getWorkflowById("workflow-uuid-3")).toBeUndefined();
  });

  it("getTaskById finds tasks", () => {
    const sipStore = useSipStore();
    const now = new Date();
    sipStore.$patch({
      currentWorkflows: {
        workflows: [
          {
            uuid: "workflow-uuid-1",
            type: api.EnduroIngestSipWorkflowTypeEnum.CreateAip,
            startedAt: now,
            status: api.EnduroIngestSipWorkflowStatusEnum.Done,
            temporalId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
            tasks: [
              {
                uuid: "task-uuid-1",
                name: "Task 1",
                startedAt: now,
                status: api.EnduroIngestSipTaskStatusEnum.Done,
              },
              {
                uuid: "task-uuid-2",
                name: "Task 2",
                startedAt: now,
                status: api.EnduroIngestSipTaskStatusEnum.Done,
              },
            ],
          },
          {
            uuid: "workflow-uuid-2",
            type: api.EnduroIngestSipWorkflowTypeEnum.CreateAndReviewAip,
            startedAt: now,
            status: api.EnduroIngestSipWorkflowStatusEnum.Done,
            temporalId: "051cf998-6f87-4461-8091-8561ebf479c4",
            tasks: [
              {
                uuid: "task-uuid-3",
                name: "Task 3",
                startedAt: now,
                status: api.EnduroIngestSipTaskStatusEnum.Done,
              },
              {
                uuid: "task-uuid-4",
                name: "Task 4",
                startedAt: now,
                status: api.EnduroIngestSipTaskStatusEnum.Done,
              },
            ],
          },
        ],
      },
    });

    expect(sipStore.getTaskById("workflow-uuid-1", "task-uuid-1")).toEqual({
      uuid: "task-uuid-1",
      name: "Task 1",
      startedAt: now,
      status: api.EnduroIngestSipTaskStatusEnum.Done,
    });
    expect(sipStore.getTaskById("workflow-uuid-1", "task-uuid-2")).toEqual({
      uuid: "task-uuid-2",
      name: "Task 2",
      startedAt: now,
      status: api.EnduroIngestSipTaskStatusEnum.Done,
    });
    expect(
      sipStore.getTaskById("workflow-uuid-1", "task-uuid-3"),
    ).toBeUndefined();

    expect(sipStore.getTaskById("workflow-uuid-2", "task-uuid-3")).toEqual({
      uuid: "task-uuid-3",
      name: "Task 3",
      startedAt: now,
      status: api.EnduroIngestSipTaskStatusEnum.Done,
    });
    expect(sipStore.getTaskById("workflow-uuid-2", "task-uuid-4")).toEqual({
      uuid: "task-uuid-4",
      name: "Task 4",
      startedAt: now,
      status: api.EnduroIngestSipTaskStatusEnum.Done,
    });
    expect(
      sipStore.getTaskById("workflow-uuid-2", "task-uuid-5"),
    ).toBeUndefined();
  });

  it("fetches current SIP", async () => {
    const layoutStore = useLayoutStore();
    const store = useSipStore();
    const mockSip: api.EnduroIngestSip = {
      uuid: "a499e8fc-7309-4e26-b39d-d8ab68466c27",
      name: "SIP 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      status: api.EnduroIngestSipStatusEnum.Ingested,
    };

    client.ingest.ingestShowSip = vi.fn().mockResolvedValue(mockSip);

    await store.fetchCurrent("1");

    expect(store.current).toEqual(mockSip);
    expect(layoutStore.breadcrumb).toEqual([
      { text: "Ingest" },
      { route: expect.any(Object), text: "SIPs" },
      { text: mockSip.name },
    ]);
  });

  it("throws an error when fetching current SIP fails", async () => {
    const layoutStore = useLayoutStore();
    const store = useSipStore();
    client.ingest.ingestShowSip = vi.fn().mockRejectedValue(
      new ResponseError(
        new Response("Not Found", {
          status: 404,
          statusText: "Not Found",
        }),
        "Response returned an error code",
      ),
    );

    await expect(store.fetchCurrent("1")).rejects.toThrow("Couldn't load SIP");
    expect(store.current).toBeNull();
    expect(layoutStore.breadcrumb).toEqual([
      { text: "Ingest" },
      { route: expect.any(Object), text: "SIPs" },
      { text: "Error" },
    ]);
  });

  it("fetches current workflows", async () => {
    const mockWorkflows: api.EnduroIngestSipWorkflows = {
      workflows: [
        {
          uuid: "workflow-uuid-1",
          startedAt: new Date("2025-01-01T00:00:00Z"),
          status: api.EnduroIngestSipWorkflowStatusEnum.Done,
          type: api.EnduroIngestSipWorkflowTypeEnum.CreateAip,
          temporalId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
          sipUuid: "a499e8fc-7309-4e26-b39d-d8ab68466c27",
        },
      ],
    };

    client.ingest.ingestListSipWorkflows = vi
      .fn()
      .mockResolvedValue(mockWorkflows);

    const store = useSipStore();
    await store.fetchCurrentWorkflows("sip-uuid");

    expect(store.currentWorkflows).toEqual(mockWorkflows);
  });

  it("throws an error when fetching workflows fails", async () => {
    const store = useSipStore();
    client.ingest.ingestListSipWorkflows = vi.fn().mockRejectedValue(
      new ResponseError(
        new Response("Not Found", {
          status: 404,
          statusText: "Not Found",
        }),
        "Response returned an error code",
      ),
    );
    await expect(store.fetchCurrentWorkflows("sip-uuid")).rejects.toThrow(
      "Couldn't load workflows",
    );
    expect(store.currentWorkflows).toBeNull();
  });

  it("suppresses a 403 Forbidden response when fetching workflows", async () => {
    const store = useSipStore();

    client.ingest.ingestListSipWorkflows = vi.fn().mockRejectedValue(
      new ResponseError(
        new Response("Forbidden", {
          status: 403,
          statusText: "Forbidden",
        }),
        "Response returned an error code",
      ),
    );

    await expect(
      store.fetchCurrentWorkflows("sip-uuid"),
    ).resolves.toBeUndefined();
    expect(store.currentWorkflows).toBeNull();
  });

  it("fetches SIPs", async () => {
    const mockSips: api.EnduroIngestSips = {
      items: [
        {
          uuid: "a499e8fc-7309-4e26-b39d-d8ab68466c27",
          name: "SIP 1",
          createdAt: new Date("2025-01-01T00:00:00Z"),
          status: api.EnduroIngestSipStatusEnum.Ingested,
        },
        {
          uuid: "30223842-0650-4f79-80bd-7bf43b810656",
          name: "SIP 2",
          createdAt: new Date("2025-01-02T00:00:00Z"),
          status: api.EnduroIngestSipStatusEnum.Ingested,
        },
      ],
      page: { limit: 20, offset: 0, total: 2 },
    };
    client.ingest.ingestListSips = vi.fn().mockResolvedValue(mockSips);

    const store = useSipStore();
    await store.fetchSips(1);

    expect(store.sips).toEqual(mockSips.items);
    expect(store.page).toEqual(mockSips.page);
  });

  describe("download", () => {
    let originalOpen: typeof window.open;

    beforeEach(() => {
      originalOpen = window.open;
      window.open = vi.fn();
      vi.useFakeTimers();
    });

    afterEach(() => {
      window.open = originalOpen;
      vi.useRealTimers();
      vi.clearAllMocks();
    });

    it("downloads in a new window", async () => {
      const store = useSipStore();
      store.$patch({ current: { uuid: "sip-uuid" } });
      client.ingest.ingestDownloadSipRequest = vi.fn().mockResolvedValue({});

      await store.download();
      expect(window.open).toHaveBeenCalledWith(
        expect.stringContaining("/ingest/sips/sip-uuid/download"),
        "_blank",
      );
      expect(store.downloadError).toBeNull();
    });

    it("does nothing if current is null", async () => {
      const store = useSipStore();
      store.$patch({ current: null });
      await store.download();
      expect(window.open).not.toHaveBeenCalled();
    });

    it("sets downloadError if the request fail", async () => {
      const store = useSipStore();
      store.$patch({ current: { uuid: "sip-uuid" } });
      const errorMsg = "Download not allowed";
      client.ingest.ingestDownloadSipRequest = vi
        .fn()
        .mockRejectedValue(
          new ResponseError(
            new Response(JSON.stringify({ message: errorMsg })),
            "API error",
          ),
        );

      await store.download();
      expect(store.downloadError).toBe(errorMsg);

      vi.advanceTimersByTime(5000);
      expect(store.downloadError).toBeNull();
    });

    it("sets downloadError if there is an unexpected error", async () => {
      const store = useSipStore();
      store.$patch({ current: { uuid: "sip-uuid" } });
      client.ingest.ingestDownloadSipRequest = vi
        .fn()
        .mockRejectedValue(new Error("unexpected error"));

      await store.download();
      expect(store.downloadError).toBe("Unexpected error downloading package");

      vi.advanceTimersByTime(5000);
      expect(store.downloadError).toBeNull();
    });
  });
});
