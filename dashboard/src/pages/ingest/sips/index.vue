<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import Dropdown from "bootstrap/js/dist/dropdown";
import Tooltip from "bootstrap/js/dist/tooltip";
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router/auto";
import type { LocationQueryValue } from "vue-router/auto";

import { api } from "@/client";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import Pager from "@/components/Pager.vue";
import ResultCounter from "@/components/ResultCounter.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import StatusLegend from "@/components/StatusLegend.vue";
import Tabs from "@/components/Tabs.vue";
import TimeDropdown from "@/components/TimeDropdown.vue";
import uploader from "@/composables/sipUploader";
import type { IngestListSipsStatusEnum } from "@/openapi-generator";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { useSipStore } from "@/stores/sip";
import { useUserStore } from "@/stores/user";
import IconInfo from "~icons/akar-icons/info-fill";
import IconAll from "~icons/clarity/blocks-group-line?font-size=20px";
import IconClose from "~icons/clarity/close-line";
import IconError from "~icons/clarity/flame-line?font-size=20px";
import IconQueued from "~icons/clarity/hourglass-line?font-size=20px";
import IconFailed from "~icons/clarity/remove-line?font-size=20px";
import IconSearch from "~icons/clarity/search-line";
import IconIngested from "~icons/clarity/success-standard-line?font-size=20px";
import IconProcessing from "~icons/clarity/sync-line?font-size=20px";
import IconPending from "~icons/clarity/warning-standard-line?font-size=20px";
import IconSIPs from "~icons/octicon/package-dependencies-24";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const sipStore = useSipStore();
const userStore = useUserStore();

const uploaderEl = ref<HTMLElement | null>(null);
const uploaderDD = ref<Dropdown | null>(null);
const uploaderDDLabel = ref("Ingested by");

const route = useRoute();
const router = useRouter();

layoutStore.updateBreadcrumb([{ text: "Ingest" }, { text: "SIPs" }]);

const el = ref<HTMLElement | null>(null);
let tooltip: Tooltip | null = null;

let showLegend = ref(false);
const toggleLegend = () => {
  showLegend.value = !showLegend.value;
  if (tooltip) tooltip.hide();
};

const tabs = computed(() => [
  {
    icon: IconAll,
    text: "All",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: undefined, page: undefined },
    }),
    show: true,
  },
  {
    icon: IconIngested,
    text: "Ingested",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: "ingested", page: undefined },
    }),
    show: true,
  },
  {
    icon: IconFailed,
    text: "Failed",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: "failed", page: undefined },
    }),
    show: true,
  },
  {
    icon: IconError,
    text: "Error",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: "error", page: undefined },
    }),
    show: true,
  },
  {
    icon: IconProcessing,
    text: "Processing",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: "processing", page: undefined },
    }),
    show: true,
  },
  {
    icon: IconQueued,
    text: "Queued",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: "queued", page: undefined },
    }),
    show: true,
  },
  {
    icon: IconPending,
    text: "Pending",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: "pending", page: undefined },
    }),
    show: true,
  },
]);

const statuses = [
  {
    status: api.EnduroIngestSipStatusEnum.Ingested,
    description: "The SIP has successfully completed all ingest processing.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Failed,
    description:
      "The SIP has failed to meet the policy-defined criteria for ingest, halting the workflow.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Processing,
    description:
      "The SIP is currently part of an active workflow and is undergoing processing.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Pending,
    description: "The SIP is part of a workflow awaiting a user decision.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Queued,
    description:
      "The SIP is about to be part of an active workflow and is awaiting processing.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Error,
    description:
      "The SIP workflow encountered a system error and ingest was aborted.",
  },
];

const changePage = (page: number) => {
  let q = { ...route.query };
  if (page <= 1) {
    delete q.page;
  } else {
    q.page = <LocationQueryValue>page.toString();
  }

  router.push({
    name: "/ingest/sips/",
    query: q,
  });
};

