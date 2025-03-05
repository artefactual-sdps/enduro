<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import Tooltip from "bootstrap/js/dist/tooltip";
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router/auto";
import type { LocationQueryValue } from "vue-router/auto";

import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import SipListLegend from "@/components/SipListLegend.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import Tabs from "@/components/Tabs.vue";
import TimeDropdown from "@/components/TimeDropdown.vue";
import UUID from "@/components/UUID.vue";
import type { IngestListSipsStatusEnum } from "@/openapi-generator";
import { useAuthStore } from "@/stores/auth";
import { useIngestStore } from "@/stores/ingest";
import { useLayoutStore } from "@/stores/layout";
import IconInfo from "~icons/akar-icons/info-fill";
import IconCaretLeft from "~icons/bi/caret-left-fill";
import IconCaretRight from "~icons/bi/caret-right-fill";
import IconSkipEnd from "~icons/bi/skip-end-fill";
import IconSkipStart from "~icons/bi/skip-start-fill";
import IconAll from "~icons/clarity/blocks-group-line?raw&font-size=20px";
import IconQueued from "~icons/clarity/clock-line?raw&font-size=20px";
import IconClose from "~icons/clarity/close-line";
import IconError from "~icons/clarity/remove-line?raw&font-size=20px";
import IconSearch from "~icons/clarity/search-line";
import IconDone from "~icons/clarity/success-standard-line?raw&font-size=20px";
import IconInProgress from "~icons/clarity/sync-line?raw&font-size=20px";
import IconSIPs from "~icons/octicon/package-dependencies-24";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const ingestStore = useIngestStore();

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

onMounted(() => {
  if (el.value) tooltip = new Tooltip(el.value);
});

const tabs = computed(() => [
  {
    icon: IconAll,
    text: "All",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: undefined },
    }),
    show: true,
  },
  {
    icon: IconDone,
    text: "Done",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: "done" },
    }),
    show: true,
  },
  {
    icon: IconError,
    text: "Error",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: "error" },
    }),
    show: true,
  },
  {
    icon: IconInProgress,
    text: "In progress",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: "in progress" },
    }),
    show: true,
  },
  {
    icon: IconQueued,
    text: "Queued",
    route: router.resolve({
      name: "/ingest/sips/",
      query: { ...route.query, status: "queued" },
    }),
    show: true,
  },
]);

const searchByName = () => {
  let q = { ...route.query };
  if (ingestStore.filters.name === "") {
    delete q.name;
  } else {
    q.name = <LocationQueryValue>ingestStore.filters.name;
  }

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

  router.push({
    name: "/ingest/sips/",
    query: q,
  });
};

