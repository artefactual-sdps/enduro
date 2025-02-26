<script setup lang="ts">
import { IngestListSipsStatusEnum } from "@/openapi-generator";
import { usePackageStore } from "@/stores/package";

const packageStore = usePackageStore();
</script>

<template>
  <div class="d-flex flex-wrap gap-3 mb-3 p-3 border bg-light">
    <h3 class="mb-0 pe-3 border-end">Filters</h3>
    <div class="d-flex gap-2">
      <label for="filter-status" class="align-self-center">Status</label>
      <select
        id="filter-status"
        v-model="packageStore.filters.status"
        @change="packageStore.fetchPackages(1)"
        class="form-select"
      >
        <option value="">any</option>
        <option v-for="item of IngestListSipsStatusEnum" :value="item">
          {{ item }}
        </option>
      </select>
    </div>
  </div>
</template>