const searchByName = () => {
  let q = { ...route.query };
  if (sipStore.filters.name === "") {
    delete q.name;
  } else {
    q.name = <LocationQueryValue>sipStore.filters.name;
  }

  // Reset the page number because the found results may reduce the total number
  // of pages.
  delete q.page;

  router.push({
    name: "/ingest/sips/",
    query: q,
  });
};

const updateCreatedAtFilter = (
  q: { [x: string]: LocationQueryValue | LocationQueryValue[] },
  start: LocationQueryValue,
  end: LocationQueryValue,
): { [x: string]: LocationQueryValue | LocationQueryValue[] } => {
  if (start) {
    q.earliestCreatedTime = start;
  } else {
    delete q.earliestCreatedTime;
  }

  if (end) {
    q.latestCreatedTime = end;
  } else {
    delete q.latestCreatedTime;
  }

  return q;
};

const updateDateFilter = (
  name: string,
  start: LocationQueryValue,
  end: LocationQueryValue,
) => {
  let q = { ...route.query };

  switch (name) {
    case "createdAt":
      q = updateCreatedAtFilter(q, start, end);
      break;
    default:
      // undefined.
      return;
  }

  // Reset the page number because the found results may reduce the total number
  // of pages.
  delete q.page;

  router.push({
    name: "/ingest/sips/",
    query: q,
  });
};

const updateUploaderFilter = () => {
  let q = { ...route.query };

  if (sipStore.filters.uploaderId) {
    q.uploaderId = sipStore.filters.uploaderId as LocationQueryValue;
  } else {
    delete q.uploaderId;
  }

  // Reset the page number because the found results may reduce the total number
  // of pages.
  delete q.page;

  router.push({
    name: "/ingest/sips/",
    query: q,
  });
};

const { execute, error } = useAsyncState(() => {
  if (route.query.name) {
    sipStore.filters.name = <string>route.query.name;
  } else {
    delete sipStore.filters.name;
  }

  if (route.query.status) {
    sipStore.filters.status = <IngestListSipsStatusEnum>route.query.status;
  } else {
    delete sipStore.filters.status;
  }

  if (route.query.earliestCreatedTime) {
    sipStore.filters.earliestCreatedTime = new Date(
      route.query.earliestCreatedTime as string,
    );
  } else {
    delete sipStore.filters.earliestCreatedTime;
  }

  if (route.query.latestCreatedTime) {
    sipStore.filters.latestCreatedTime = new Date(
      route.query.latestCreatedTime as string,
    );
  } else {
    delete sipStore.filters.latestCreatedTime;
  }

  if (route.query.uploaderId) {
    sipStore.filters.uploaderId = <string>route.query.uploaderId;
  } else {
    delete sipStore.filters.uploaderId;
  }

  return sipStore.fetchSips(
    route.query.page ? parseInt(<string>route.query.page) : 1,
  );
}, null);

watch(
  () => route.query,
  () => {
    // Execute fetchSips when the query changes.
    execute();
  },
);

onMounted(() => {
  if (el.value) tooltip = new Tooltip(el.value);
  if (uploaderEl.value) uploaderDD.value = new Dropdown(uploaderEl.value);

  // Fetch the users for the uploader dropdown.
  if (authStore.checkAttributes(["ingest:users:list"])) {
    userStore.fetchUsers().catch((err) => {
      console.error("fetch users:", err);
      error.value = "Failed to fetch users";
    });
  }
});
</script>

