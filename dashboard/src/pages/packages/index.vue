<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import Tooltip from "bootstrap/js/dist/tooltip";
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router/auto";
import type { LocationQueryValue } from "vue-router/auto";

import PackageListLegend from "@/components/PackageListLegend.vue";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import Tabs from "@/components/Tabs.vue";
import TimeDropdown from "@/components/TimeDropdown.vue";
import UUID from "@/components/UUID.vue";
import type { PackageListStatusEnum } from "@/openapi-generator";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { usePackageStore } from "@/stores/package";
import IconInfoFill from "~icons/akar-icons/info-fill";
import IconCaretLeftFill from "~icons/bi/caret-left-fill";
import IconCaretRightFill from "~icons/bi/caret-right-fill";
import IconSkipEndFill from "~icons/bi/skip-end-fill";
import IconSkipStartFill from "~icons/bi/skip-start-fill";
import RawIconBlocksGroupLine from "~icons/clarity/blocks-group-line?raw&font-size=20px";
import IconBundleLine from "~icons/clarity/bundle-line";
import RawIconClockLine from "~icons/clarity/clock-line?raw&font-size=20px";
import IconCloseLine from "~icons/clarity/close-line";
import RawIconRemoveLine from "~icons/clarity/remove-line?raw&font-size=20px";
import IconSearch from "~icons/clarity/search-line";
import RawIconSuccessLine from "~icons/clarity/success-standard-line?raw&font-size=20px";
import RawIconSyncLine from "~icons/clarity/sync-line?raw&font-size=20px";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const packageStore = usePackageStore();

const route = useRoute();
const router = useRouter();

layoutStore.updateBreadcrumb([{ text: "Packages" }]);

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
    icon: RawIconBlocksGroupLine,
    text: "All",
    route: router.resolve({
      name: "/packages/",
      query: { ...route.query, status: undefined },
    }),
    show: true,
  },
  {
    icon: RawIconSuccessLine,
    text: "Done",
    route: router.resolve({
      name: "/packages/",
      query: { ...route.query, status: "done" },
    }),
    show: true,
  },
  {
    icon: RawIconRemoveLine,
    text: "Error",
    route: router.resolve({
      name: "/packages/",
      query: { ...route.query, status: "error" },
    }),
    show: true,
  },
  {
    icon: RawIconSyncLine,
    text: "In progress",
    route: router.resolve({
      name: "/packages/",
      query: { ...route.query, status: "in progress" },
    }),
    show: true,
  },
  {
    icon: RawIconClockLine,
    text: "Queued",
    route: router.resolve({
      name: "/packages/",
      query: { ...route.query, status: "queued" },
    }),
    show: true,
  },
]);

const doSearch = () => {
  let q = { ...route.query };
  if (packageStore.filters.name === "") {
    delete q.name;
  } else {
    q.name = packageStore.filters.name;
  }

  router.push({
    name: "/packages/",
    query: q,
  });
};

const updateDateFilter = (name: string, value: LocationQueryValue) => {
  let q = { ...route.query };
  if (value === null || value === "") {
    delete q[name];
  } else {
    q[name] = value;
  }

  router.push({
    name: "/packages/",
    query: q,
  });
};

const { execute, error } = useAsyncState(() => {
  if (route.query.name) {
    packageStore.filters.name = <string>route.query.name;
  }
  if (route.query.status) {
    packageStore.filters.status = <PackageListStatusEnum>route.query.status;
  }
  if (route.query.earliestCreatedTime) {
    packageStore.filters.earliestCreatedTime = new Date(
      route.query.earliestCreatedTime as string,
    );
  }

  return packageStore.fetchPackages(1);
}, null);

watch(
  () => [route.query.status, route.query.name, route.query.earliestCreatedTime],
  ([newStatus, newName, newEarliest]) => {
    packageStore.filters.status = newStatus as PackageListStatusEnum;

    if (newName) {
      packageStore.filters.name = newName as string;
    }

    if (newEarliest) {
      packageStore.filters.earliestCreatedTime = new Date(
        newEarliest as string,
      );
    } else {
      packageStore.filters.earliestCreatedTime = undefined;
    }

    return packageStore.fetchPackages(1);
  },
);
</script>

<template>
  <div class="container-xxl">
    <h1 class="d-flex mb-0">
      <IconBundleLine class="me-3 text-dark" />Packages
    </h1>

    <div class="text-muted mb-3">
      Showing {{ packageStore.page.offset + 1 }} -
      {{ packageStore.lastResultOnPage }} of
      {{ packageStore.page.total }}
    </div>

    <PageLoadingAlert :execute="execute" :error="error" />

    <div class="d-flex flex-wrap gap-3 mb-3">
      <div>
        <form id="packageSearch" @submit.prevent="doSearch">
          <div class="input-group">
            <input
              type="text"
              v-model.trim="packageStore.filters.name"
              class="form-control"
              name="name"
              placeholder="Search by name"
              aria-label="Search by name"
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
              <IconCloseLine />
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
          fieldname="earliestCreatedTime"
          @change="
            (name: string, value: LocationQueryValue) =>
              updateDateFilter(name, value)
          "
        />
      </div>
    </div>

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
                  <IconInfoFill style="font-size: 1.2em" aria-hidden="true" />
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
                :to="{ name: '/packages/[id]/', params: { id: pkg.id } }"
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
              ><IconSkipStartFill
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
              ><IconCaretLeftFill
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
            :key="pg"
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
              ><IconCaretRightFill
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
              ><IconSkipEndFill
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
