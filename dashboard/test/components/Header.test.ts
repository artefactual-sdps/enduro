import Header from "@/components/Header.vue";
import { useLayoutStore } from "@/stores/layout";
import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";

describe("Header.vue", () => {
  afterEach(() => cleanup());

  it("collapses and expands the sidebar", async () => {
    const { getByRole } = render(Header, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              layout: {
                sidebarCollapsed: false,
              },
            },
            stubActions: false,
          }),
        ],
      },
    });

    const layoutStore = useLayoutStore();

    const expandButton = getByRole("button", {
      name: "Collapse navigation",
    });

    await fireEvent.click(expandButton);
    expect(layoutStore.sidebarCollapsed).toEqual(true);

    const collapseButton = getByRole("button", {
      name: "Expand navigation",
    });

    await fireEvent.click(collapseButton);
    expect(layoutStore.sidebarCollapsed).toEqual(false);
  });

  it("displays the breadcrumb navigation", async () => {
    const { getByRole } = render(Header, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              layout: { breadcrumb: [{ text: "Packages" }] },
            },
          }),
        ],
      },
    });

    getByRole("navigation", { name: "Breadcrumb" });
  });
});
