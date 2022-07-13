<script setup lang="ts">
import PackageListLegend from "@/components/PackageListLegend.vue";
import PackageStatusBadge from "@/components/PackageStatusBadge.vue";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import { usePackageStore } from "@/stores/package";
import { useAsyncState } from "@vueuse/core";
import { useRouter } from "vue-router";
import IconInfoFill from "~icons/akar-icons/info-fill";

const router = useRouter();
const packageStore = usePackageStore();

const { execute, error } = useAsyncState(() => {
  return packageStore.fetchPackages();
}, null);

const openPackage = (id: number) => {
  router.push({ name: "packages-id", params: { id } });
};

let showLegend = $ref(false);
const toggleLegend = () => (showLegend = !showLegend);
</script>

<template>
  <div class="container-xxl pt-3">
    <h2>Packages</h2>
    <PageLoadingAlert :execute="execute" :error="error" />
    <PackageListLegend v-model="showLegend" />
    <table class="table table-bordered table-hover table-linked table-enduro">
      <thead>
        <tr>
          <th scope="col">ID</th>
          <th scope="col">Name</th>
          <th scope="col">UUID</th>
          <th scope="col">Started</th>
          <th scope="col">Location</th>
          <th scope="col" class="text-nowrap">
            Status
            <a href="#" @click.prevent="toggleLegend"><IconInfoFill /></a>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="pkg in packageStore.packages"
          :key="pkg.id"
          @click="openPackage(pkg.id)"
        >
          <td scope="row">{{ pkg.id }}</td>
          <td>
            <router-link
              :to="{ name: 'packages-id', params: { id: pkg.id } }"
              >{{ pkg.name }}</router-link
            >
          </td>
          <td class="font-monospace">{{ pkg.aipId }}</td>
          <td>{{ $filters.formatDateTime(pkg.startedAt) }}</td>
          <td>{{ pkg.location }}</td>
          <td>
            <PackageStatusBadge :status="pkg.status" />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
