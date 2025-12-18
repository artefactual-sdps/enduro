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
import uploader from "@/composables/uploader";
import type { IngestListBatchesStatusEnum } from "@/openapi-generator";
import { useAuthStore } from "@/stores/auth";
import { useBatchStore } from "@/stores/batch";
import { useLayoutStore } from "@/stores/layout";
import { useUserStore } from "@/stores/user";
import IconInfo from "~icons/akar-icons/info-fill";
import IconAll from "~icons/clarity/blocks-group-line?font-size=20px";
import IconClose from "~icons/clarity/close-line";
import IconCanceled from "~icons/clarity/cursor-hand-open-line?font-size=20px";
import IconQueued from "~icons/clarity/hourglass-line?font-size=20px";
import IconBatches from "~icons/clarity/layers-line";
import IconFailed from "~icons/clarity/remove-line?font-size=20px";
import IconSearch from "~icons/clarity/search-line";
import IconIngested from "~icons/clarity/success-standard-line?font-size=20px";
import IconProcessing from "~icons/clarity/sync-line?font-size=20px";
import IconPending from "~icons/clarity/warning-standard-line?font-size=20px";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const batchStore = useBatchStore();
const userStore = useUserStore();

const uploaderEl = ref<HTMLElement | null>(null);
const uploaderDD = ref<Dropdown | null>(null);
const uploaderDDLabel = ref("Ingested by");

const route = useRoute();
const router = useRouter();

layoutStore.updateBreadcrumb([{ text: "Ingest" }, { text: "Batches" }]);

const el = ref<HTMLElement | null>(null);
let tooltip: Tooltip | null = null;

const showLegend = ref(false);
const toggleLegend = () => {
  showLegend.value = !showLegend.value;
  tooltip?.hide();
};

const tabs = computed(() => [
  {
    icon: IconAll,
    text: "All",
    route: router.resolve({
      name: "/ingest/batches/",
      query: { ...route.query, status: undefined, page: undefined },
    }),
    show: true,
  },
  {
    icon: IconIngested,
    text: "Ingested",
    route: router.resolve({
      name: "/ingest/batches/",
      query: {
        ...route.query,
        status: api.EnduroIngestBatchStatusEnum.Ingested,
        page: undefined,
      },
    }),
    show: true,
  },
  {
    icon: IconFailed,
    text: "Failed",
    route: router.resolve({
      name: "/ingest/batches/",
      query: {
        ...route.query,
        status: api.EnduroIngestBatchStatusEnum.Failed,
        page: undefined,
      },
    }),
    show: true,
  },
  {
    icon: IconCanceled,
    text: "Canceled",
    route: router.resolve({
      name: "/ingest/batches/",
      query: {
        ...route.query,
        status: api.EnduroIngestBatchStatusEnum.Canceled,
        page: undefined,
      },
    }),
    show: true,
  },
  {
    icon: IconProcessing,
    text: "Processing",
    route: router.resolve({
      name: "/ingest/batches/",
      query: {
        ...route.query,
        status: api.EnduroIngestBatchStatusEnum.Processing,
        page: undefined,
      },
    }),
    show: true,
  },
  {
    icon: IconQueued,
    text: "Queued",
    route: router.resolve({
      name: "/ingest/batches/",
      query: {
        ...route.query,
        status: api.EnduroIngestBatchStatusEnum.Queued,
        page: undefined,
      },
    }),
    show: true,
  },
  {
    icon: IconPending,
    text: "Pending",
    route: router.resolve({
      name: "/ingest/batches/",
      query: {
        ...route.query,
        status: api.EnduroIngestBatchStatusEnum.Pending,
        page: undefined,
      },
    }),
    show: true,
  },
]);

const statuses = [
  {
    status: api.EnduroIngestBatchStatusEnum.Ingested,
    description: "The batch has successfully completed ingest processing.",
  },
  {
    status: api.EnduroIngestBatchStatusEnum.Failed,
    description:
      "The batch failed to meet the policy-defined criteria for ingest.",
  },
  {
    status: api.EnduroIngestBatchStatusEnum.Processing,
    description:
      "The batch is currently part of an active workflow and is processing.",
  },
  {
    status: api.EnduroIngestBatchStatusEnum.Pending,
    description: "The batch is part of a workflow awaiting a user decision.",
  },
  {
    status: api.EnduroIngestBatchStatusEnum.Queued,
    description:
      "The batch is about to be part of an active workflow and is awaiting processing.",
  },
  {
    status: api.EnduroIngestBatchStatusEnum.Canceled,
    description: "The batch was canceled before it finished processing.",
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
    name: "/ingest/batches/",
    query: q,
  });
};

