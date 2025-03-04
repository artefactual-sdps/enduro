import { createTestingPinia } from "@pinia/testing";
import { cleanup, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

import { api } from "@/client";
import PackageDetailsCard from "@/components/PackageDetailsCard.vue";
import { useIngestStore } from "@/stores/ingest";

describe("PackageDetailsCard.vue", () => {
  afterEach(() => cleanup());

  it("watches download requests from the store", async () => {
    render(PackageDetailsCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              ingest: {
                currentSip: {
                  aipId: "89229d18-5554-4e0d-8c4e-d0d88afd3bae",
                  status: api.EnduroIngestSipStatusEnum.Pending,
                } as api.EnduroIngestSip,
              },
            },
          }),
        ],
      },
    });

    vi.stubGlobal("open", vi.fn());

    // Someone requests the download of the AIP via the ingest store.
    const ingestStore = useIngestStore();
    ingestStore.ui.download.request();
    await nextTick();

    // Then we observe that the component download function is executed.
    expect(window.open).toBeCalledWith(
      "http://localhost:3000/api/storage/aips/89229d18-5554-4e0d-8c4e-d0d88afd3bae/download",
      "_blank",
    );
  });

  it("renders when the package is in pending status", async () => {
    const { getByText } = render(PackageDetailsCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              ingest: {
                currentSip: {} as api.EnduroIngestSip,
                currentPreservationActions: {
                  actions: [
                    {
                      status:
                        api.EnduroIngestSipPreservationActionStatusEnum.Pending,
                      type: api.EnduroIngestSipPreservationActionTypeEnum
                        .MovePackage,
                    },
                  ],
                } as api.SIPPreservationActions,
              },
            },
          }),
        ],
        mocks: {
          $filters: {
            getPreservationActionLabel: () => "Move package",
          },
        },
      },
    });

    getByText("PENDING");
    getByText("(Move package)");
  });

  it("shows the download button", async () => {
    const { getByRole } = render(PackageDetailsCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              ingest: {
                currentSip: {
                  aipId: "89229d18-5554-4e0d-8c4e-d0d88afd3bae",
                  status: api.EnduroIngestSipStatusEnum.Done,
                } as api.EnduroIngestSip,
              },
            },
          }),
        ],
      },
    });

    getByRole("button", { name: "Download" });
  });

  it("hides the download button", async () => {
    const { queryByRole } = render(PackageDetailsCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              ingest: {
                currentSip: {
                  aipId: "89229d18-5554-4e0d-8c4e-d0d88afd3bae",
                  status: api.EnduroIngestSipStatusEnum.Done,
                } as api.EnduroIngestSip,
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

    expect(queryByRole("button", { name: "Download" })).toBeNull();
  });
});
