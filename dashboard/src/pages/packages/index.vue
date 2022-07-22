<script setup lang="ts">
import PackageListLegend from "@/components/PackageListLegend.vue";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import UUID from "@/components/UUID.vue";
import { usePackageStore } from "@/stores/package";
import { useAsyncState } from "@vueuse/core";
import { useRouter } from "vue-router";
import IconInfoFill from "~icons/akar-icons/info-fill";

const router = useRouter();
const packageStore = usePackageStore();

const { execute, error } = useAsyncState(() => {
  return packageStore.fetchPackages();
}, null);

let showLegend = $ref(false);
const toggleLegend = () => (showLegend = !showLegend);
</script>

<template>
  <div class="container-xxl pt-3">
    <h2>Packages</h2>
    <PageLoadingAlert :execute="execute" :error="error" />
    <PackageListLegend v-model="showLegend" />
    <table class="table table-bordered table-hover">
      <thead>
        <tr>
          <th scope="col">ID</th>
          <th scope="col">Name</th>
          <th scope="col">UUID</th>
          <th scope="col">Started</th>
          <th scope="col">Location</th>
          <th scope="col">
            <span class="d-flex">
              Status
              <button
                class="btn btn-sm btn-link text-decoration-none ms-auto p-0 ps-1"
                type="button"
                @click="toggleLegend"
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
          <td>
            <UUID :id="pkg.aipId" />
          </td>
          <td>{{ $filters.formatDateTime(pkg.startedAt) }}</td>
          <td>{{ pkg.location }}</td>
          <td>
            <StatusBadge :status="pkg.status" />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
