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
      "<div class=\\"card mb-3\\">
        <div class=\\"card-body\\">
          <!--v-if-->
          <!--v-if-->
          <h4 class=\\"card-title\\">Location</h4>
          <p class=\\"card-text\\"><span><div class=\\"d-flex align-items-start gap-2\\"><span class=\\"font-monospace\\">f8635e46-a320-4152-9a2c-98a28eeb50d1</span>
            <!--v-if-->
        </div></span></p>
        <div class=\\"actions\\"><button type=\\"button\\" class=\\"btn btn-primary btn-sm\\">Choose storage location</button></div>
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

    vi.mock("../../src/dialogs", () => {
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
      "<div class=\\"card mb-3\\">
        <div class=\\"card-body\\">
          <!--v-if-->
          <!--v-if-->
          <h4 class=\\"card-title\\">Location</h4>
          <p class=\\"card-text\\"><span>Not available yet.</span></p>
          <div class=\\"actions\\"><button type=\\"button\\" class=\\"btn btn-primary btn-sm\\" disabled=\\"\\">Choose storage location</button></div>
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
      "<div class=\\"card mb-3\\">
        <div class=\\"card-body\\">
          <!--v-if-->
          <!--v-if-->
          <h4 class=\\"card-title\\">Location</h4>
          <p class=\\"card-text\\"><span>Package rejected.</span></p>
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
      "<div class=\\"card mb-3\\">
        <div class=\\"card-body\\">
          <!--v-if-->
          <!--v-if-->
          <h4 class=\\"card-title\\">Location</h4>
          <p class=\\"card-text\\"><span><div class=\\"d-flex align-items-start gap-2\\"><span class=\\"font-monospace\\">f8635e46-a320-4152-9a2c-98a28eeb50d1</span>
            <!--v-if-->
        </div></span></p>
        <div class=\\"actions\\"><button type=\\"button\\" class=\\"btn btn-primary btn-sm\\" disabled=\\"\\">Choose storage location</button></div>
      </div>
      </div>"
    `);
  });
});
