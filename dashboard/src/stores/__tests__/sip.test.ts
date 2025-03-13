import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, client } from "@/client";
import { useLayoutStore } from "@/stores/layout";
import type { Pager } from "@/stores/sip";
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

  it("hasNextPage", () => {
    const sipStore = useSipStore();

    sipStore.$patch({
      page: { limit: 20, offset: 0, total: 20 },
    });
    expect(sipStore.hasNextPage).toEqual(false);

    sipStore.$patch({
      page: { limit: 20, offset: 20, total: 40 },
    });
    expect(sipStore.hasNextPage).toEqual(false);

    sipStore.$patch({
      page: { limit: 20, offset: 0, total: 21 },
    });
    expect(sipStore.hasNextPage).toEqual(true);
  });

  it("hasPrevPage", () => {
    const sipStore = useSipStore();

    sipStore.$patch({
      page: { limit: 20, offset: 0, total: 40 },
    });
    expect(sipStore.hasPrevPage).toEqual(false);

    sipStore.$patch({
      page: { limit: 20, offset: 20, total: 40 },
    });
    expect(sipStore.hasPrevPage).toEqual(true);
  });

  it("returns lastResultOnPage", () => {
    const sipStore = useSipStore();

    sipStore.$patch({
      page: { limit: 20, offset: 0, total: 40 },
    });
    expect(sipStore.lastResultOnPage).toEqual(20);

    sipStore.$patch({
      page: { limit: 20, offset: 0, total: 7 },
    });
    expect(sipStore.lastResultOnPage).toEqual(7);

    sipStore.$patch({
      page: { limit: 20, offset: 20, total: 35 },
    });
    expect(sipStore.lastResultOnPage).toEqual(35);
  });

  it("getActionById finds actions", () => {
    const sipStore = useSipStore();
    const now = new Date();
    sipStore.$patch({
      currentPreservationActions: {
        actions: [
          {
            id: 1,
            type: api.EnduroIngestSipPreservationActionTypeEnum.CreateAip,
            startedAt: now,
            status: api.EnduroIngestSipPreservationActionStatusEnum.Done,
            workflowId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
          },
          {
            id: 2,
            type: api.EnduroIngestSipPreservationActionTypeEnum.MovePackage,
            startedAt: now,
            status: api.EnduroIngestSipPreservationActionStatusEnum.Done,
            workflowId: "051cf998-6f87-4461-8091-8561ebf479c4",
          },
        ],
      },
    });

    expect(sipStore.getActionById(1)).toEqual({
      id: 1,
      type: api.EnduroIngestSipPreservationActionTypeEnum.CreateAip,
      startedAt: now,
      status: api.EnduroIngestSipPreservationActionStatusEnum.Done,
      workflowId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
    });
    expect(sipStore.getActionById(2)).toEqual({
      id: 2,
      type: api.EnduroIngestSipPreservationActionTypeEnum.MovePackage,
      startedAt: now,
      status: api.EnduroIngestSipPreservationActionStatusEnum.Done,
      workflowId: "051cf998-6f87-4461-8091-8561ebf479c4",
    });
    expect(sipStore.getActionById(3)).toBeUndefined();
  });

  it("getTaskById finds tasks", () => {
    const sipStore = useSipStore();
    const now = new Date();
    sipStore.$patch({
      currentPreservationActions: {
        actions: [
          {
            id: 1,
            type: api.EnduroIngestSipPreservationActionTypeEnum.CreateAip,
            startedAt: now,
            status: api.EnduroIngestSipPreservationActionStatusEnum.Done,
            workflowId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
            tasks: [
              {
                id: 1,
                name: "Task 1",
                startedAt: now,
                status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
                taskId: "1",
              },
              {
                id: 2,
                name: "Task 2",
                startedAt: now,
                status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
                taskId: "2",
              },
            ],
          },
          {
            id: 2,
            type: api.EnduroIngestSipPreservationActionTypeEnum.MovePackage,
            startedAt: now,
            status: api.EnduroIngestSipPreservationActionStatusEnum.Done,
            workflowId: "051cf998-6f87-4461-8091-8561ebf479c4",
            tasks: [
              {
                id: 3,
                name: "Task 3",
                startedAt: now,
                status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
                taskId: "3",
              },
              {
                id: 4,
                name: "Task 4",
                startedAt: now,
                status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
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
      status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
      taskId: "1",
    });
    expect(sipStore.getTaskById(1, 2)).toEqual({
      id: 2,
      name: "Task 2",
      startedAt: now,
      status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
      taskId: "2",
    });
    expect(sipStore.getTaskById(1, 3)).toBeUndefined();

    expect(sipStore.getTaskById(2, 3)).toEqual({
      id: 3,
      name: "Task 3",
      startedAt: now,
      status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
      taskId: "3",
    });
    expect(sipStore.getTaskById(2, 4)).toEqual({
      id: 4,
      name: "Task 4",
      startedAt: now,
      status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
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
    const mockPreservationActions: api.SIPPreservationActions = {
      actions: [
        {
          id: 1,
          startedAt: new Date("2025-01-01T00:00:00Z"),
          status: api.EnduroIngestSipPreservationActionStatusEnum.Done,
          type: api.EnduroIngestSipPreservationActionTypeEnum.CreateAip,
          workflowId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
        },
      ],
    };

    client.ingest.ingestShowSip = vi.fn().mockResolvedValue(mockSip);
    client.ingest.ingestListSipPreservationActions = vi
      .fn()
      .mockResolvedValue(mockPreservationActions);

    const store = useSipStore();
    await store.fetchCurrent("1");

    expect(store.current).toEqual(mockSip);
    expect(store.currentPreservationActions).toEqual(mockPreservationActions);

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

  it("updates the pager", () => {
    const sipStore = useSipStore();

    sipStore.$patch({
      page: { limit: 20, offset: 60, total: 125 },
    });
    sipStore.updatePager();
    expect(sipStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 4,
      first: 1,
      last: 7,
      total: 7,
      pages: [1, 2, 3, 4, 5, 6, 7],
    });

    sipStore.$patch({
      page: { limit: 20, offset: 160, total: 573 },
    });
    sipStore.updatePager();
    expect(sipStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 9,
      first: 6,
      last: 12,
      total: 29,
      pages: [6, 7, 8, 9, 10, 11, 12],
    });

    sipStore.$patch({
      page: { limit: 20, offset: 540, total: 573 },
    });
    sipStore.updatePager();
    expect(sipStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 28,
      first: 23,
      last: 29,
      total: 29,
      pages: [23, 24, 25, 26, 27, 28, 29],
    });
  });
});
