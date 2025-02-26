<script setup lang="ts">
import UUID from "@/components/UUID.vue";
import { useStorageStore } from "@/stores/storage";

const storageStore = useStorageStore();
</script>

<template>
  <div v-if="storageStore.current">
    <div class="text-muted mb-3">
      Showing {{ storageStore.current_packages.length }} /
      {{ storageStore.current_packages.length }}
    </div>

    <div class="table-responsive mb-3">
      <table class="table table-bordered mb-0">
        <thead>
          <tr>
            <th scope="col">#</th>
            <th scope="col">Name</th>
            <th scope="col">Size</th>
            <th scope="col">UUID</th>
            <th scope="col">Deposited</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(pkg, index) in storageStore.current_packages"
            :key="pkg.uuid"
          >
            <td>{{ index + 1 }}</td>
            <td>{{ pkg.name }}</td>
            <td></td>
            <td>
              <UUID :id="pkg.uuid" />
            </td>
            <td>{{ $filters.formatDateTime(pkg.createdAt) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
