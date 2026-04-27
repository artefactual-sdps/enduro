import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "@/client";
import BatchReviewAlert from "@/components/BatchReviewAlert.vue";
import { useBatchStore } from "@/stores/batch";

const openDialogMock = vi.hoisted(() => vi.fn());
vi.mock("vue3-promise-dialog", () => ({ openDialog: openDialogMock }));

describe("BatchReviewAlert.vue", () => {
  afterEach(() => {
    cleanup();
    vi.resetAllMocks();
  });

  it("renders nothing when the batch is not pending", () => {
    const { queryByText } = render(BatchReviewAlert, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              batch: { current: null },
            },
          }),
        ],
      },
    });

    expect(queryByText("Review batch")).toBeNull();
  });

  it("shows buttons and submits the review", async () => {
    openDialogMock.mockResolvedValue(true);
    const { getByRole, getByText } = render(BatchReviewAlert, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              batch: {
                current: {
                  status: api.EnduroIngestBatchStatusEnum.Pending,
                  identifier: "Batch 101",
                },
              },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: ["ingest:batches:review"],
              },
            },
          }),
        ],
      },
    });

    const batchStore = useBatchStore();

    getByText("Review batch");

    await fireEvent.click(
      getByRole("button", { name: "Continue partial ingest" }),
    );
    expect(batchStore.reviewBatch).toHaveBeenCalledWith(true);

    await fireEvent.click(getByRole("button", { name: "Cancel batch" }));
    expect(batchStore.reviewBatch).toHaveBeenCalledWith(false);
  });

  it("does not submit the review when confirmation is cancelled", async () => {
    openDialogMock.mockResolvedValue(false);
    const { getByRole } = render(BatchReviewAlert, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              batch: {
                current: {
                  status: api.EnduroIngestBatchStatusEnum.Pending,
                },
              },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: ["ingest:batches:review"],
              },
            },
          }),
        ],
      },
    });

    const batchStore = useBatchStore();

    await fireEvent.click(
      getByRole("button", { name: "Continue partial ingest" }),
    );
    expect(batchStore.reviewBatch).not.toHaveBeenCalled();

    await fireEvent.click(getByRole("button", { name: "Cancel batch" }));
    expect(batchStore.reviewBatch).not.toHaveBeenCalled();
  });

  it("shows the alert without review buttons when permissions are missing", () => {
    const { getByText, queryByRole } = render(BatchReviewAlert, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              batch: {
                current: {
                  status: api.EnduroIngestBatchStatusEnum.Pending,
                },
              },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: [],
              },
            },
          }),
        ],
      },
    });

    getByText("Review batch");
    expect(queryByRole("button", { name: "Continue" })).toBeNull();
    expect(queryByRole("button", { name: "Cancel" })).toBeNull();
  });
});
