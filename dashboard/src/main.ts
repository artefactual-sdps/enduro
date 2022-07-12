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
  formatPreservationActionStatus(
    value: api.EnduroPackagePreservationActionResponseBodyStatusEnum
  ) {
    switch (value) {
      case api.EnduroPackagePreservationActionResponseBodyStatusEnum.Complete:
        return "bg-success";
      case api.EnduroPackagePreservationActionResponseBodyStatusEnum.Failed:
        return "bg-danger";
      case api.EnduroPackagePreservationActionResponseBodyStatusEnum.Processing:
        return "bg-warning";
      default:
        return "bg-secondary";
    }
  },
  formatPreservationTaskStatus(
    value: api.EnduroPackagePreservationTaskResponseBodyStatusEnum
  ) {
    switch (value) {
      case api.EnduroPackagePreservationTaskResponseBodyStatusEnum.Complete:
        return "bg-success";
      case api.EnduroPackagePreservationTaskResponseBodyStatusEnum.Failed:
        return "bg-danger";
      case api.EnduroPackagePreservationTaskResponseBodyStatusEnum.Processing:
        return "bg-warning";
      default:
        return "bg-secondary";
    }
  },
};
