import Tabs from "@/components/Tabs.vue";
import { cleanup, render } from "@testing-library/vue";
import { afterEach, describe, it } from "vitest";
import { createRouter, createMemoryHistory } from "vue-router";

describe("Tabs.vue", () => {
  afterEach(() => cleanup());

  it("renders", async () => {
    const { getByRole } = render(Tabs, {
      props: {
        tabs: [
          {
            icon: "",
            text: "Route1",
            route: { name: "route1" },
            show: true,
          },
          {
            icon: "",
            text: "Route2",
            route: { name: "route2" },
            show: true,
          },
        ],
      },
      global: {
        plugins: [
          createRouter({
            history: createMemoryHistory(),
            routes: [
              { name: "index", path: "", component: {} },
              { name: "route1", path: "/route1", component: {} },
              { name: "route2", path: "/route2", component: {} },
            ],
          }),
        ],
      },
    });

    getByRole("navigation", { name: "Tabs" });
    getByRole("list");
    getByRole("link", { name: "Route1" });
    getByRole("link", { name: "Route2" });
  });
});
