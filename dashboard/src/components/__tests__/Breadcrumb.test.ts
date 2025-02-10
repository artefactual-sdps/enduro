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
                  { text: "Ingest" },
                  { route: { name: "sips" }, text: "SIPs" },
                  { text: "SIP.zip" },
                ],
              },
            },
          }),
          createRouter({
            history: createWebHistory(),
            routes: [
              { name: "index", path: "", component: {} },
              { name: "sips", path: "/ingest/sips", component: {} },
            ],
          }),
        ],
      },
    });

    getByRole("navigation", { name: "Breadcrumb" });
    getByRole("list");
    getByText("Ingest");
    getByRole("link", { name: "SIPs" });
    getByRole("listitem", { current: "page" });
    getByText("SIP.zip");
  });
});
