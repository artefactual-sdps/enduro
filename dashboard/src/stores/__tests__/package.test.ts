import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it } from "vitest";

import { api } from "@/client";
import type { Pager } from "@/stores/package";
import { usePackageStore } from "@/stores/package";

describe("usePackageStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("isPending", () => {
    const packageStore = usePackageStore();
    const now = new Date();

    expect(packageStore.isPending).toEqual(false);

    packageStore.$patch({
      current: {
        createdAt: now,
        id: 1,
        status: api.EnduroStoredPackageStatusEnum.Pending,
      },
    });
    expect(packageStore.isPending).toEqual(true);
  });

  it("isDone", () => {
    const packageStore = usePackageStore();
    const now = new Date();

    expect(packageStore.isDone).toEqual(false);

    packageStore.$patch({
      current: {
        createdAt: now,
        id: 1,
        status: api.EnduroStoredPackageStatusEnum.Done,
      },
    });
    expect(packageStore.isDone).toEqual(true);
  });

  it("isMovable", () => {
    const packageStore = usePackageStore();
    const now = new Date();

    expect(packageStore.isMovable).toEqual(false);

    packageStore.$patch({
      current: {
        createdAt: now,
        id: 1,
        status: api.EnduroStoredPackageStatusEnum.Done,
      },
    });
    expect(packageStore.isMovable).toEqual(true);
  });

  it("isMoving", () => {
    const packageStore = usePackageStore();

    expect(packageStore.isMoving).toEqual(false);

    packageStore.$patch({ locationChanging: true });
    expect(packageStore.isMoving).toEqual(true);
  });

  it("isRejected", () => {
    const packageStore = usePackageStore();
    const now = new Date();

    expect(packageStore.isRejected).toEqual(false);

    packageStore.$patch({
      current: {
        createdAt: now,
        id: 1,
        status: api.EnduroStoredPackageStatusEnum.Done,
      },
    });
    expect(packageStore.isRejected).toEqual(true);
  });

  it("hasNextPage", () => {
    const packageStore = usePackageStore();

    packageStore.$patch({
      page: { limit: 20, offset: 0, total: 20 },
    });
    expect(packageStore.hasNextPage).toEqual(false);

    packageStore.$patch({
      page: { limit: 20, offset: 20, total: 40 },
    });
    expect(packageStore.hasNextPage).toEqual(false);

    packageStore.$patch({
      page: { limit: 20, offset: 0, total: 21 },
    });
    expect(packageStore.hasNextPage).toEqual(true);
  });

  it("hasPrevPage", () => {
    const packageStore = usePackageStore();

    packageStore.$patch({
      page: { limit: 20, offset: 0, total: 40 },
    });
    expect(packageStore.hasPrevPage).toEqual(false);

    packageStore.$patch({
      page: { limit: 20, offset: 20, total: 40 },
    });
    expect(packageStore.hasPrevPage).toEqual(true);
  });

  it("returns lastResultOnPage", () => {
    const packageStore = usePackageStore();

    packageStore.$patch({
      page: { limit: 20, offset: 0, total: 40 },
    });
    expect(packageStore.lastResultOnPage).toEqual(20);

    packageStore.$patch({
      page: { limit: 20, offset: 0, total: 7 },
    });
    expect(packageStore.lastResultOnPage).toEqual(7);

    packageStore.$patch({
      page: { limit: 20, offset: 20, total: 35 },
    });
    expect(packageStore.lastResultOnPage).toEqual(35);
  });

  it("getActionById finds actions", () => {
    const packageStore = usePackageStore();
    const now = new Date();
    packageStore.$patch({
      current_preservation_actions: {
        actions: [
          {
            id: 1,
            type: api.EnduroPackagePreservationActionTypeEnum.CreateAip,
            startedAt: now,
            status: api.EnduroPackagePreservationActionStatusEnum.Done,
            workflowId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
          },
          {
            id: 2,
            type: api.EnduroPackagePreservationActionTypeEnum.MovePackage,
            startedAt: now,
            status: api.EnduroPackagePreservationActionStatusEnum.Done,
            workflowId: "051cf998-6f87-4461-8091-8561ebf479c4",
          },
        ],
      },
    });

    expect(packageStore.getActionById(1)).toEqual({
      id: 1,
      type: api.EnduroPackagePreservationActionTypeEnum.CreateAip,
      startedAt: now,
      status: api.EnduroPackagePreservationActionStatusEnum.Done,
      workflowId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
    });
    expect(packageStore.getActionById(2)).toEqual({
      id: 2,
      type: api.EnduroPackagePreservationActionTypeEnum.MovePackage,
      startedAt: now,
      status: api.EnduroPackagePreservationActionStatusEnum.Done,
      workflowId: "051cf998-6f87-4461-8091-8561ebf479c4",
    });
    expect(packageStore.getActionById(3)).toBeUndefined();
  });

  it("getTaskById finds tasks", () => {
    const packageStore = usePackageStore();
    const now = new Date();
    packageStore.$patch({
      current_preservation_actions: {
        actions: [
          {
            id: 1,
            type: api.EnduroPackagePreservationActionTypeEnum.CreateAip,
            startedAt: now,
            status: api.EnduroPackagePreservationActionStatusEnum.Done,
            workflowId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
            tasks: [
              {
                id: 1,
                name: "Task 1",
                startedAt: now,
                status: api.EnduroPackagePreservationTaskStatusEnum.Done,
                taskId: "1",
              },
              {
                id: 2,
                name: "Task 2",
                startedAt: now,
                status: api.EnduroPackagePreservationTaskStatusEnum.Done,
                taskId: "2",
              },
            ],
          },
          {
            id: 2,
            type: api.EnduroPackagePreservationActionTypeEnum.MovePackage,
            startedAt: now,
            status: api.EnduroPackagePreservationActionStatusEnum.Done,
            workflowId: "051cf998-6f87-4461-8091-8561ebf479c4",
            tasks: [
              {
                id: 3,
                name: "Task 3",
                startedAt: now,
                status: api.EnduroPackagePreservationTaskStatusEnum.Done,
                taskId: "3",
              },
              {
                id: 4,
                name: "Task 4",
                startedAt: now,
                status: api.EnduroPackagePreservationTaskStatusEnum.Done,
                taskId: "4",
              },
            ],
          },
        ],
      },
    });

    expect(packageStore.getTaskById(1, 1)).toEqual({
      id: 1,
      name: "Task 1",
      startedAt: now,
      status: api.EnduroPackagePreservationTaskStatusEnum.Done,
      taskId: "1",
    });
    expect(packageStore.getTaskById(1, 2)).toEqual({
      id: 2,
      name: "Task 2",
      startedAt: now,
      status: api.EnduroPackagePreservationTaskStatusEnum.Done,
      taskId: "2",
    });
    expect(packageStore.getTaskById(1, 3)).toBeUndefined();

    expect(packageStore.getTaskById(2, 3)).toEqual({
      id: 3,
      name: "Task 3",
      startedAt: now,
      status: api.EnduroPackagePreservationTaskStatusEnum.Done,
      taskId: "3",
    });
    expect(packageStore.getTaskById(2, 4)).toEqual({
      id: 4,
      name: "Task 4",
      startedAt: now,
      status: api.EnduroPackagePreservationTaskStatusEnum.Done,
      taskId: "4",
    });
    expect(packageStore.getTaskById(2, 5)).toBeUndefined();
  });

  it("updates the pager", () => {
    const packageStore = usePackageStore();

    packageStore.$patch({
      page: { limit: 20, offset: 60, total: 125 },
    });
    packageStore.updatePager();
    expect(packageStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 4,
      first: 1,
      last: 7,
      total: 7,
      pages: [1, 2, 3, 4, 5, 6, 7],
    });

    packageStore.$patch({
      page: { limit: 20, offset: 160, total: 573 },
    });
    packageStore.updatePager();
    expect(packageStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 9,
      first: 6,
      last: 12,
      total: 29,
      pages: [6, 7, 8, 9, 10, 11, 12],
    });

    packageStore.$patch({
      page: { limit: 20, offset: 540, total: 573 },
    });
    packageStore.updatePager();
    expect(packageStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 28,
      first: 23,
      last: 29,
      total: 29,
      pages: [23, 24, 25, 26, 27, 28, 29],
    });
  });
});
