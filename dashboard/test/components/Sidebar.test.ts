import Sidebar from "@/components/Sidebar.vue";
import { useStateStore } from "@/stores/state";
import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { flushPromises } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";
import { createRouter, createWebHistory } from "vue-router";

describe("Sidebar.vue", () => {
  afterEach(() => cleanup());

  it("renders the navigation links", async () => {
    const { getByRole } = render(Sidebar, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              state: {
                sidebarCollapsed: false,
              },
            },
          }),
          createRouter({
            history: createWebHistory(),
            routes: [
              { name: "packages", path: "/packages", component: {} },
              { name: "locations", path: "/locations", component: {} },
            ],
          }),
        ],
      },
    });

    getByRole("navigation", { name: "Navigation" });
    const packagesLink = getByRole("link", { name: "Packages" });
    const locationsLink = getByRole("link", { name: "Locations" });

    fireEvent.click(packagesLink);
    await flushPromises();
    expect(packagesLink.getAttribute("aria-current")).toEqual("page");

    fireEvent.click(locationsLink);
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
              state: {
                sidebarCollapsed: false,
              },
            },
          }),
        ],
      },
    });

    const stateStore = useStateStore();
    stateStore.sidebarCollapsed = true;
    await nextTick();

    expect(container.firstElementChild?.getAttribute("class")).toContain(
      "collapsed"
    );

    stateStore.sidebarCollapsed = false;
    await nextTick();

    expect(container.firstElementChild?.getAttribute("class")).not.toContain(
      "collapsed"
    );
  });
});
