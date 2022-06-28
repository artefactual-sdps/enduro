import { api, client } from "@/client";
import { defineStore, acceptHMRUpdate } from "pinia";

export const useStorageStore = defineStore("storage", {
  state: () => ({
    package: null as api.StorageShowResponseBody | null,
    locations: [] as Array<api.StoredLocationResponse>,
  }),
  getters: {},
  actions: {
    async fetchPackage(id: string) {
      try {
        this.package = await client.storage.storageShow({ aipId: id });
      } catch (error) {
        return error;
      }
    },
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
