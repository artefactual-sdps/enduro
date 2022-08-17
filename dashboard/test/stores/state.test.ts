import { useStateStore } from "../../src/stores/state";
import { setActivePinia, createPinia } from "pinia";
import { expect, describe, it, beforeEach } from "vitest";

describe("useStateStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("toggles the sidebarCollapsed property", () => {
    const stateStore = useStateStore();
    stateStore.sidebarCollapsed = false;

    stateStore.toggleSidebar();
    expect(stateStore.sidebarCollapsed).toEqual(true);
  });

  it("updates the breadcrumb property", () => {
    const stateStore = useStateStore();
    const breadcrumb = [{ text: "Packages" }];

    stateStore.updateBreadcrumb(breadcrumb);
    expect(stateStore.breadcrumb).toEqual(breadcrumb);
  });
});
