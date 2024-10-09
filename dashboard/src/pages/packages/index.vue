<script setup lang="ts">
import PackageListLegend from "@/components/PackageListLegend.vue";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import UUID from "@/components/UUID.vue";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { usePackageStore } from "@/stores/package";
import { useAsyncState } from "@vueuse/core";
import Tooltip from "bootstrap/js/dist/tooltip";
import { onMounted } from "vue";
import IconInfoFill from "~icons/akar-icons/info-fill";
import IconBundleLine from "~icons/clarity/bundle-line";
import IconSkipEndFill from "~icons/bi/skip-end-fill";
import IconSkipStartFill from "~icons/bi/skip-start-fill";
import IconCaretRightFill from "~icons/bi/caret-right-fill";
import IconCaretLeftFill from "~icons/bi/caret-left-fill";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
layoutStore.updateBreadcrumb([{ text: "Packages" }]);

const packageStore = usePackageStore();

const { execute, error } = useAsyncState(() => {
  return packageStore.fetchPackages(1);
}, null);

const el = $ref<HTMLElement | null>(null);
let tooltip: Tooltip | null = null;

onMounted(() => {
  if (el) tooltip = new Tooltip(el);
});

let showLegend = $ref(false);
const toggleLegend = () => {
  showLegend = !showLegend;
  if (tooltip) tooltip.hide();
};
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
