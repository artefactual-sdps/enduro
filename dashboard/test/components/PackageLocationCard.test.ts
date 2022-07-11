import { api } from "../../src/client";
import PackageLocationCard from "../../src/components/PackageLocationCard.vue";
import { usePackageStore } from "../../src/stores/package";
import { createTestingPinia } from "@pinia/testing";
import { render, fireEvent } from "@testing-library/vue";
import { describe, it, vi, expect } from "vitest";

describe("PackageLocationCard.vue", () => {
  it("renders when the package is stored", async () => {
    const { getByText, html } = render(PackageLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            stubActions: false,
            initialState: {
              package: {
                current: {
                  status: api.PackageShowResponseBodyStatusEnum.Done,
                  location: "perma-aips-1",
                },
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
          <h5 class=\\"card-title\\">Location</h5>
          <p class=\\"card-text\\"><span>perma-aips-1</span></p>
          <div class=\\"actions\\"><button type=\\"button\\" class=\\"btn btn-primary btn-sm\\" disabled=\\"false\\">Choose storage location</button></div>
        </div>
      </div>"
    `);

    vi.mock("../../src/dialogs", () => {
      return {
        openPackageLocationDialog: vi.fn().mockResolvedValue("perma-aips-2"),
      };
    });

    const packageStore = usePackageStore();
    packageStore.move.mockImplementation(() => {
      packageStore.current.status =
        api.EnduroStoredPackageResponseBodyStatusEnum.InProgress;
      packageStore.locationChanging = true;
    });

    const button = getByText("Choose storage location");
    await fireEvent.click(button);

    expect(html()).toMatchInlineSnapshot(`
      "<div class=\\"card mb-3\\">
        <div class=\\"card-body\\">
          <!--v-if-->
          <div class=\\"alert alert-info\\" role=\\"alert\\"> The package is being moved into a new location. </div>
          <h5 class=\\"card-title\\">Location</h5>
          <p class=\\"card-text\\"><span>perma-aips-1</span></p>
          <div class=\\"actions\\"><button type=\\"button\\" class=\\"btn btn-primary btn-sm\\" disabled=\\"true\\"><span class=\\"spinner-grow spinner-grow-sm me-2\\" role=\\"status\\" aria-hidden=\\"true\\"></span> Moving... </button></div>
        </div>
      </div>"
    `);
  });

  it("renders when the package location is not available", async () => {
    const { html } = render(PackageLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div class=\\"card mb-3\\">
        <div class=\\"card-body\\">
          <!--v-if-->
          <!--v-if-->
          <h5 class=\\"card-title\\">Location</h5>
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
                  location: undefined,
                },
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
          <h5 class=\\"card-title\\">Location</h5>
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
                  location: "perma-aips-1",
                },
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
          <h5 class=\\"card-title\\">Location</h5>
          <p class=\\"card-text\\"><span>perma-aips-1</span></p>
          <div class=\\"actions\\"><button type=\\"button\\" class=\\"btn btn-primary btn-sm\\" disabled=\\"true\\">Choose storage location</button></div>
        </div>
      </div>"
    `);
  });
});
