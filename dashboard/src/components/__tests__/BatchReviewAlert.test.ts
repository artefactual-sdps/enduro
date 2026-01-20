import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "@/client";
import BatchReviewAlert from "@/components/BatchReviewAlert.vue";
import { useBatchStore } from "@/stores/batch";

describe("BatchReviewAlert.vue", () => {
  afterEach(() => cleanup());

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
    const { getByRole, getByText } = render(BatchReviewAlert, {
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

    getByText("Review batch");

    await fireEvent.click(getByRole("button", { name: "Continue" }));
    await fireEvent.click(getByRole("button", { name: "Cancel" }));

    expect(batchStore.reviewBatch).toHaveBeenCalledWith(true);
    expect(batchStore.reviewBatch).toHaveBeenCalledWith(false);
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
