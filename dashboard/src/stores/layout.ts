import { defineStore } from "pinia";

type BreadcrumbItem = {
  routeName?: string;
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
