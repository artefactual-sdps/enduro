<script setup lang="ts">
import UUID from "@/components/UUID.vue";
import { useAuthStore } from "@/stores/auth";
import { useLocationStore } from "@/stores/location";

const authStore = useAuthStore();
const locationStore = useLocationStore();
</script>

<template>
  <div v-if="locationStore.current">
    <div class="text-muted mb-3">
      Showing {{ locationStore.currentAips.length }} /
      {{ locationStore.currentAips.length }}
    </div>

    <div class="table-responsive mb-3">
      <table class="table table-bordered mb-0">
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">UUID</th>
            <th scope="col">Deposited</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="aip in locationStore.currentAips" :key="aip.uuid">
            <td>
              <RouterLink
                v-if="authStore.checkAttributes(['storage:aips:read'])"
                :to="{ name: '/storage/aips/[id]/', params: { id: aip.uuid } }"
              >
                {{ aip.name }}
              </RouterLink>
              <span v-else>{{ aip.name }}</span>
            </td>
            <td><UUID :id="aip.uuid" /></td>
            <td>{{ $filters.formatDateTime(aip.createdAt) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
