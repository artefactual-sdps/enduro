import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { flushPromises } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";

import Sidebar from "@/components/Sidebar.vue";
import { useLayoutStore } from "@/stores/layout";

const router = createRouter({
  // Vue Router throws history.state warning when the second click is fired,
  // createMemoryHistory does not have this problem.
  history: createMemoryHistory(),
  routes: [
    { name: "index", path: "", component: {} },
    { name: "sips", path: "/ingest/sips", component: {} },
    { name: "upload", path: "/ingest/upload", component: {} },
    { name: "locations", path: "/storage/locations", component: {} },
    { name: "aips", path: "/storage/aips", component: {} },
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
    const homeLink = getByRole("link", { name: "Home" });
    const sipsLink = getByRole("link", { name: "SIPs" });
    const uploadLink = getByRole("link", { name: "Upload SIPs" });
    const locationsLink = getByRole("link", { name: "Locations" });
    const aipsLink = getByRole("link", { name: "AIPs" });

    fireEvent.click(homeLink);
    await flushPromises();
    expect(homeLink.getAttribute("aria-current")).toEqual("page");

    fireEvent.click(sipsLink);
    await flushPromises();
    expect(sipsLink.getAttribute("aria-current")).toEqual("page");

    fireEvent.click(uploadLink);
    await flushPromises();
    expect(uploadLink.getAttribute("aria-current")).toEqual("page");

    fireEvent.click(locationsLink);
    await flushPromises();
    expect(locationsLink.getAttribute("aria-current")).toEqual("page");

    fireEvent.click(aipsLink);
    await flushPromises();
    expect(aipsLink.getAttribute("aria-current")).toEqual("page");
  });

  it("hides the navigation links based on auth. attributes", async () => {
    const { getByRole, queryByRole } = render(Sidebar, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              layout: {
                sidebarCollapsed: false,
              },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: [],
              },
            },
          }),
          router,
        ],
      },
    });

    getByRole("navigation", { name: "Navigation" });
    getByRole("link", { name: "Home" });
    expect(queryByRole("link", { name: "SIPs" })).toBeNull();
    expect(queryByRole("link", { name: "Upload SIPs" })).toBeNull();
    expect(queryByRole("link", { name: "Locations" })).toBeNull();
    expect(queryByRole("link", { name: "AIPs" })).toBeNull();
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
      "collapsed",
    );

    layoutStore.sidebarCollapsed = false;
    await nextTick();

    expect(container.firstElementChild?.getAttribute("class")).not.toContain(
      "collapsed",
    );
  });
});
