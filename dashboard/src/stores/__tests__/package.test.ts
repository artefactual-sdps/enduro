import { usePackageStore } from "@/stores/package";
import { setActivePinia, createPinia } from "pinia";
import { expect, describe, it, beforeEach } from "vitest";
import { api } from "@/client";

describe("usePackageStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("getActionById finds actions", () => {
    const packageStore = usePackageStore();
    const now = new Date();
    packageStore.$patch((state) => {
      state.current_preservation_actions = {
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
      };
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
    packageStore.$patch((state) => {
      state.current_preservation_actions = {
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
      };
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
});
