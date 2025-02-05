import { api, client } from "@/client";
import { useLayoutStore } from "@/stores/layout";
import { defineStore, acceptHMRUpdate } from "pinia";
import router from "@/router";

export const useStorageStore = defineStore("storage", {
  state: () => ({
    locations: [] as Array<api.LocationResponse>,
    current: null as api.LocationResponse | null,
    current_packages: [] as Array<api.PackageResponse>,
  }),
  getters: {},
  actions: {
    async fetchLocations() {
      const resp = await client.storage.storageLocations();
      this.locations = resp;
    },
    async fetchCurrent(uuid: string) {
      this.$reset();

      this.current = await client.storage.storageShowLocation({ uuid: uuid });

      // Update breadcrumb. TODO: should this be done in the component?
      const layoutStore = useLayoutStore();
      layoutStore.updateBreadcrumb([
        { text: "Storage" },
        { route: router.resolve({ name: "/locations/" }), text: "Locations" },
        { text: this.current.name },
      ]);

      await Promise.all([
        client.storage.storageLocationPackages({ uuid: uuid }).then((resp) => {
          this.current_packages = resp;
        }),
      ]);
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useStorageStore, import.meta.hot));
}
