import App from "./App.vue";
import { api } from "./client";
import "./styles/main.scss";
import humanizeDuration from "humanize-duration";
import moment from "moment";
import { createPinia } from "pinia";
import { createApp } from "vue";
import { PromiseDialog } from "vue3-promise-dialog";
import { createRouter, createWebHistory } from "vue-router";
import routes from "~pages";

const router = createRouter({
  history: createWebHistory("/"),
  routes,
  strict: false,
});

const app = createApp(App);
app.use(router);
app.use(createPinia());
app.use(PromiseDialog);
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
    value: api.EnduroPackagePreservationActionsActionResponseBodyStatusEnum
  ) {
    switch (value) {
      case api.EnduroPackagePreservationActionsActionResponseBodyStatusEnum
        .Complete:
        return "bg-success";
      case api.EnduroPackagePreservationActionsActionResponseBodyStatusEnum
        .Failed:
        return "bg-danger";
      case api.EnduroPackagePreservationActionsActionResponseBodyStatusEnum
        .Processing:
        return "bg-warning";
      default:
        return "bg-secondary";
    }
  },
};