const searchByIdentifier = () => {
  let q = { ...route.query };
  if (batchStore.filters.identifier === "") {
    delete q.identifier;
  } else {
    q.identifier = <LocationQueryValue>batchStore.filters.identifier;
  }

  // Reset the page number because the found results may reduce the total number
  // of pages.
  delete q.page;

  router.push({
    name: "/ingest/batches/",
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
    name: "/ingest/batches/",
    query: q,
  });
};

const updateUploaderFilter = () => {
  let q = { ...route.query };

  if (batchStore.filters.uploaderId) {
    q.uploaderId = batchStore.filters.uploaderId as LocationQueryValue;
  } else {
    delete q.uploaderId;
  }

  // Reset the page number because the found results may reduce the total number
  // of pages.
  delete q.page;

  router.push({
    name: "/ingest/batches/",
    query: q,
  });
};

const { execute, error } = useAsyncState(() => {
  if (route.query.identifier) {
    batchStore.filters.identifier = <string>route.query.identifier;
  } else {
    delete batchStore.filters.identifier;
  }

  if (route.query.status) {
    batchStore.filters.status = <IngestListBatchesStatusEnum>route.query.status;
  } else {
    delete batchStore.filters.status;
  }

  if (route.query.earliestCreatedTime) {
    batchStore.filters.earliestCreatedTime = new Date(
      route.query.earliestCreatedTime as string,
    );
  } else {
    delete batchStore.filters.earliestCreatedTime;
  }

  if (route.query.latestCreatedTime) {
    batchStore.filters.latestCreatedTime = new Date(
      route.query.latestCreatedTime as string,
    );
  } else {
    delete batchStore.filters.latestCreatedTime;
  }

  if (route.query.uploaderId) {
    batchStore.filters.uploaderId = <string>route.query.uploaderId;
  } else {
    delete batchStore.filters.uploaderId;
  }

  return batchStore.fetchBatches(
    route.query.page ? parseInt(<string>route.query.page) : 1,
  );
}, null);

watch(
  () => route.query,
  () => {
    // Execute fetchBatches when the query changes.
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
    <h1 class="d-flex mb-0"><IconBatches class="me-3 text-dark" />Batches</h1>

    <div class="text-muted mb-3">
      <ResultCounter
        :offset="batchStore.page.offset"
        :limit="batchStore.page.limit"
        :total="batchStore.page.total"
      />
    </div>

    <PageLoadingAlert :execute="execute" :error="error" />

    <div class="d-flex flex-wrap gap-3 mb-3">
      <div>
        <form id="batchSearch" @submit.prevent="searchByIdentifier">
          <div class="input-group">
            <input
              v-model.trim="batchStore.filters.identifier"
              type="text"
              class="form-control"
              name="identifier"
              placeholder="Search by identifier"
              aria-label="Search by identifier"
            />
            <button
              class="btn btn-secondary"
              type="reset"
              aria-label="Reset search"
              @click="
                batchStore.filters.identifier = '';
                searchByIdentifier();
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
            v-show="batchStore.filters.uploaderId !== ''"
            id="dd-uploader-reset"
            class="btn btn-secondary"
            type="reset"
            aria-label="Reset 'Ingested by' filter"
            @click="
              batchStore.filters.uploaderId = '';
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
                  batchStore.filters.uploaderId = user.uuid;
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
          :start="batchStore.filters.earliestCreatedTime"
          :end="batchStore.filters.latestCreatedTime"
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
            <th scope="col">Identifier</th>
            <th scope="col">SIPs</th>
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
                  <IconInfo class="fs-6" aria-hidden="true" />
                  <span class="visually-hidden"
                    >Toggle batch status legend</span
                  >
                </button>
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="batch in batchStore.batches" :key="batch.uuid">
            <td>
              <RouterLink
                v-if="authStore.checkAttributes(['ingest:batches:read'])"
                :to="{
                  name: '/ingest/batches/[id]/',
                  params: { id: batch.uuid },
                }"
              >
                {{ batch.identifier }}
              </RouterLink>
              <span v-else>{{ batch.identifier }}</span>
            </td>
            <td>{{ batch.sipsCount }}</td>
            <td>{{ uploader(batch) }}</td>
            <td>{{ $filters.formatDateTime(batch.startedAt) }}</td>
            <td>
              <StatusBadge :status="batch.status" type="package" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-if="batchStore.page.total > batchStore.page.limit">
      <Pager
        :offset="batchStore.page.offset"
        :limit="batchStore.page.limit"
        :total="batchStore.page.total"
        @page-change="(page) => changePage(page)"
      />
      <div class="text-muted mb-3 text-center">
        <ResultCounter
          :offset="batchStore.page.offset"
          :limit="batchStore.page.limit"
          :total="batchStore.page.total"
        />
      </div>
    </div>
  </div>
</template>
