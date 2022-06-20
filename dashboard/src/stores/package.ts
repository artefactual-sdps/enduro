import { clientProviderKey, Client, api } from "../client";
import { defineStore, acceptHMRUpdate } from "pinia";
import { inject } from "vue";

export const usePackageStore = defineStore("package", {
  state: () => ({
    current: null as api.PackageShowResponseBody | null,
    current_preservation_actions:
      null as api.PackagePreservationActionsResponseBody | null,
  }),
  actions: {
    fetchCurrent(id: string) {
      this.reset();
      const packageId = +id;
      if (Number.isNaN(packageId)) {
        return;
      }
      const client = inject(clientProviderKey) as Client;
      client.package.packageShow({ id: packageId }).then((payload) => {
        this.current = payload;
      });
      client.package
        .packagePreservationActions({ id: packageId })
        .then((payload) => {
          this.current_preservation_actions = payload;
        });
    },
    confirm() {},
    reject() {},
    reset() {
      this.current = null;
      this.current_preservation_actions = null;
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(usePackageStore, import.meta.hot));
}
