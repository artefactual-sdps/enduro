<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { useStorageStore } from "@/stores/storage";
import IconRackServerLine from "~icons/clarity/rack-server-line";

const layoutStore = useLayoutStore();
layoutStore.updateBreadcrumb([{ text: "Locations" }]);

const storageStore = useStorageStore();
storageStore.fetchLocations();
</script>

<template>
  <div class="container-xxl">
    <h1 class="d-flex mb-3">
      <IconRackServerLine class="me-3 text-dark" />Locations
    </h1>
    <div class="table-responsive mb-3">
      <table class="table table-bordered mb-0">
        <thead>
          <tr>
            <th scope="col" class="text-nowrap">Location name</th>
            <th scope="col">Source</th>
            <th scope="col">Purpose</th>
            <th scope="col">Capacity</th>
            <th scope="col">Packages</th>
            <th scope="col">UUID</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, index) in storageStore.locations">
            <td>{{ item.name }}</td>
            <td>{{ $filters.getLocationSourceLabel(item.source) }}</td>
            <td>{{ $filters.getLocationPurposeLabel(item.purpose) }}</td>
            <td></td>
            <td></td>
            <td>{{ item.uuid }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
