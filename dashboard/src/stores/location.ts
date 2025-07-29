import { acceptHMRUpdate, defineStore } from "pinia";

import { api, client } from "@/client";
import router from "@/router";
import { useLayoutStore } from "@/stores/layout";

export const useLocationStore = defineStore("location", {
  state: () => ({
    // Location currently displayed.
    current: null as api.LocationResponse | null,

    // AIPs of the current Location.
    currentAips: [] as Array<api.AIPResponse>,

    // A list of Locations shown during searches.
    locations: [] as Array<api.LocationResponse>,
  }),
  getters: {},
  actions: {
    async fetchCurrent(uuid: string) {
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
    async fetchLocations() {
      const resp = await client.storage.storageListLocations();
      this.locations = resp;
    },
    async fetchLocationsDebounced() {
      return this.fetchLocations();
    },
  },
  debounce: {
    fetchLocationsDebounced: [500, { isImmediate: false }],
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useLocationStore, import.meta.hot));
}
