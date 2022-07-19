import { api } from "../../src/client";
import PreservationActionCollapse from "../../src/components/PreservationActionCollapse.vue";
import { createTestingPinia } from "@pinia/testing";
import { render, fireEvent } from "@testing-library/vue";
import { describe, it, expect, vi } from "vitest";
import { nextTick } from "vue";

describe("PreservationActionCollapse.vue", () => {
  it("renders, expands and collapses", async () => {
    const now = new Date();
    const { getByText, getByRole, emitted, unmount, rerender } = render(
      PreservationActionCollapse,
      {
        props: {
          action: {
            completedAt: now,
            id: 1,
            startedAt: now,
            status:
              api.EnduroPackagePreservationTaskResponseBodyStatusEnum.Done,
            tasks: [
              {
                completedAt: now,
                id: 1,
                name: "Task 1",
                startedAt: now,
                status:
                  api.EnduroPackagePreservationTaskResponseBodyStatusEnum.Done,
                taskId: "f8fa23fa-d749-497e-a5bc-39f637372c1a",
              },
              {
                completedAt: now,
                id: 2,
                name: "Task 2",
                startedAt: now,
                status:
                  api.EnduroPackagePreservationTaskResponseBodyStatusEnum.Done,
                taskId: "66067aba-cc4f-40b0-b7f2-4eca7b3cfcf6",
              },
            ] as api.EnduroPackagePreservationTaskResponseBody[],
            type: api.EnduroPackagePreservationActionResponseBodyTypeEnum
              .MovePackage,
            workflowId: "move-workflow-ba438fbe-c57e-41ae-b29e-65b6bb04b650",
          } as api.EnduroPackagePreservationActionResponseBody,
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
              formatPreservationActionStatus: () => "",
              formatPreservationTaskStatus: () => "",
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

    await rerender({ toggleAll: true });
    expect(emitted()["update:toggleAll"][2]).toStrictEqual([null]);

    unmount();
  });
});
