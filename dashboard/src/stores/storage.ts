import { api, client } from "@/client";
import { defineStore, acceptHMRUpdate } from "pinia";

export const useStorageStore = defineStore("storage", {
  state: () => ({
    locations: [] as Array<api.StoredLocationResponse>,
  }),
  getters: {},
  actions: {
    async fetchLocations() {
      try {
        this.locations = await client.storage.storageList();
      } catch (error) {
        return error;
      }
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useStorageStore, import.meta.hot));
}
