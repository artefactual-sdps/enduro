import { useLayoutStore } from "../../src/stores/layout";
import { setActivePinia, createPinia } from "pinia";
import { expect, describe, it, beforeEach } from "vitest";

describe("useLayoutStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("toggles the sidebarCollapsed property", () => {
    const layoutStore = useLayoutStore();
    layoutStore.sidebarCollapsed = false;

    layoutStore.toggleSidebar();
    expect(layoutStore.sidebarCollapsed).toEqual(true);
  });

  it("updates the breadcrumb property", () => {
    const layoutStore = useLayoutStore();
    const breadcrumb = [{ text: "Packages" }];

    layoutStore.updateBreadcrumb(breadcrumb);
    expect(layoutStore.breadcrumb).toEqual(breadcrumb);
  });
});
