import "./styles/main.scss";
import { PiniaDebounce } from "@pinia/plugin-debounce";
import { createPinia } from "pinia";
import { debounce } from "ts-debounce";
import { createApp } from "vue";
import { PromiseDialog } from "vue3-promise-dialog";

import App from "./App.vue";
import { api } from "./client";
import {
  FormatDateTime,
  FormatDateTimeString,
  FormatDuration,
} from "./composables/dateFormat";
import router from "./router";

const pinia = createPinia();
pinia.use(PiniaDebounce(debounce));

const app = createApp(App);
app.use(router);
app.use(pinia);
app.use(PromiseDialog);
app.mount("#app");

interface Filters {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: (...value: any[]) => string;
}

declare module "@vue/runtime-core" {
  interface ComponentCustomProperties {
    $filters: Filters;
  }
}

app.config.globalProperties.$filters = {
  formatDateTimeString(value: string) {
    return FormatDateTimeString(value);
  },
  formatDateTime(value: Date | undefined) {
    return FormatDateTime(value);
  },
  formatDuration(from: Date, to: Date) {
    return FormatDuration(from, to);
  },
  getWorkflowLabel(value: api.EnduroIngestSipWorkflowTypeEnum) {
    switch (value) {
      case api.EnduroIngestSipWorkflowTypeEnum.CreateAip:
        return "Create AIP";
      case api.EnduroIngestSipWorkflowTypeEnum.CreateAndReviewAip:
        return "Create and Review AIP";
      case api.EnduroIngestSipWorkflowTypeEnum.MovePackage:
        return "Move package";
      default:
        return value;
    }
  },
  getLocationSourceLabel(value: api.LocationSourceEnum) {
    switch (value) {
      case api.LocationSourceEnum.Minio:
        return "MinIO";
      case api.LocationSourceEnum.Sftp:
        return "SFTP";
      case api.LocationSourceEnum.Amss:
        return "AMSS";
      case api.LocationSourceEnum.Unspecified:
        return "Unspecified";
      default:
        return value;
    }
  },
  getLocationPurposeLabel(value: api.LocationPurposeEnum) {
    switch (value) {
      case api.LocationPurposeEnum.AipStore:
        return "AIP Store";
      case api.LocationPurposeEnum.Unspecified:
        return "Unspecified";
      default:
        return value;
    }
  },
};
