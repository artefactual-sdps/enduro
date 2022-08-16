import { useStateStore } from "../../src/stores/state";
import { setActivePinia, createPinia } from "pinia";
import { expect, describe, it, beforeEach } from "vitest";

describe("useStateStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("modifies the sidebarCollapsed property", () => {
    const stateStore = useStateStore();

    stateStore.collapseSidebar();
    expect(stateStore.sidebarCollapsed).toEqual(true);

    stateStore.expandSidebar();
    expect(stateStore.sidebarCollapsed).toEqual(false);
  });

  it("updates the breadcrumb property", () => {
    const stateStore = useStateStore();
    const breadcrumb = [{ text: "Packages" }];

    stateStore.updateBreadcrumb(breadcrumb);
    expect(stateStore.breadcrumb).toEqual(breadcrumb);
  });
});
