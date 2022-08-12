import { defineStore } from "pinia";

export const useStateStore = defineStore("state", {
  state: () => ({ sidebarCollapsed: false as boolean }),
  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed;
    },
  },
});
