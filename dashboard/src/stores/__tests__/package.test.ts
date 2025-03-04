import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it } from "vitest";

import { api } from "@/client";
import type { Pager } from "@/stores/ingest";
import { useIngestStore } from "@/stores/ingest";

describe("useingestStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("isPending", () => {
    const ingestStore = useIngestStore();
    const now = new Date();

    expect(ingestStore.isPending).toEqual(false);

    ingestStore.$patch({
      currentSip: {
        createdAt: now,
        id: 1,
        status: api.EnduroIngestSipStatusEnum.Pending,
      },
    });
    expect(ingestStore.isPending).toEqual(true);
  });

  it("isDone", () => {
    const ingestStore = useIngestStore();
    const now = new Date();

    expect(ingestStore.isDone).toEqual(false);

    ingestStore.$patch({
      currentSip: {
        createdAt: now,
        id: 1,
        status: api.EnduroIngestSipStatusEnum.Done,
      },
    });
    expect(ingestStore.isDone).toEqual(true);
  });

  it("isMovable", () => {
    const ingestStore = useIngestStore();
    const now = new Date();

    expect(ingestStore.isMovable).toEqual(false);

    ingestStore.$patch({
      currentSip: {
        createdAt: now,
        id: 1,
        status: api.EnduroIngestSipStatusEnum.Done,
      },
    });
    expect(ingestStore.isMovable).toEqual(true);
  });

  it("isMoving", () => {
    const ingestStore = useIngestStore();

    expect(ingestStore.isMoving).toEqual(false);

    ingestStore.$patch({ locationChanging: true });
    expect(ingestStore.isMoving).toEqual(true);
  });

  it("isRejected", () => {
    const ingestStore = useIngestStore();
    const now = new Date();

    expect(ingestStore.isRejected).toEqual(false);

    ingestStore.$patch({
      currentSip: {
        createdAt: now,
        id: 1,
        status: api.EnduroIngestSipStatusEnum.Done,
      },
    });
    expect(ingestStore.isRejected).toEqual(true);
  });

  it("hasNextPage", () => {
    const ingestStore = useIngestStore();

    ingestStore.$patch({
      page: { limit: 20, offset: 0, total: 20 },
    });
    expect(ingestStore.hasNextPage).toEqual(false);

    ingestStore.$patch({
      page: { limit: 20, offset: 20, total: 40 },
    });
    expect(ingestStore.hasNextPage).toEqual(false);

    ingestStore.$patch({
      page: { limit: 20, offset: 0, total: 21 },
    });
    expect(ingestStore.hasNextPage).toEqual(true);
  });

  it("hasPrevPage", () => {
    const ingestStore = useIngestStore();

    ingestStore.$patch({
      page: { limit: 20, offset: 0, total: 40 },
    });
    expect(ingestStore.hasPrevPage).toEqual(false);

    ingestStore.$patch({
      page: { limit: 20, offset: 20, total: 40 },
    });
    expect(ingestStore.hasPrevPage).toEqual(true);
  });

  it("returns lastResultOnPage", () => {
    const ingestStore = useIngestStore();

    ingestStore.$patch({
      page: { limit: 20, offset: 0, total: 40 },
    });
    expect(ingestStore.lastResultOnPage).toEqual(20);

    ingestStore.$patch({
      page: { limit: 20, offset: 0, total: 7 },
    });
    expect(ingestStore.lastResultOnPage).toEqual(7);

    ingestStore.$patch({
      page: { limit: 20, offset: 20, total: 35 },
    });
    expect(ingestStore.lastResultOnPage).toEqual(35);
  });

  it("getActionById finds actions", () => {
    const ingestStore = useIngestStore();
    const now = new Date();
    ingestStore.$patch({
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

    expect(ingestStore.getActionById(1)).toEqual({
      id: 1,
      type: api.EnduroIngestSipPreservationActionTypeEnum.CreateAip,
      startedAt: now,
      status: api.EnduroIngestSipPreservationActionStatusEnum.Done,
      workflowId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
    });
    expect(ingestStore.getActionById(2)).toEqual({
      id: 2,
      type: api.EnduroIngestSipPreservationActionTypeEnum.MovePackage,
      startedAt: now,
      status: api.EnduroIngestSipPreservationActionStatusEnum.Done,
      workflowId: "051cf998-6f87-4461-8091-8561ebf479c4",
    });
    expect(ingestStore.getActionById(3)).toBeUndefined();
  });

  it("getTaskById finds tasks", () => {
    const ingestStore = useIngestStore();
    const now = new Date();
    ingestStore.$patch({
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

    expect(ingestStore.getTaskById(1, 1)).toEqual({
      id: 1,
      name: "Task 1",
      startedAt: now,
      status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
      taskId: "1",
    });
    expect(ingestStore.getTaskById(1, 2)).toEqual({
      id: 2,
      name: "Task 2",
      startedAt: now,
      status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
      taskId: "2",
    });
    expect(ingestStore.getTaskById(1, 3)).toBeUndefined();

    expect(ingestStore.getTaskById(2, 3)).toEqual({
      id: 3,
      name: "Task 3",
      startedAt: now,
      status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
      taskId: "3",
    });
    expect(ingestStore.getTaskById(2, 4)).toEqual({
      id: 4,
      name: "Task 4",
      startedAt: now,
      status: api.EnduroIngestSipPreservationTaskStatusEnum.Done,
      taskId: "4",
    });
    expect(ingestStore.getTaskById(2, 5)).toBeUndefined();
  });

  it("updates the pager", () => {
    const ingestStore = useIngestStore();

    ingestStore.$patch({
      page: { limit: 20, offset: 60, total: 125 },
    });
    ingestStore.updatePager();
    expect(ingestStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 4,
      first: 1,
      last: 7,
      total: 7,
      pages: [1, 2, 3, 4, 5, 6, 7],
    });

    ingestStore.$patch({
      page: { limit: 20, offset: 160, total: 573 },
    });
    ingestStore.updatePager();
    expect(ingestStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 9,
      first: 6,
      last: 12,
      total: 29,
      pages: [6, 7, 8, 9, 10, 11, 12],
    });

    ingestStore.$patch({
      page: { limit: 20, offset: 540, total: 573 },
    });
    ingestStore.updatePager();
    expect(ingestStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 28,
      first: 23,
      last: 29,
      total: 29,
      pages: [23, 24, 25, 26, 27, 28, 29],
    });
  });
});
