import { api } from "@/client";
import PackageLocationCard from "@/components/PackageLocationCard.vue";
import { usePackageStore } from "@/stores/package";
import { createTestingPinia } from "@pinia/testing";
import { render, fireEvent, cleanup } from "@testing-library/vue";
import { describe, it, vi, expect, afterEach } from "vitest";

describe("PackageLocationCard.vue", () => {
  afterEach(() => cleanup());

  it("renders when the package is stored", async () => {
    const { html } = render(PackageLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              package: {
                current: {
                  status: api.EnduroStoredPackageStatusEnum.Done,
                  locationId: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroStoredPackage,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div class="card mb-3">
        <div class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 class="card-title">Location</h4>
          <p class="card-text"><span><div class="d-flex align-items-start gap-2"><span class="font-monospace">f8635e46-a320-4152-9a2c-98a28eeb50d1</span><button class="btn btn-sm btn-link link-secondary p-0" data-bs-toggle="tooltip" data-bs-title="Copy to clipboard">
              <!-- Copied visual hint. -->
              <!-- Copy icon. --><span><svg viewBox="0 0 24 24" width="1.2em" height="1.2em" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M8 4v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7.242a2 2 0 0 0-.602-1.43L16.083 2.57A2 2 0 0 0 14.685 2H10a2 2 0 0 0-2 2"></path><path d="M16 18v2a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2h2"></path></g></svg><span class="visually-hidden">Copy to clipboard</span></span>
            </button>
        </div></span></p>
        <!--v-if-->
      </div>
      </div>"
    `);
  });

  it("renders when the package location is moved", async () => {
    const { getByText } = render(PackageLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            stubActions: false,
            initialState: {
              package: {
                current: {
                  status: api.EnduroStoredPackageStatusEnum.Done,
                  locationId: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroStoredPackage,
              },
            },
          }),
        ],
      },
    });

    getByText("f8635e46-a320-4152-9a2c-98a28eeb50d1");

    const packageStore = usePackageStore();

    const moveMock = vi.fn().mockImplementation(packageStore.move);
    moveMock.mockImplementation(async () => {
      packageStore.$patch((state) => {
        if (!state.current) return;
        state.current.status = api.EnduroStoredPackageStatusEnum.InProgress;
        state.locationChanging = true;
      });
    });
    packageStore.move = moveMock;

    vi.mock("@/dialogs", () => {
      return {
        openPackageLocationDialog: () => "fe675e52-c761-46d0-8605-fae4bd10303e",
      };
    });

    const button = getByText("Choose storage location");
    await fireEvent.click(button);

    getByText("The package is being moved into a new location.");
  });

  it("renders when the package location is not available", async () => {
    const { html } = render(PackageLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              package: {
                current: {
                  status: api.EnduroStoredPackageStatusEnum.InProgress,
                } as api.EnduroStoredPackage,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div class="card mb-3">
        <div class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 class="card-title">Location</h4>
          <p class="card-text"><span>Not available yet.</span></p>
          <!--v-if-->
        </div>
      </div>"
    `);
  });

  it("renders when the package is rejected", async () => {
    const { html } = render(PackageLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              package: {
                current: {
                  status: api.EnduroStoredPackageStatusEnum.Done,
                  locationId: undefined,
                } as api.EnduroStoredPackage,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div class="card mb-3">
        <div class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 class="card-title">Location</h4>
          <p class="card-text"><span>Package rejected.</span></p>
          <!--v-if-->
        </div>
      </div>"
    `);
  });

  it("renders when the package is moving", async () => {
    const { html } = render(PackageLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              package: {
                current: {
                  status: api.EnduroStoredPackageStatusEnum.InProgress,
                  locationId: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroStoredPackage,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div class="card mb-3">
        <div class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 class="card-title">Location</h4>
          <p class="card-text"><span><div class="d-flex align-items-start gap-2"><span class="font-monospace">f8635e46-a320-4152-9a2c-98a28eeb50d1</span><button class="btn btn-sm btn-link link-secondary p-0" data-bs-toggle="tooltip" data-bs-title="Copy to clipboard">
              <!-- Copied visual hint. -->
              <!-- Copy icon. --><span><svg viewBox="0 0 24 24" width="1.2em" height="1.2em" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M8 4v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7.242a2 2 0 0 0-.602-1.43L16.083 2.57A2 2 0 0 0 14.685 2H10a2 2 0 0 0-2 2"></path><path d="M16 18v2a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2h2"></path></g></svg><span class="visually-hidden">Copy to clipboard</span></span>
            </button>
        </div></span></p>
        <!--v-if-->
      </div>
      </div>"
    `);
  });
});
