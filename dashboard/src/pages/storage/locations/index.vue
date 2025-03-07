<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";

import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import UUID from "@/components/UUID.vue";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { useStorageStore } from "@/stores/storage";
import IconLocations from "~icons/octicon/server-24";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
layoutStore.updateBreadcrumb([{ text: "Storage" }, { text: "Locations" }]);

const storageStore = useStorageStore();
const { execute, error } = useAsyncState(() => {
  return storageStore.fetchLocations();
}, null);
</script>

<template>
  <div class="container-xxl">
    <h1 class="d-flex mb-0">
      <IconLocations class="me-3 text-dark" />Locations
    </h1>
    <div class="text-muted mb-3">
      Showing {{ storageStore.locations.length }} /
      {{ storageStore.locations.length }}
    </div>
    <PageLoadingAlert :execute="execute" :error="error" />
    <div class="table-responsive mb-3">
      <table class="table table-bordered mb-0">
        <thead>
          <tr>
            <th scope="col" class="text-nowrap">Name</th>
            <th scope="col">Source</th>
            <th scope="col">Purpose</th>
            <th scope="col">Capacity</th>
            <th scope="col">AIPs</th>
            <th scope="col">UUID</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in storageStore.locations" :key="item.uuid">
            <td>
              <router-link
                v-if="authStore.checkAttributes(['storage:locations:read'])"
                :to="{
                  name: '/storage/locations/[id]/',
                  params: { id: item.uuid },
                }"
                >{{ item.name }}</router-link
              >
              <span v-else>{{ item.name }}</span>
            </td>
            <td>{{ $filters.getLocationSourceLabel(item.source) }}</td>
            <td>{{ $filters.getLocationPurposeLabel(item.purpose) }}</td>
            <td></td>
            <td></td>
            <td>
              <UUID :id="item.uuid" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
