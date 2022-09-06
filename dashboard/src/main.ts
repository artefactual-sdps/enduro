import App from "./App.vue";
import { client, api } from "./client";
import "./styles/main.scss";
import { PiniaDebounce } from "@pinia/plugin-debounce";
import humanizeDuration from "humanize-duration";
import moment from "moment";
import { createPinia, PiniaVuePlugin } from "pinia";
import { debounce } from "ts-debounce";
import { createApp } from "vue";
import { PromiseDialog } from "vue3-promise-dialog";
import { createRouter, createWebHistory } from "vue-router";
import routes from "~pages";

const router = createRouter({
  history: createWebHistory("/"),
  routes,
  strict: false,
});

const pinia = createPinia();
pinia.use(PiniaDebounce(debounce));

const app = createApp(App);
app.use(router);
app.use(pinia);
app.use(PromiseDialog);
app.mount("#app");

client.connectPackageMonitor();

interface Filters {
  [key: string]: (...value: any[]) => string;
}

declare module "@vue/runtime-core" {
  interface ComponentCustomProperties {
    $filters: Filters;
  }
}

app.config.globalProperties.$filters = {
  formatDateTimeString(value: string) {
    const date = new Date(value);
    return date.toLocaleString();
  },
  formatDateTime(value: Date | undefined) {
    if (!value) {
      return "";
    }
    return value.toLocaleString();
  },
  formatDuration(from: Date, to: Date) {
    const diff = moment(to).diff(from);
    return humanizeDuration(moment.duration(diff).asMilliseconds());
  },
  getPreservationActionLabel(
    value: api.EnduroPackagePreservationActionResponseBodyTypeEnum
  ) {
    switch (value) {
      case api.EnduroPackagePreservationActionResponseBodyTypeEnum.CreateAip:
        return "Create AIP";
      case api.EnduroPackagePreservationActionResponseBodyTypeEnum
        .CreateAndReviewAip:
        return "Create and Review AIP";
      case api.EnduroPackagePreservationActionResponseBodyTypeEnum.MovePackage:
        return "Move package";
      default:
        return value;
    }
  },
  getLocationSourceLabel(value: api.StorageAddLocationRequestBodySourceEnum) {
    switch (value) {
      case api.StorageAddLocationRequestBodySourceEnum.Minio:
        return "MinIO";
      case api.StorageAddLocationRequestBodySourceEnum.Unspecified:
        return "Unspecified";
      default:
        return value;
    }
  },
  getLocationPurposeLabel(value: api.StorageAddLocationRequestBodyPurposeEnum) {
    switch (value) {
      case api.StorageAddLocationRequestBodyPurposeEnum.AipStore:
        return "AIP Store";
      case api.StorageAddLocationRequestBodyPurposeEnum.Unspecified:
        return "Unspecified";
      default:
        return value;
    }
  },
};
