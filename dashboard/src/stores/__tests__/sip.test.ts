import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, client } from "@/client";
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
        id: 1,
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
            id: 1,
            type: api.EnduroIngestSipWorkflowTypeEnum.CreateAip,
            startedAt: now,
            status: api.EnduroIngestSipWorkflowStatusEnum.Done,
            temporalId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
          },
          {
            id: 2,
            type: api.EnduroIngestSipWorkflowTypeEnum.CreateAndReviewAip,
            startedAt: now,
            status: api.EnduroIngestSipWorkflowStatusEnum.Done,
            temporalId: "051cf998-6f87-4461-8091-8561ebf479c4",
          },
        ],
      },
    });

    expect(sipStore.getWorkflowById(1)).toEqual({
      id: 1,
      type: api.EnduroIngestSipWorkflowTypeEnum.CreateAip,
      startedAt: now,
      status: api.EnduroIngestSipWorkflowStatusEnum.Done,
      temporalId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
    });
    expect(sipStore.getWorkflowById(2)).toEqual({
      id: 2,
      type: api.EnduroIngestSipWorkflowTypeEnum.CreateAndReviewAip,
      startedAt: now,
      status: api.EnduroIngestSipWorkflowStatusEnum.Done,
      temporalId: "051cf998-6f87-4461-8091-8561ebf479c4",
    });
    expect(sipStore.getWorkflowById(3)).toBeUndefined();
  });

  it("getTaskById finds tasks", () => {
    const sipStore = useSipStore();
    const now = new Date();
    sipStore.$patch({
      currentWorkflows: {
        workflows: [
          {
            id: 1,
            type: api.EnduroIngestSipWorkflowTypeEnum.CreateAip,
            startedAt: now,
            status: api.EnduroIngestSipWorkflowStatusEnum.Done,
            temporalId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
            tasks: [
              {
                id: 1,
                name: "Task 1",
                startedAt: now,
                status: api.EnduroIngestSipTaskStatusEnum.Done,
                taskId: "1",
              },
              {
                id: 2,
                name: "Task 2",
                startedAt: now,
                status: api.EnduroIngestSipTaskStatusEnum.Done,
                taskId: "2",
              },
            ],
          },
          {
            id: 2,
            type: api.EnduroIngestSipWorkflowTypeEnum.CreateAndReviewAip,
            startedAt: now,
            status: api.EnduroIngestSipWorkflowStatusEnum.Done,
            temporalId: "051cf998-6f87-4461-8091-8561ebf479c4",
            tasks: [
              {
                id: 3,
                name: "Task 3",
                startedAt: now,
                status: api.EnduroIngestSipTaskStatusEnum.Done,
                taskId: "3",
              },
              {
                id: 4,
                name: "Task 4",
                startedAt: now,
                status: api.EnduroIngestSipTaskStatusEnum.Done,
                taskId: "4",
              },
            ],
          },
        ],
      },
    });

    expect(sipStore.getTaskById(1, 1)).toEqual({
      id: 1,
      name: "Task 1",
      startedAt: now,
      status: api.EnduroIngestSipTaskStatusEnum.Done,
      taskId: "1",
    });
    expect(sipStore.getTaskById(1, 2)).toEqual({
      id: 2,
      name: "Task 2",
      startedAt: now,
      status: api.EnduroIngestSipTaskStatusEnum.Done,
      taskId: "2",
    });
    expect(sipStore.getTaskById(1, 3)).toBeUndefined();

    expect(sipStore.getTaskById(2, 3)).toEqual({
      id: 3,
      name: "Task 3",
      startedAt: now,
      status: api.EnduroIngestSipTaskStatusEnum.Done,
      taskId: "3",
    });
    expect(sipStore.getTaskById(2, 4)).toEqual({
      id: 4,
      name: "Task 4",
      startedAt: now,
      status: api.EnduroIngestSipTaskStatusEnum.Done,
      taskId: "4",
    });
    expect(sipStore.getTaskById(2, 5)).toBeUndefined();
  });

  it("fetches current", async () => {
    const mockSip: api.EnduroIngestSip = {
      id: 1,
      name: "SIP 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      status: api.EnduroIngestSipStatusEnum.Done,
    };
    const mockWorkflows: api.SIPWorkflows = {
      workflows: [
        {
          id: 1,
          startedAt: new Date("2025-01-01T00:00:00Z"),
          status: api.EnduroIngestSipWorkflowStatusEnum.Done,
          type: api.EnduroIngestSipWorkflowTypeEnum.CreateAip,
          temporalId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
        },
      ],
    };

    client.ingest.ingestShowSip = vi.fn().mockResolvedValue(mockSip);
    client.ingest.ingestListSipWorkflows = vi
      .fn()
      .mockResolvedValue(mockWorkflows);

    const store = useSipStore();
    await store.fetchCurrent("1");

    expect(store.current).toEqual(mockSip);
    expect(store.currentWorkflows).toEqual(mockWorkflows);

    const layoutStore = useLayoutStore();
    expect(layoutStore.breadcrumb).toEqual([
      { text: "Ingest" },
      { route: expect.any(Object), text: "SIPs" },
      { text: mockSip.name },
    ]);
  });

  it("fetches SIPs", async () => {
    const mockSips: api.SIPs = {
      items: [
        {
          id: 1,
          name: "SIP 1",
          createdAt: new Date("2025-01-01T00:00:00Z"),
          status: api.EnduroIngestSipStatusEnum.Done,
        },
        {
          id: 2,
          name: "SIP 2",
          createdAt: new Date("2025-01-02T00:00:00Z"),
          status: api.EnduroIngestSipStatusEnum.Done,
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
});
