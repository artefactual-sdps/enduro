import { defineStore } from "pinia";
import type { RouteLocation } from "vue-router";

type BreadcrumbItem = {
  route?: RouteLocation;
  text?: string;
};

export const useLayoutStore = defineStore("layout", {
  state: () => ({
    sidebarCollapsed: false as boolean,
    breadcrumb: [] as Array<BreadcrumbItem>,
  }),
  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed;
    },
    updateBreadcrumb(breadcrumb: Array<BreadcrumbItem>) {
      this.breadcrumb = breadcrumb;
    },
  },
});
