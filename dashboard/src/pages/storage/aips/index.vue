<script setup lang="ts">
import type { PackageListStatusEnum } from "@/openapi-generator";

import PackageListLegend from "@/components/PackageListLegend.vue";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import Tabs from "@/components/Tabs.vue";
import Tooltip from "bootstrap/js/dist/tooltip";
import UUID from "@/components/UUID.vue";
import { onMounted } from "vue";
import { useAsyncState } from "@vueuse/core";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { usePackageStore } from "@/stores/package";
import { useRoute, useRouter, type LocationQueryValue } from "vue-router/auto";
import { computed, ref, watch } from "vue";

// General icons.
import IconInfo from "~icons/akar-icons/info-fill";
import IconSIPs from "~icons/octicon/package-dependencies-24";
import IconSearch from "~icons/clarity/search-line";
import IconClose from "~icons/clarity/close-line";

// Pager icons.
import IconSkipEnd from "~icons/bi/skip-end-fill";
import IconSkipStart from "~icons/bi/skip-start-fill";
import IconCaretRight from "~icons/bi/caret-right-fill";
import IconCaretLeft from "~icons/bi/caret-left-fill";

// Tab icons.
import IconAll from "~icons/clarity/blocks-group-line?raw&font-size=20px";
import IconDone from "~icons/clarity/success-standard-line?raw&font-size=20px";
import IconError from "~icons/clarity/remove-line?raw&font-size=20px";
import IconInProgress from "~icons/clarity/sync-line?raw&font-size=20px";
import IconQueued from "~icons/clarity/clock-line?raw&font-size=20px";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const packageStore = usePackageStore();

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

const doSearch = () => {
  router.push({
    name: "/ingest/sips/",
    query: { ...route.query, name: packageStore.filters.name },
  });
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

const { execute, error } = useAsyncState(() => {
  if (route.query.name) {
    packageStore.filters.name = <string>route.query.name;
  }
  if (route.query.status) {
    packageStore.filters.status = <PackageListStatusEnum>route.query.status;
  }
  return packageStore.fetchPackages(1);
}, null);

watch(
  () => route.query.status,
  (newStatus) => {
    packageStore.filters.status = newStatus as PackageListStatusEnum;
    return packageStore.fetchPackages(1);
  },
);

watch(
  () => route.query.name,
  (newSearch) => {
    packageStore.filters.name = newSearch as string;
    return packageStore.fetchPackages(1);
  },
);
</script>

<template>
  <div class="container-xxl">
    <h1 class="d-flex mb-0"><IconSIPs class="me-3 text-dark" />SIPs</h1>

    <div class="text-muted mb-3">
      Showing {{ packageStore.page.offset + 1 }} -
      {{ packageStore.lastResultOnPage }} of
      {{ packageStore.page.total }}
    </div>

    <PageLoadingAlert :execute="execute" :error="error" />

    <form id="packageSearch" @submit.prevent="doSearch">
      <div class="input-group w-50 mb-3">
        <input
          type="text"
          v-model.trim="packageStore.filters.name"
          class="form-control"
          name="name"
          placeholder="Search"
          aria-label="Package name"
        />
        <button
          class="btn btn-secondary"
          @click="
            packageStore.filters.name = '';
            doSearch();
          "
          type="reset"
          aria-label="Reset search"
        >
          <IconClose />
        </button>
        <button class="btn btn-primary" type="submit" aria-label="Do search">
          <IconSearch />
        </button>
      </div>
    </form>

    <Tabs :tabs="tabs" param="status" />
    <PackageListLegend v-model="showLegend" />

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
                  <span class="visually-hidden"
                    >Toggle package status legend</span
                  >
                </button>
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="pkg in packageStore.packages" :key="pkg.id">
            <td scope="row">{{ pkg.id }}</td>
            <td>
              <router-link
                v-if="authStore.checkAttributes(['package:read'])"
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
    <div v-if="packageStore.pager.total > 1">
      <nav role="navigation" aria-label="Pagination navigation">
        <ul class="pagination justify-content-center">
          <li v-if="packageStore.pager.total > packageStore.pager.maxPages">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled: packageStore.pager.current == 1,
              }"
              aria-label="Go to first page"
              title="First page"
              @click.prevent="packageStore.fetchPackages(1)"
              ><IconSkipStart
            /></a>
          </li>
          <li class="page-item">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled: !packageStore.hasPrevPage,
              }"
              aria-label="Go to previous page"
              title="Previous page"
              @click.prevent="packageStore.prevPage"
              ><IconCaretLeft
            /></a>
          </li>
          <li
            v-if="packageStore.pager.first > 1"
            class="d-none d-sm-block"
            aria-hidden="true"
          >
            <a href="#" class="page-link disabled">…</a>
          </li>
          <li
            v-for="pg in packageStore.pager.pages"
            :class="{ 'd-none d-sm-block': pg != packageStore.pager.current }"
          >
            <a
              href="#"
              :class="{
                'page-link': true,
                active: pg == packageStore.pager.current,
              }"
              @click.prevent="packageStore.fetchPackages(pg)"
              :aria-label="
                pg == packageStore.pager.current
                  ? 'Current page, page ' + pg
                  : 'Go to page ' + pg
              "
              :aria-current="pg == packageStore.pager.current"
              >{{ pg }}</a
            >
          </li>
          <li
            v-if="packageStore.pager.last < packageStore.pager.total"
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
                disabled: !packageStore.hasNextPage,
              }"
              aria-label="Go to next page"
              title="Next page"
              @click.prevent="packageStore.nextPage"
              ><IconCaretRight
            /></a>
          </li>
          <li v-if="packageStore.pager.total > packageStore.pager.maxPages">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled:
                  packageStore.pager.current == packageStore.pager.total,
              }"
              aria-label="Go to last page"
              title="Last page"
              @click.prevent="
                packageStore.fetchPackages(packageStore.pager.total)
              "
              ><IconSkipEnd
            /></a>
          </li>
        </ul>
      </nav>
      <div class="text-muted mb-3 text-center">
        Showing packages {{ packageStore.page.offset + 1 }} -
        {{ packageStore.lastResultOnPage }} of
        {{ packageStore.page.total }}
      </div>
    </div>
  </div>
</template>
