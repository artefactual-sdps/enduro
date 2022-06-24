import { api, client } from "../client";
import { defineStore, acceptHMRUpdate } from "pinia";

export const usePackageStore = defineStore("package", {
  state: () => ({
    current: null as api.PackageShowResponseBody | null,
    current_preservation_actions:
      null as api.PackagePreservationActionsResponseBody | null,
  }),
  getters: {
    isPending(): boolean {
      return (
        this.current?.status ==
        api.EnduroStoredPackageResponseBodyStatusEnum.Pending
      );
    },
  },
  actions: {
    fetchCurrent(id: string) {
      this.reset();
      const packageId = +id;
      if (Number.isNaN(packageId)) {
        return;
      }
      client.package.packageShow({ id: packageId }).then((payload) => {
        this.current = payload;
      });
      client.package
        .packagePreservationActions({ id: packageId })
        .then((payload) => {
          this.current_preservation_actions = payload;
        });
    },
    confirm() {
      if (!this.current) return;
      client.package
        .packageConfirm({
          id: this.current.id,
          confirmRequestBody: { location: "perma-aips-2" },
        })
        .then((payload) => {
          if (!this.current) return;
          this.current.status =
            api.EnduroStoredPackageResponseBodyStatusEnum.InProgress;
        });
    },
    reject() {
      if (!this.current) return;
      client.package.packageReject({ id: this.current.id }).then((payload) => {
        if (!this.current) return;
        this.current.status =
          api.EnduroStoredPackageResponseBodyStatusEnum.InProgress;
      });
    },
    reset() {
      this.current = null;
      this.current_preservation_actions = null;
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(usePackageStore, import.meta.hot));
}
