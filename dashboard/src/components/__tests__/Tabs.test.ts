import { cleanup, render } from "@testing-library/vue";
import { afterEach, describe, expect, it } from "vitest";
import { createMemoryHistory, createRouter } from "vue-router";

import Tabs from "@/components/Tabs.vue";
import RawIconHomeLine from "~icons/clarity/home-line?raw&width=2em&height=2em";

describe("Tabs.vue", () => {
  afterEach(() => cleanup());

  it("renders", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { name: "index", path: "", component: {} },
        { name: "route1", path: "/route1", component: {} },
        { name: "route2", path: "/route2", component: {} },
      ],
    });

    const { getByRole, queryByRole } = render(Tabs, {
      props: {
        tabs: [
          {
            icon: RawIconHomeLine,
            text: "Route1",
            route: router.resolve("/route1"),
            show: true,
          },
          {
            icon: RawIconHomeLine,
            text: "Route2",
            route: router.resolve("/route2"),
            show: false,
          },
        ],
        param: "",
      },
      global: {
        plugins: [router],
      },
    });

    getByRole("navigation", { name: "Tabs" });
    getByRole("list");
    getByRole("link", { name: "Route1" });
    expect(queryByRole("link", { name: "Route2" })).toBeNull();
  });
});