<template>
  <div class="container-xxl">
    <h1 class="d-flex mb-0"><IconSIPs class="me-3 text-dark" />SIPs</h1>

    <div class="text-muted mb-3">
      <ResultCounter
        :offset="sipStore.page.offset"
        :limit="sipStore.page.limit"
        :total="sipStore.page.total"
      />
    </div>

    <PageLoadingAlert :execute="execute" :error="error" />

    <div class="d-flex flex-wrap gap-3 mb-3">
      <div>
        <form id="sipSearch" @submit.prevent="searchByName">
          <div class="input-group">
            <input
              v-model.trim="sipStore.filters.name"
              type="text"
              class="form-control"
              name="name"
              placeholder="Search by name"
              aria-label="Search by name"
            />
            <button
              class="btn btn-secondary"
              type="reset"
              aria-label="Reset search"
              @click="
                sipStore.filters.name = '';
                searchByName();
              "
            >
              <IconClose />
            </button>
            <button
              class="btn btn-primary"
              type="submit"
              aria-label="Do search"
            >
              <IconSearch />
            </button>
          </div>
        </form>
      </div>
      <div
        v-if="
          authStore.checkAttributes(['ingest:users:list']) && userStore.hasUsers
        "
      >
        <div ref="uploaderEl" class="dropdown">
          <button
            id="dd-uploader-button"
            class="btn btn-primary dropdown-toggle"
            type="button"
            data-bs-toggle="dropdown"
            aria-label="Toggle 'Ingested by' filter dropdown"
            aria-expanded="false"
          >
            {{ uploaderDDLabel }}
          </button>
          <button
            v-show="sipStore.filters.uploaderId !== ''"
            id="dd-uploader-reset"
            class="btn btn-secondary"
            type="reset"
            aria-label="Reset 'Ingested by' filter"
            @click="
              sipStore.filters.uploaderId = '';
              uploaderDDLabel = 'Ingested by';
              updateUploaderFilter();
            "
          >
            <IconClose />
          </button>
          <ul class="dropdown-menu">
            <li
              v-for="user in userStore.users"
              :key="user.uuid"
              :value="user.uuid"
            >
              <a
                class="dropdown-item"
                href="#"
                @click.prevent="
                  sipStore.filters.uploaderId = user.uuid;
                  uploaderDDLabel = userStore.getHandle(user);
                  updateUploaderFilter();
                "
                >{{ userStore.getHandle(user) }}</a
              >
            </li>
          </ul>
        </div>
      </div>
      <div>
        <TimeDropdown
          name="createdAt"
          label="Started"
          :start="sipStore.filters.earliestCreatedTime"
          :end="sipStore.filters.latestCreatedTime"
          @change="
            (
              name: string,
              start: LocationQueryValue,
              end: LocationQueryValue,
            ) => updateDateFilter(name, start, end)
          "
        />
      </div>
    </div>

    <Tabs :tabs="tabs" param="status" />
    <StatusLegend
      :show="showLegend"
      :items="statuses"
      @update:show="(val) => (showLegend = val)"
    />

    <div class="table-responsive mb-3">
      <table class="table table-bordered mb-0">
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">Ingested by</th>
            <th scope="col">Started</th>
            <th scope="col">
              <span class="d-flex gap-2">
                Status
                <button
                  ref="el"
                  class="btn btn-sm btn-link text-decoration-none ms-auto p-0"
                  type="button"
                  data-bs-toggle="tooltip"
                  data-bs-title="Toggle legend"
                  @click="toggleLegend"
                >
                  <IconInfo style="font-size: 1.2em" aria-hidden="true" />
                  <span class="visually-hidden">Toggle SIP status legend</span>
                </button>
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="sip in sipStore.sips" :key="sip.uuid">
            <td>
              <RouterLink
                v-if="authStore.checkAttributes(['ingest:sips:read'])"
                :to="{ name: '/ingest/sips/[id]/', params: { id: sip.uuid } }"
              >
                {{ sip.name }}
              </RouterLink>
              <span v-else>{{ sip.name }}</span>
            </td>
            <td>{{ uploader(sip) }}</td>
            <td>{{ $filters.formatDateTime(sip.startedAt) }}</td>
            <td>
              <StatusBadge :status="sip.status" type="package" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-if="sipStore.page.total > sipStore.page.limit">
      <Pager
        :offset="sipStore.page.offset"
        :limit="sipStore.page.limit"
        :total="sipStore.page.total"
        @page-change="(page) => changePage(page)"
      />
      <div class="text-muted mb-3 text-center">
        <ResultCounter
          :offset="sipStore.page.offset"
          :limit="sipStore.page.limit"
          :total="sipStore.page.total"
        />
      </div>
    </div>
  </div>
</template>
