import { acceptHMRUpdate, defineStore } from "pinia";

import { api, client } from "@/client";
import router from "@/router";
import { useLayoutStore } from "@/stores/layout";

export const useStorageStore = defineStore("storage", {
  state: () => ({
    locations: [] as Array<api.LocationResponse>,
    current: null as api.LocationResponse | null,
    currentAips: [] as Array<api.AIPResponse>,
  }),
  getters: {},
  actions: {
    async fetchLocations() {
      const resp = await client.storage.storageListLocations();
      this.locations = resp;
    },
    async fetchCurrent(uuid: string) {
      this.$reset();

      this.current = await client.storage.storageShowLocation({ uuid: uuid });

      // Update breadcrumb. TODO: should this be done in the component?
      const layoutStore = useLayoutStore();
      layoutStore.updateBreadcrumb([
        { text: "Storage" },
        {
          route: router.resolve({ name: "/storage/locations/" }),
          text: "Locations",
        },
        { text: this.current.name },
      ]);

      await Promise.all([
        client.storage.storageListLocationAips({ uuid: uuid }).then((resp) => {
          this.currentAips = resp;
        }),
      ]);
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useStorageStore, import.meta.hot));
}
