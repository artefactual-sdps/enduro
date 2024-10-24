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
import { useRoute, useRouter } from "vue-router/auto";
import { watch } from "vue";

// General icons.
import IconInfoFill from "~icons/akar-icons/info-fill";
import IconBundleLine from "~icons/clarity/bundle-line";

// Pager icons.
import IconSkipEndFill from "~icons/bi/skip-end-fill";
import IconSkipStartFill from "~icons/bi/skip-start-fill";
import IconCaretRightFill from "~icons/bi/caret-right-fill";
import IconCaretLeftFill from "~icons/bi/caret-left-fill";

// Tab icons.
import RawIconCheckCircleLine from "~icons/clarity/check-circle-line?raw&font-size=20px";
import RawIconTimesCircleLine from "~icons/clarity/times-circle-line?raw&font-size=20px";
import RawIconPlayLine from "~icons/clarity/play-line?raw&font-size=20px";
import RawIconBarsLine from "~icons/clarity/bars-line?raw&font-size=20px";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const packageStore = usePackageStore();

const route = useRoute("/packages/");
const router = useRouter();

layoutStore.updateBreadcrumb([{ text: "Packages" }]);

const el = $ref<HTMLElement | null>(null);
let tooltip: Tooltip | null = null;

let showLegend = $ref(false);
const toggleLegend = () => {
  showLegend = !showLegend;
  if (tooltip) tooltip.hide();
};

onMounted(() => {
  if (el) tooltip = new Tooltip(el);
  packageStore.filters.status = <PackageListStatusEnum>route.query.status;
});

const tabs = [
  {
    text: "All",
    route: router.resolve({
      name: "/packages/",
    }),
    show: true,
  },
  {
    icon: RawIconCheckCircleLine,
    text: "Done",
    route: router.resolve({
      name: "/packages/",
      query: { status: "done" },
    }),
    show: true,
  },
  {
    icon: RawIconPlayLine,
    text: "Error",
    route: router.resolve({
      name: "/packages/",
      query: { status: "error" },
    }),
    show: true,
  },
  {
    icon: RawIconTimesCircleLine,
    text: "In progress",
    route: router.resolve({
      name: "/packages/",
      query: { status: "in progress" },
    }),
    show: true,
  },
  {
    icon: RawIconBarsLine,
    text: "Queued",
    route: router.resolve({
      name: "/packages/",
      query: { status: "queued" },
    }),
    show: true,
  },
];

const { execute, error } = useAsyncState(() => {
  return packageStore.fetchPackages(1);
}, null);

watch(
  () => route.query.status,
  (newStatus) => {
    packageStore.filters.status = newStatus as PackageListStatusEnum;
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
    <PackageListLegend v-model="showLegend" />

    <Tabs :tabs="tabs" />
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
