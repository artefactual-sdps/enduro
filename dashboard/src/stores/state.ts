import { defineStore } from "pinia";

type BreadcrumbItem = {
  routeName?: string;
  text?: string;
};

export const useStateStore = defineStore("state", {
  state: () => ({
    sidebarCollapsed: false as boolean,
    breadcrumb: [] as Array<BreadcrumbItem>,
  }),
  actions: {
    expandSidebar() {
      this.sidebarCollapsed = false;
    },
    collapseSidebar() {
      this.sidebarCollapsed = true;
    },
    updateBreadcrumb(breadcrumb: Array<BreadcrumbItem>) {
      this.breadcrumb = breadcrumb;
    },
  },
});
