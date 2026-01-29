import { acceptHMRUpdate, defineStore } from "pinia";

import { api, client } from "@/client";
import { logError } from "@/helpers/logs";

export const useAboutStore = defineStore("about", {
  state: () => ({
    loaded: false,
    childWorkflows: [] as Array<api.EnduroChildworkflow>,
    preservationSystem: "a3m",
    uploadMaxSize: 0,
    version: "",
  }),
  getters: {
    // formattedUploadMaxSize returns the upload max size formatted as a string
    // with appropriate units (bytes, KiB, MiB, GiB).
    formattedUploadMaxSize: (state) => {
      const size = state.uploadMaxSize;
      if (size >= 1024 ** 4) {
        return `${(size / 1024 ** 4).toFixed(2)} TiB`;
      } else if (size >= 1024 ** 3) {
        return `${(size / 1024 ** 3).toFixed(2)} GiB`;
      } else if (size >= 1024 ** 2) {
        return `${(size / 1024 ** 2).toFixed(2)} MiB`;
      } else if (size >= 1024) {
        return `${(size / 1024).toFixed(2)} KiB`;
      } else {
        return `${size} bytes`;
      }
    },
    // formattedVersion returns the version prefixed with "v" or "unknown"
    // if not set.
    formattedVersion: (state) => {
      if (!state.version) {
        return "unknown";
      }
      return "v" + state.version;
    },
  },
  actions: {
    // fetch retrieves the about data from the API and updates the store state.
    async fetch() {
      return client.about
        .aboutAbout()
        .then((resp) => {
          this.loaded = true;
          this.preservationSystem = resp.preservationSystem;
          this.uploadMaxSize = resp.uploadMaxSize;
          this.version = resp.version;
          if (resp.childWorkflows) {
            this.childWorkflows = resp.childWorkflows;
          }
        })
        .catch((e) => {
          logError(e, "Error fetching about data");
        });
    },
    // Load fetches the about data from the API if it hasn't been loaded yet.
    async load() {
      if (!this.loaded) {
        return this.fetch();
      }
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useAboutStore, import.meta.hot));
}
