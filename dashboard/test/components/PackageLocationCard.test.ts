import { api } from "../../src/client";
import PackageLocationCard from "../../src/components/PackageLocationCard.vue";
import { createTestingPinia } from "@pinia/testing";
import { render } from "@testing-library/vue";
import { describe, it, vi, expect } from "vitest";

describe("PackageLocationCard.vue", () => {
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
                  location: "perma-aips-1",
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
          <h5 class=\\"card-title\\">Location</h5>
          <p class=\\"card-text\\"><span>perma-aips-1</span></p>
          <div class=\\"actions\\"><button type=\\"button\\" class=\\"btn btn-primary btn-sm\\" disabled=\\"false\\">Choose storage location</button></div>
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
          <h5 class=\\"card-title\\">Location</h5>
          <p class=\\"card-text\\"><span>perma-aips-1</span></p>
          <div class=\\"actions\\"><button type=\\"button\\" class=\\"btn btn-primary btn-sm\\" disabled=\\"true\\">Choose storage location</button></div>
        </div>
      </div>"
    `);
  });
});
