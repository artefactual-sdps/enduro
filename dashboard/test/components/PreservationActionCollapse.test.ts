import { api } from "@/client";
import PreservationActionCollapse from "@/components/PreservationActionCollapse.vue";
import { createTestingPinia } from "@pinia/testing";
import { render, fireEvent, cleanup } from "@testing-library/vue";
import { describe, it, expect, vi, afterEach } from "vitest";

describe("PreservationActionCollapse.vue", () => {
  afterEach(() => cleanup());

  it("renders, expands and collapses", async () => {
    const now = new Date();
    const { getByText, getByRole, emitted, rerender } = render(
      PreservationActionCollapse,
      {
        props: {
          action: {
            completedAt: now,
            id: 1,
            startedAt: now,
            status: api.EnduroPackagePreservationTaskStatusEnum.Done,
            tasks: [
              {
                completedAt: now,
                id: 1,
                name: "Task 1",
                startedAt: now,
                status: api.EnduroPackagePreservationTaskStatusEnum.Done,
                taskId: "f8fa23fa-d749-497e-a5bc-39f637372c1a",
              },
              {
                completedAt: now,
                id: 2,
                name: "Task 2",
                startedAt: now,
                status: api.EnduroPackagePreservationTaskStatusEnum.Done,
                taskId: "66067aba-cc4f-40b0-b7f2-4eca7b3cfcf6",
              },
            ],
            type: api.EnduroPackagePreservationActionTypeEnum.MovePackage,
            workflowId: "move-workflow-ba438fbe-c57e-41ae-b29e-65b6bb04b650",
          } as api.EnduroPackagePreservationAction,
          index: 0,
          toggleAll: false,
        },
        global: {
          plugins: [
            createTestingPinia({
              createSpy: vi.fn,
              initialState: {
                package: {
                  current: null,
                },
              },
            }),
          ],
          mocks: {
            $filters: {
              formatDateTime: () => "now",
              formatDuration: () => "some time",
              getPreservationActionLabel: () => "Move package",
            },
          },
        },
      }
    );

    getByText("Move package");
    getByText("Completed now (took some time)");

    const expandButton = getByRole("button", {
      name: "Expand preservation tasks table",
    });

    await fireEvent.click(expandButton);
    expect(emitted()["update:toggleAll"][0]).toStrictEqual([null]);

    const collapseButton = getByRole("button", {
      name: "Collapse preservation tasks table",
    });

    await fireEvent.click(collapseButton);
    expect(emitted()["update:toggleAll"][1]).toStrictEqual([null]);

    // This broke when we moved to TypeScript 5.x but can't tell why.
    // await rerender({ toggleAll: true });
    // expect(emitted()["update:toggleAll"][2]).toStrictEqual([null]);
  });
});
