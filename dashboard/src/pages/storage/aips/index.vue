<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import { computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router/auto";
import type { LocationQueryValue } from "vue-router/auto";

import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import Tabs from "@/components/Tabs.vue";
import TimeDropdown from "@/components/TimeDropdown.vue";
import UUID from "@/components/UUID.vue";
import type { StorageListAipsStatusEnum } from "@/openapi-generator";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import IconCaretLeft from "~icons/bi/caret-left-fill";
import IconCaretRight from "~icons/bi/caret-right-fill";
import IconSkipEnd from "~icons/bi/skip-end-fill";
import IconSkipStart from "~icons/bi/skip-start-fill";
import IconAll from "~icons/clarity/blocks-group-line?raw&font-size=20px";
import IconAIPs from "~icons/clarity/bundle-line";
import IconClose from "~icons/clarity/close-line";
import IconError from "~icons/clarity/remove-line?raw&font-size=20px";
import IconSearch from "~icons/clarity/search-line";
import IconDone from "~icons/clarity/success-standard-line?raw&font-size=20px";
import IconProcessing from "~icons/clarity/sync-line?raw&font-size=20px";
import IconPending from "~icons/clarity/warning-standard-line?raw&font-size=20px";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const aipStore = useAipStore();

const route = useRoute();
const router = useRouter();

layoutStore.updateBreadcrumb([{ text: "Storage" }, { text: "AIPs" }]);

const tabs = computed(() => [
  {
    icon: IconAll,
    text: "All",
    route: router.resolve({
      name: "/storage/aips/",
      query: { ...route.query, status: undefined },
    }),
    show: true,
  },
  {
    icon: IconDone,
    text: "Stored",
    route: router.resolve({
      name: "/storage/aips/",
      query: { ...route.query, status: "stored" },
    }),
    show: true,
  },
  {
    icon: IconError,
    text: "Deleted",
    route: router.resolve({
      name: "/storage/aips/",
      query: { ...route.query, status: "deleted" },
    }),
    show: true,
  },
  {
    icon: IconPending,
    text: "Pending",
    route: router.resolve({
      name: "/storage/aips/",
      query: { ...route.query, status: "pending" },
    }),
    show: true,
  },
  {
    icon: IconProcessing,
    text: "Processing",
    route: router.resolve({
      name: "/storage/aips/",
      query: { ...route.query, status: "processing" },
    }),
    show: true,
  },
]);

const searchByName = () => {
  let q = { ...route.query };
  if (aipStore.filters.name === "") {
    delete q.name;
  } else {
    q.name = <LocationQueryValue>aipStore.filters.name;
  }

  router.push({
    name: "/storage/aips/",
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
    name: "/storage/aips/",
    query: q,
  });
};

const { execute, error } = useAsyncState(() => {
  if (route.query.name) {
    aipStore.filters.name = <string>route.query.name;
  } else {
    delete aipStore.filters.name;
  }

  if (route.query.status) {
    aipStore.filters.status = <StorageListAipsStatusEnum>route.query.status;
  } else {
    delete aipStore.filters.status;
  }

  if (route.query.earliestCreatedTime) {
    aipStore.filters.earliestCreatedTime = new Date(
      route.query.earliestCreatedTime as string,
    );
  } else {
    delete aipStore.filters.earliestCreatedTime;
  }
  if (route.query.latestCreatedTime) {
    aipStore.filters.latestCreatedTime = new Date(
      route.query.latestCreatedTime as string,
    );
  } else {
    delete aipStore.filters.latestCreatedTime;
  }

  return aipStore.fetchAips(
    route.query.page ? parseInt(<string>route.query.page) : 1,
  );
}, null);

watch(
  () => route.query,
  () => {
    // Execute fetchAips when the query changes.
    execute();
  },
);
</script>

<template>
  <div class="container-xxl">
    <h1 class="d-flex mb-0"><IconAIPs class="me-3 text-dark" />AIPs</h1>

    <div class="text-muted mb-3">
      Showing {{ aipStore.page.offset + 1 }} -
      {{ aipStore.lastResultOnPage }} of
      {{ aipStore.page.total }}
    </div>

    <PageLoadingAlert :execute="execute" :error="error" />

    <div class="d-flex flex-wrap gap-3 mb-3">
      <div>
        <form id="sipSearch" @submit.prevent="searchByName">
          <div class="input-group">
            <input
              type="text"
              v-model.trim="aipStore.filters.name"
              class="form-control"
              name="name"
              placeholder="Search by name"
              aria-label="Search by name"
            />
            <button
              class="btn btn-secondary"
              @click="
                aipStore.filters.name = '';
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
          label="Deposited"
          :start="aipStore.filters.earliestCreatedTime"
          :end="aipStore.filters.latestCreatedTime"
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

    <div class="table-responsive mb-3">
      <table class="table table-bordered mb-0">
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">UUID</th>
            <th scope="col">Deposited</th>
            <th scope="col">Location</th>
            <th scope="col">Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="aip in aipStore.aips" :key="aip.uuid">
            <td>
              <router-link
                v-if="authStore.checkAttributes(['storage:aips:read'])"
                :to="{ name: '/storage/aips/[id]/', params: { id: aip.uuid } }"
                >{{ aip.name }}</router-link
              >
              <span v-else>{{ aip.name }}</span>
            </td>
            <td><UUID :id="aip.uuid" /></td>
            <td>{{ $filters.formatDateTime(aip.createdAt) }}</td>
            <td><UUID :id="aip.locationId" /></td>
            <td><StatusBadge :status="aip.status" /></td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-if="aipStore.pager.total > 1">
      <nav role="navigation" aria-label="Pagination navigation">
        <ul class="pagination justify-content-center">
          <li v-if="aipStore.pager.total > aipStore.pager.maxPages">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled: aipStore.pager.current == 1,
              }"
              aria-label="Go to first page"
              title="First page"
              @click.prevent="aipStore.fetchAips(1)"
              ><IconSkipStart
            /></a>
          </li>
          <li class="page-item">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled: !aipStore.hasPrevPage,
              }"
              aria-label="Go to previous page"
              title="Previous page"
              @click.prevent="aipStore.prevPage"
              ><IconCaretLeft
            /></a>
          </li>
          <li
            v-if="aipStore.pager.first > 1"
            class="d-none d-sm-block"
            aria-hidden="true"
          >
            <a href="#" class="page-link disabled">…</a>
          </li>
          <li
            v-for="pg in aipStore.pager.pages"
            :key="pg"
            :class="{ 'd-none d-sm-block': pg != aipStore.pager.current }"
          >
            <a
              href="#"
              :class="{
                'page-link': true,
                active: pg == aipStore.pager.current,
              }"
              @click.prevent="aipStore.fetchAips(pg)"
              :aria-label="
                pg == aipStore.pager.current
                  ? 'Current page, page ' + pg
                  : 'Go to page ' + pg
              "
              :aria-current="pg == aipStore.pager.current"
              >{{ pg }}</a
            >
          </li>
          <li
            v-if="aipStore.pager.last < aipStore.pager.total"
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
                disabled: !aipStore.hasNextPage,
              }"
              aria-label="Go to next page"
              title="Next page"
              @click.prevent="aipStore.nextPage"
              ><IconCaretRight
            /></a>
          </li>
          <li v-if="aipStore.pager.total > aipStore.pager.maxPages">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled: aipStore.pager.current == aipStore.pager.total,
              }"
              aria-label="Go to last page"
              title="Last page"
              @click.prevent="aipStore.fetchAips(aipStore.pager.total)"
              ><IconSkipEnd
            /></a>
          </li>
        </ul>
      </nav>
      <div class="text-muted mb-3 text-center">
        Showing AIPs {{ aipStore.page.offset + 1 }} -
        {{ aipStore.lastResultOnPage }} of
        {{ aipStore.page.total }}
      </div>
    </div>
  </div>
</template>
