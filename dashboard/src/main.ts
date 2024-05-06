import App from "./App.vue";
import { api } from "./client";
import router from "./router";
import "./styles/main.scss";
import { PiniaDebounce } from "@pinia/plugin-debounce";
import humanizeDuration from "humanize-duration";
import moment from "moment";
import { createPinia } from "pinia";
import { debounce } from "ts-debounce";
import { createApp } from "vue";
import { PromiseDialog } from "vue3-promise-dialog";

const pinia = createPinia();
pinia.use(PiniaDebounce(debounce));

const app = createApp(App);
app.use(router);
app.use(pinia);
app.use({
  install: (app: any): any => {
    PromiseDialog.install(app, {});
  },
});
app.mount("#app");

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
    return moment(String(date)).format("YYYY-MM-DD HH:mm:ss");
  },
  formatDateTime(value: Date | undefined) {
    if (!value) {
      return "";
    }
    return moment(String(value)).format("YYYY-MM-DD HH:mm:ss");
  },
  formatDuration(from: Date, to: Date) {
    const diff = moment(to).diff(from);
    return humanizeDuration(moment.duration(diff).asMilliseconds());
  },
  getPreservationActionLabel(
    value: api.EnduroPackagePreservationActionTypeEnum,
  ) {
    switch (value) {
      case api.EnduroPackagePreservationActionTypeEnum.CreateAip:
        return "Create AIP";
      case api.EnduroPackagePreservationActionTypeEnum.CreateAndReviewAip:
        return "Create and Review AIP";
      case api.EnduroPackagePreservationActionTypeEnum.MovePackage:
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
