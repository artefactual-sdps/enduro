<script setup lang="ts">
import PackageListLegend from "@/components/PackageListLegend.vue";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import Tabs from "@/components/Tabs.vue";
import UUID from "@/components/UUID.vue";
import { useLayoutStore } from "@/stores/layout";
import { usePackageStore } from "@/stores/package";
import { useAsyncState } from "@vueuse/core";
import Tooltip from "bootstrap/js/dist/tooltip";
import { onMounted } from "vue";
import IconInfoFill from "~icons/akar-icons/info-fill";
import IconBundleLine from "~icons/clarity/bundle-line";
import RawIconBundleLine from "~icons/clarity/bundle-line?raw&font-size=20px";
import RawIconClockLine from "~icons/clarity/clock-line?raw&font-size=20px";
import RawIconContainerVolumeLine from "~icons/clarity/container-volume-line?raw&font-size=20px";
import RawIconRemoveLine from "~icons/clarity/remove-line?raw&font-size=20px";
import RawIconThumbsDownLine from "~icons/clarity/thumbs-down-line?raw&font-size=20px";
import RawIconWarningLine from "~icons/clarity/warning-line?raw&font-size=22px";

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

const tabs = [
  { icon: RawIconBundleLine, text: "All", route: { name: "packages" } },
  {
    icon: RawIconWarningLine,
    text: "Pending",
    route: { name: "packages" },
  },
  {
    icon: RawIconContainerVolumeLine,
    text: "Stored",
    route: { name: "packages" },
  },
  {
    icon: RawIconClockLine,
    text: "In progress",
    route: { name: "packages" },
  },
  {
    icon: RawIconThumbsDownLine,
    text: "Rejected",
    route: { name: "packages" },
  },
  {
    icon: RawIconRemoveLine,
    text: "Error",
    route: { name: "packages" },
  },
];
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

    <Tabs :tabs="tabs" />

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
                :to="{ name: 'packages-id', params: { id: pkg.id } }"
                >{{ pkg.name }}</router-link
              >
            </td>
            <td><UUID :id="pkg.aipId" /></td>
            <td>{{ $filters.formatDateTime(pkg.startedAt) }}</td>
            <td><UUID :id="pkg.locationId" /></td>
            <td><StatusBadge :status="pkg.status" /></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
