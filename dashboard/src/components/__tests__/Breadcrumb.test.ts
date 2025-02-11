import { createTestingPinia } from "@pinia/testing";
import { cleanup, render } from "@testing-library/vue";
import { afterEach, describe, it, vi } from "vitest";
import { createRouter, createWebHistory } from "vue-router";

import Breadcrumb from "@/components/Breadcrumb.vue";

describe("Breadcrumb.vue", () => {
  afterEach(() => cleanup());

  it("renders", async () => {
    const { getByRole, getByText } = render(Breadcrumb, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              layout: {
                breadcrumb: [
                  { route: { name: "packages" }, text: "Packages" },
                  { text: "Package.zip" },
                ],
              },
            },
          }),
          createRouter({
            history: createWebHistory(),
            routes: [
              { name: "index", path: "", component: {} },
              { name: "packages", path: "/packages", component: {} },
            ],
          }),
        ],
      },
    });

    getByRole("navigation", { name: "Breadcrumb" });
    getByRole("list");
    getByRole("listitem", { current: "page" });
    getByRole("link", { name: "Packages" });
    getByText("Package.zip");
  });
});
