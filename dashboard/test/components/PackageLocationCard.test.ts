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
                  status: api.PackageShowResponseBodyStatusEnum.Done,
                  locationId: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                  locationName: "perma-aips-1",
                } as api.PackageShowResponseBody,
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
          <p class=\\"card-text\\"><span>perma-aips-1</span></p>
          <div class=\\"actions\\"><button type=\\"button\\" class=\\"btn btn-primary btn-sm\\" disabled=\\"false\\">Choose storage location</button></div>
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
                  status: api.PackageShowResponseBodyStatusEnum.Done,
                  locationId: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                  locationName: "perma-aips-1",
                } as api.PackageShowResponseBody,
              },
            },
          }),
        ],
      },
    });

    getByText("perma-aips-1");

    const packageStore = usePackageStore();

    const moveMock = vi.fn().mockImplementation(packageStore.move);
    moveMock.mockImplementation(async () => {
      packageStore.$patch((state) => {
        if (!state.current) return;
        state.current.status =
          api.EnduroStoredPackageResponseBodyStatusEnum.InProgress;
        state.locationChanging = true;
      });
    });
    packageStore.move = moveMock;

    vi.mock("../../src/dialogs", () => {
      return {
        openPackageLocationDialog: () => {
          return {
            locationId: "fe675e52-c761-46d0-8605-fae4bd10303e",
            locationName: "perma-aips-2",
          };
        },
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
                  status: api.PackageShowResponseBodyStatusEnum.InProgress,
                } as api.PackageShowResponseBody,
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
          <div class=\\"actions\\"><button type=\\"button\\" class=\\"btn btn-primary btn-sm\\" disabled=\\"true\\">Choose storage location</button></div>
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
                  status: api.PackageShowResponseBodyStatusEnum.Done,
                  locationId: undefined,
                } as api.PackageShowResponseBody,
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
                  status: api.PackageShowResponseBodyStatusEnum.InProgress,
                  locationId: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                  locationName: "perma-aips-1",
                } as api.PackageShowResponseBody,
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
          <p class=\\"card-text\\"><span>perma-aips-1</span></p>
          <div class=\\"actions\\"><button type=\\"button\\" class=\\"btn btn-primary btn-sm\\" disabled=\\"true\\">Choose storage location</button></div>
        </div>
      </div>"
    `);
  });
});
