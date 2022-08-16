import Breadcrumb from "@/components/Breadcrumb.vue";
import { createTestingPinia } from "@pinia/testing";
import { cleanup, render } from "@testing-library/vue";
import { afterEach, describe, it, vi } from "vitest";
import { createRouter, createWebHistory } from "vue-router";

describe("Breadcrumb.vue", () => {
  afterEach(() => cleanup());

  it("renders", async () => {
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ name: "packages", path: "/packages", component: {} }],
    });
    const { getByRole, getByText } = render(Breadcrumb, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              state: {
                breadcrumb: [
                  { routeName: "packages", text: "Packages" },
                  { text: "Package.zip" },
                ],
              },
            },
          }),
          router,
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
