import Sidebar from "@/components/Sidebar.vue";
import { useLayoutStore } from "@/stores/layout";
import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { flushPromises } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";
import { createRouter, createMemoryHistory } from "vue-router";

const router = createRouter({
  // Vue Router throws history.state warning when the second click is fired,
  // createMemoryHistory does not have this problem.
  history: createMemoryHistory(),
  routes: [
    { name: "index", path: "", component: {} },
    { name: "packages", path: "/packages", component: {} },
    { name: "locations", path: "/locations", component: {} },
  ],
});

describe("Sidebar.vue", () => {
  afterEach(() => cleanup());

  it("renders the navigation links", async () => {
    const { getByRole } = render(Sidebar, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              layout: {
                sidebarCollapsed: false,
              },
            },
          }),
          router,
        ],
      },
    });

    getByRole("navigation", { name: "Navigation" });
    const packagesLink = getByRole("link", { name: "Packages" });
    const locationsLink = getByRole("link", { name: "Locations" });

    fireEvent.click(packagesLink);
    await flushPromises();
    expect(packagesLink.getAttribute("aria-current")).toEqual("page");

    await fireEvent.click(locationsLink);
    await flushPromises();
    expect(locationsLink.getAttribute("aria-current")).toEqual("page");
  });

  it("collapses and expands", async () => {
    const { container } = render(Sidebar, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              layout: {
                sidebarCollapsed: false,
              },
            },
          }),
          router,
        ],
      },
    });

    const layoutStore = useLayoutStore();
    layoutStore.sidebarCollapsed = true;
    await nextTick();

    expect(container.firstElementChild?.getAttribute("class")).toContain(
      "collapsed"
    );

    layoutStore.sidebarCollapsed = false;
    await nextTick();

    expect(container.firstElementChild?.getAttribute("class")).not.toContain(
      "collapsed"
    );
  });
});
