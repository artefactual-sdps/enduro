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

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
layoutStore.updateBreadcrumb([{ text: "Packages" }]);

const packageStore = usePackageStore();

const { execute, error } = useAsyncState(() => {
  return packageStore.fetchPackages();
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
      Showing {{ packageStore.packages.length }} /
      {{ packageStore.packages.length }}
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
    <div>
      <nav aria-label="Package list pages">
        <ul class="pagination justify-content-center">
          <li class="page-item">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled: !packageStore.hasPrevPage,
              }"
              @click.prevent="packageStore.prevPage"
              >Previous</a
            >
          </li>
          <li class="page-item">
            <a
              href="#"
              :class="{
                'page-link': true,
                disabled: !packageStore.hasNextPage,
              }"
              @click.prevent="packageStore.nextPage"
              >Next</a
            >
          </li>
        </ul>
      </nav>
    </div>
  </div>
</template>
