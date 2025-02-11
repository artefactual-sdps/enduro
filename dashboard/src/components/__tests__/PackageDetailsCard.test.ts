import { createTestingPinia } from "@pinia/testing";
import { cleanup, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

import { api } from "@/client";
import PackageDetailsCard from "@/components/PackageDetailsCard.vue";
import { usePackageStore } from "@/stores/package";

describe("PackageDetailsCard.vue", () => {
  afterEach(() => cleanup());

  it("watches download requests from the store", async () => {
    render(PackageDetailsCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              package: {
                current: {
                  aipId: "89229d18-5554-4e0d-8c4e-d0d88afd3bae",
                  status: api.EnduroStoredPackageStatusEnum.Pending,
                } as api.EnduroStoredPackage,
              },
            },
          }),
        ],
      },
    });

    vi.stubGlobal("open", vi.fn());

    // Someone requests the download of the AIP via the package store.
    const packageStore = usePackageStore();
    packageStore.ui.download.request();
    await nextTick();

    // Then we observe that the component download function is executed.
    expect(window.open).toBeCalledWith(
      "http://localhost:3000/api/storage/package/89229d18-5554-4e0d-8c4e-d0d88afd3bae/download",
      "_blank",
    );
  });

  it("renders when the package is in pending status", async () => {
    const now = new Date();
    const { getByText } = render(PackageDetailsCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              package: {
                current: {} as api.EnduroStoredPackage,
                current_preservation_actions: {
                  actions: [
                    {
                      status:
                        api.EnduroPackagePreservationActionStatusEnum.Pending,
                      type: api.EnduroPackagePreservationActionTypeEnum
                        .MovePackage,
                    },
                  ],
                } as api.EnduroPackagePreservationActions,
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
              package: {
                current: {
                  aipId: "89229d18-5554-4e0d-8c4e-d0d88afd3bae",
                  status: api.EnduroStoredPackageStatusEnum.Done,
                } as api.EnduroStoredPackage,
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
              package: {
                current: {
                  aipId: "89229d18-5554-4e0d-8c4e-d0d88afd3bae",
                  status: api.EnduroStoredPackageStatusEnum.Done,
                } as api.EnduroStoredPackage,
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
