import { api } from "@/client";
import PackageDetailsCard from "@/components/PackageDetailsCard.vue";
import { usePackageStore } from "@/stores/package";
import { createTestingPinia } from "@pinia/testing";
import { render, cleanup } from "@testing-library/vue";
import { expect, describe, it, vi, afterEach } from "vitest";
import { nextTick } from "vue";

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
      "///api/storage/package/89229d18-5554-4e0d-8c4e-d0d88afd3bae/download",
      "_blank"
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
});
