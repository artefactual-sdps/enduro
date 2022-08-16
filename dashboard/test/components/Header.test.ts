import Header from "@/components/Header.vue";
import { useStateStore } from "@/stores/state";
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
              state: {
                sidebarCollapsed: false,
              },
            },
          }),
        ],
      },
    });

    const stateStore = useStateStore();

    const expandButton = getByRole("button", {
      name: "Collapse navigation",
    });

    await fireEvent.click(expandButton);
    expect(stateStore.sidebarCollapsed).toEqual(true);

    const collapseButton = getByRole("button", {
      name: "Expand navigation",
    });

    await fireEvent.click(collapseButton);
    expect(stateStore.sidebarCollapsed).toEqual(false);
  });

  it("displays the breadcrumb navigation", async () => {
    const { getByRole } = render(Header, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              state: { breadcrumb: [{ text: "Packages" }] },
            },
          }),
        ],
      },
    });

    getByRole("navigation", { name: "Breadcrumb" });
  });
});
