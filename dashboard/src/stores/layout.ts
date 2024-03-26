import auth from "@/auth";
import router from "@/router";
import type { User } from "oidc-client-ts";
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
    user: null as User | null,
  }),
  getters: {
    isUserValid(): boolean {
      return this.user != null && !this.user.expired;
    },
    getUserDisplayName(): string | undefined {
      return (
        this.user?.profile.preferred_username ||
        this.user?.profile.name ||
        this.user?.profile.email
      );
    },
  },
  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed;
    },
    updateBreadcrumb(breadcrumb: Array<BreadcrumbItem>) {
      this.breadcrumb = breadcrumb;
    },
    // Load the currently authenticated user.
    async loadUser() {
      if (this.user === null) {
        const user = await auth.getUser();
        this.setUser(user);
      }
    },
    setUser(user: User | null) {
      this.user = user;
    },
    removeUser() {
      // Dex doesn't allow to end sessions upstream:
      // https://github.com/dexidp/dex/issues/1697.
      auth.removeUser().then(() => {
        this.user = null;
        router.push({ name: "/" });
      });
    },
  },
});
