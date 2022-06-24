import { api, client } from "@/client";
import { defineStore, acceptHMRUpdate } from "pinia";

export const useStorageStore = defineStore("storage", {
  state: () => ({
    locations: [] as Array<{ name: string }>,
  }),
  getters: {},
  actions: {
    fetchLocations() {
      this.reset();
      this.locations = [{ name: "perma-aips-1" }, { name: "perma-aips-2" }];
    },
    reset() {
      this.locations = [];
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useStorageStore, import.meta.hot));
}
