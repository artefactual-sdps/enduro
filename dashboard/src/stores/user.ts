import { acceptHMRUpdate, defineStore } from "pinia";

import { api, client } from "@/client";
import { logError } from "@/helpers/logs";

// TODO: Reduce the default page size once we have an "ingested by" filter that
// allows us to filter the user list with an autocomplete or search box.
const defaultPageSize = 100;

export const useUserStore = defineStore("user", {
  state: () => ({
    // A list of Users shown during searches.
    users: [] as Array<api.EnduroIngestUser>,

    // Page is a subset of the total User list.
    page: { limit: defaultPageSize } as api.EnduroPage,

    // Search filters for the users list.
    filters: {
      email: undefined as string | undefined,
      name: undefined as string | undefined,
    },
  }),
  getters: {
    hasUsers(): boolean {
      return this.users.length > 0;
    },
  },
  actions: {
    async fetchUsers() {
      try {
        const response = await client.ingest.ingestListUsers({
          email: this.filters.email,
          name: this.filters.name,
          limit: this.page.limit,
          offset: this.page.offset,
        });
        this.users = response.items;
        this.page = response.page;
      } catch (e: unknown) {
        logError(e, "Failed to fetch users");
      }
    },
    getHandle(user: api.EnduroIngestUser): string {
      return user.name || user.email || user.uuid;
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useUserStore, import.meta.hot));
}
