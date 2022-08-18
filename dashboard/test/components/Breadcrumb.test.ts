import Breadcrumb from "@/components/Breadcrumb.vue";
import { createTestingPinia } from "@pinia/testing";
import { cleanup, render } from "@testing-library/vue";
import { afterEach, describe, it, vi } from "vitest";
import { createRouter, createWebHistory } from "vue-router";

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
                  { routeName: "packages", text: "Packages" },
                  { text: "Package.zip" },
                ],
              },
            },
          }),
          createRouter({
            history: createWebHistory(),
            routes: [{ name: "packages", path: "/packages", component: {} }],
          }),
        ],
      },
      routes: [],
    });

    getByRole("navigation", { name: "Breadcrumb" });
    getByRole("list");
    getByRole("link", { name: "Packages" });
    getByRole("listitem", { current: "page" });
    getByText("Package.zip");
  });
});
