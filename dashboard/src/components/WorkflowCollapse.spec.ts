import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "@/client";
import WorkflowCollapse from "@/components/WorkflowCollapse.vue";
import { useSipStore } from "@/stores/sip";

const filters = {
  formatDateTime: () => "2026-04-14 00:00:00",
  formatDuration: () => "1 minute",
  getWorkflowLabel: (value: string) => value,
};

const ingestWorkflow = (
  overrides: Partial<api.EnduroIngestSipWorkflow> = {},
): api.EnduroIngestSipWorkflow => ({
  sipUuid: "sip-uuid",
  startedAt: new Date("2026-04-14T00:00:00Z"),
  status: api.EnduroIngestSipWorkflowStatusEnum.Pending,
  tasks: [
    {
      name: "Task 1",
      startedAt: new Date("2026-04-14T00:00:00Z"),
      status: api.EnduroIngestSipTaskStatusEnum.Done,
      uuid: "task-uuid",
      workflowUuid: "workflow-uuid",
    },
  ],
  temporalId: "temporal-id",
  type: api.EnduroIngestSipWorkflowTypeEnum.CreateAip,
  uuid: "workflow-uuid",
  ...overrides,
});

const useSipState = (
  overrides: Record<string, unknown> = {},
): Record<string, unknown> => ({
  current: {
    status: api.EnduroIngestSipStatusEnum.Pending,
    uuid: "sip-uuid",
  },
  currentDecision: null,
  ...overrides,
});

const renderWorkflow = (
  workflow: api.EnduroIngestSipWorkflow,
  sip: Record<string, unknown> = {},
) => {
  return render(WorkflowCollapse, {
    props: {
      workflow,
      index: 0,
      of: 2,
    },
    global: {
      mocks: {
        $filters: filters,
      },
      plugins: [
        createTestingPinia({
          createSpy: vi.fn,
          initialState: {
            auth: {
              attributes: ["ingest:sips:decision", "ingest:sips:review"],
              config: { enabled: true, abac: { enabled: true } },
            },
            sip: useSipState(sip),
          },
        }),
      ],
    },
  });
};

describe("WorkflowCollapse.vue", () => {
  afterEach(() => cleanup());

  it("shows the SIP review alert for pending review workflows", () => {
    const { container, getByRole, queryByText } = renderWorkflow(
      ingestWorkflow({
        type: api.EnduroIngestSipWorkflowTypeEnum.CreateAndReviewAip,
      }),
    );

    expect(getByRole("alert").textContent).toContain("Task: Review AIP");
    expect(queryByText("Task: User decision required")).toBeNull();
    expect(container.querySelector("#wf0-tasks")?.classList).toContain("show");
  });

  it("shows the SIP decision alert when a child decision is pending", () => {
    const { container, getByLabelText, getByRole, getByText, queryByText } =
      renderWorkflow(ingestWorkflow(), {
        currentDecision: {
          message: "Choose how to continue",
          options: ["Continue", "Cancel"],
        },
      });

    expect(getByRole("alert").textContent).toContain(
      "Task: User decision required",
    );
    getByText("Choose how to continue");
    expect((getByLabelText("Continue") as HTMLInputElement).checked).toBe(
      false,
    );
    expect(queryByText("Task: Review AIP")).toBeNull();
    expect(container.querySelector("#wf0-tasks")?.classList).toContain("show");
  });

  it("fetches the current decision when an ingest workflow becomes pending", async () => {
    const workflow = ingestWorkflow({
      status: api.EnduroIngestSipWorkflowStatusEnum.InProgress,
    });
    const { rerender } = renderWorkflow(workflow);
    const sipStore = useSipStore();
    sipStore.fetchCurrentDecision = vi.fn(() => Promise.resolve());

    await rerender({
      workflow: {
        ...workflow,
        status: api.EnduroIngestSipWorkflowStatusEnum.Pending,
      },
      index: 0,
      of: 2,
    });

    expect(sipStore.fetchCurrentDecision).toHaveBeenCalledWith("sip-uuid");
  });

  it("refreshes the current decision when an ingest workflow leaves pending", async () => {
    const workflow = ingestWorkflow();
    const { rerender } = renderWorkflow(workflow);
    const sipStore = useSipStore();
    sipStore.fetchCurrentDecision = vi.fn(() => Promise.resolve());

    await rerender({
      workflow: {
        ...workflow,
        status: api.EnduroIngestSipWorkflowStatusEnum.InProgress,
      },
      index: 0,
      of: 2,
    });

    expect(sipStore.fetchCurrentDecision).toHaveBeenCalledWith("sip-uuid");
  });

  it("keeps expanded note state with its task after a live update", async () => {
    const firstTask = {
      ...ingestWorkflow().tasks?.[0],
      note: "First line\nFirst task details",
    } as api.EnduroIngestSipTask;
    const workflow = ingestWorkflow({
      status: api.EnduroIngestSipWorkflowStatusEnum.InProgress,
      tasks: [firstTask],
    });
    const { container, rerender } = renderWorkflow(workflow);

    await fireEvent.click(
      container.querySelector("#pt-task-uuid-note-toggle") as HTMLElement,
    );

    await rerender({
      workflow: {
        ...workflow,
        tasks: [
          firstTask,
          {
            ...firstTask,
            name: "Task 2",
            note: "Second line\nSecond task details",
            uuid: "second-task-uuid",
          },
        ],
      },
      index: 0,
      of: 2,
    });

    expect(
      (container.querySelector("#pt-task-uuid-note-more") as HTMLElement).style
        .display,
    ).not.toBe("none");
    expect(
      (container.querySelector("#pt-second-task-uuid-note-more") as HTMLElement)
        .style.display,
    ).toBe("none");
  });
});
