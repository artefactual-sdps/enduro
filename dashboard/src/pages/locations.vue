<script setup lang="ts">
import { useStateStore } from "@/stores/state";

const stateStore = useStateStore();
stateStore.updateBreadcrumb([{ text: "Locations" }]);
</script>

<template>
  <div class="container-xxl">
    <h2>Locations</h2>
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
        <tbody></tbody>
      </table>
    </div>
  </div>
</template>