const { execute, error } = useAsyncState(() => {
  if (route.query.name) {
    ingestStore.filters.name = <string>route.query.name;
  } else {
    delete ingestStore.filters.name;
  }

  if (route.query.status) {
    ingestStore.filters.status = <IngestListSipsStatusEnum>route.query.status;
  } else {
    delete ingestStore.filters.status;
  }

  if (route.query.earliestCreatedTime) {
    ingestStore.filters.earliestCreatedTime = new Date(
      route.query.earliestCreatedTime as string,
    );
  } else {
    delete ingestStore.filters.earliestCreatedTime;
  }

  if (route.query.latestCreatedTime) {
    ingestStore.filters.latestCreatedTime = new Date(
      route.query.latestCreatedTime as string,
    );
  } else {
    delete ingestStore.filters.latestCreatedTime;
  }

  return ingestStore.fetchSips(
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
</script>

<template>
  <div class="container-xxl">
    <h1 class="d-flex mb-0"><IconSIPs class="me-3 text-dark" />SIPs</h1>

    <div class="text-muted mb-3">
      Showing {{ ingestStore.page.offset + 1 }} -
      {{ ingestStore.lastResultOnPage }} of
      {{ ingestStore.page.total }}
    </div>

    <PageLoadingAlert :execute="execute" :error="error" />

    <div class="d-flex flex-wrap gap-3 mb-3">
      <div>
        <form id="sipSearch" @submit.prevent="searchByName">
          <div class="input-group">
            <input
              type="text"
              v-model.trim="ingestStore.filters.name"
              class="form-control"
              name="name"
              placeholder="Search by name"
              aria-label="Search by name"
            />
            <button
              class="btn btn-secondary"
              @click="
                ingestStore.filters.name = '';
                searchByName();
              "
              type="reset"
              aria-label="Reset search"
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
      <div>
        <TimeDropdown
          name="createdAt"
          label="Started"
          :start="ingestStore.filters.earliestCreatedTime"
          :end="ingestStore.filters.latestCreatedTime"
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
    <SipListLegend v-model="showLegend" />

    <div class="table-responsive mb-3">
      <table class="table table-bordered mb-0">
        <thead>
          <tr>
            <th scope="col">ID</th>
            <th scope="col">Name</th>
            <th scope="col">UUID</th>
            <th scope="col">Started</th>
            <th scope="col">Location</th>
            <th scope="col">
              <span class="d-flex gap-2">
                Status
                <button
                  ref="el"
                  class="btn btn-sm btn-link text-decoration-none ms-auto p-0"
                  type="button"
                  @click="toggleLegend"
                  data-bs-toggle="tooltip"
                  data-bs-title="Toggle legend"
                >
                  <IconInfo style="font-size: 1.2em" aria-hidden="true" />
                  <span class="visually-hidden">Toggle SIP status legend</span>
                </button>
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="pkg in ingestStore.sips" :key="pkg.id">
            <td scope="row">{{ pkg.id }}</td>
            <td>
              <router-link
                v-if="authStore.checkAttributes(['ingest:sips:read'])"
                :to="{ name: '/ingest/sips/[id]/', params: { id: pkg.id } }"
                >{{ pkg.name }}</router-link
              >
              <span v-else>{{ pkg.name }}</span>
            </td>
            <td>
              <UUID :id="pkg.aipId" />
            </td>
            <td>{{ $filters.formatDateTime(pkg.startedAt) }}</td>
            <td>
              <UUID :id="pkg.locationId" />
            </td>
            <td>
              <StatusBadge :status="pkg.status" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-if="ingestStore.pager.total > 1">
      <nav role="navigation" aria-label="Pagination navigation">
        <ul class="pagination justify-content-center">
          <li v-if="ingestStore.pager.total > ingestStore.pager.maxPages">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled: ingestStore.pager.current == 1,
              }"
              aria-label="Go to first page"
              title="First page"
              @click.prevent="ingestStore.fetchSips(1)"
              ><IconSkipStart
            /></a>
          </li>
          <li class="page-item">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled: !ingestStore.hasPrevPage,
              }"
              aria-label="Go to previous page"
              title="Previous page"
              @click.prevent="ingestStore.prevPage"
              ><IconCaretLeft
            /></a>
          </li>
          <li
            v-if="ingestStore.pager.first > 1"
            class="d-none d-sm-block"
            aria-hidden="true"
          >
            <a href="#" class="page-link disabled">…</a>
          </li>
          <li
            v-for="pg in ingestStore.pager.pages"
            :key="pg"
            :class="{ 'd-none d-sm-block': pg != ingestStore.pager.current }"
          >
            <a
              href="#"
              :class="{
                'page-link': true,
                active: pg == ingestStore.pager.current,
              }"
              @click.prevent="ingestStore.fetchSips(pg)"
              :aria-label="
                pg == ingestStore.pager.current
                  ? 'Current page, page ' + pg
                  : 'Go to page ' + pg
              "
              :aria-current="pg == ingestStore.pager.current"
              >{{ pg }}</a
            >
          </li>
          <li
            v-if="ingestStore.pager.last < ingestStore.pager.total"
            class="d-none d-sm-block"
            aria-hidden="true"
          >
            <a href="#" class="page-link disabled">…</a>
          </li>
          <li class="page-item">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled: !ingestStore.hasNextPage,
              }"
              aria-label="Go to next page"
              title="Next page"
              @click.prevent="ingestStore.nextPage"
              ><IconCaretRight
            /></a>
          </li>
          <li v-if="ingestStore.pager.total > ingestStore.pager.maxPages">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled: ingestStore.pager.current == ingestStore.pager.total,
              }"
              aria-label="Go to last page"
              title="Last page"
              @click.prevent="ingestStore.fetchSips(ingestStore.pager.total)"
              ><IconSkipEnd
            /></a>
          </li>
        </ul>
      </nav>
      <div class="text-muted mb-3 text-center">
        Showing SIPs {{ ingestStore.page.offset + 1 }} -
        {{ ingestStore.lastResultOnPage }} of
        {{ ingestStore.page.total }}
      </div>
    </div>
  </div>
</template>
