import { api, client } from "@/client";
import { defineStore, acceptHMRUpdate } from "pinia";

export const usePackageStore = defineStore("package", {
  state: () => ({
    current: null as api.PackageShowResponseBody | null,
    current_preservation_actions:
      null as api.PackagePreservationActionsResponseBody | null,
    locationChanging: false,
  }),
  getters: {
    isPending(): boolean {
      return (
        this.current?.status ==
        api.EnduroStoredPackageResponseBodyStatusEnum.Pending
      );
    },
    isMovable(): boolean {
      return !this.locationChanging && Boolean(this.current?.location);
    },
  },
  actions: {
    async fetchCurrent(id: string) {
      this.reset();
      const packageId = +id;
      if (Number.isNaN(packageId)) {
        return;
      }

      try {
        this.current = await client.package.packageShow({ id: packageId });
        this.current_preservation_actions =
          await client.package.packagePreservationActions({ id: packageId });
      } catch (error) {
        return error;
      }

      const error = await this.moveStatus();
      if (error) {
        return error;
      }
    },
    confirm(locationName: string) {
      if (!this.current) return;
      client.package
        .packageConfirm({
          id: this.current.id,
          confirmRequestBody: { location: locationName },
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
    async move(locationName: string) {
      if (!this.current) return;
      try {
        await client.package
          .packageMove({
            id: this.current.id,
            moveRequestBody: { location: locationName },
          })
          .then((payload) => {
            if (!this.current) return;
            this.$patch((state) => {
              state.locationChanging = true;
              if (!state.current) return;
              state.current.status =
                api.EnduroStoredPackageResponseBodyStatusEnum.InProgress;
            });
          });
      } catch (error) {
        return error;
      }
    },
    async moveStatus() {
      if (!this.current) return;
      let resp;
      try {
        resp = await client.package.packageMoveStatus({
          id: this.current?.id,
        });
      } catch (error) {
        return error;
      }
      this.locationChanging = !resp.done;
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
