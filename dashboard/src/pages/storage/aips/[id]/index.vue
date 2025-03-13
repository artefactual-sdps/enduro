<script setup lang="ts">
import StatusBadge from "@/components/StatusBadge.vue";
import UUID from "@/components/UUID.vue";
import { useAipStore } from "@/stores/aip";

const aipStore = useAipStore();
</script>

<template>
  <div v-if="aipStore.current">
    <div class="row">
      <div class="col-md-6">
        <h2>AIP details</h2>
        <dl>
          <dt>Name</dt>
          <dd>{{ aipStore.current.name }}</dd>
          <dt>UUID</dt>
          <dd><UUID :id="aipStore.current.uuid" /></dd>
          <dt>Status</dt>
          <dd><StatusBadge :status="aipStore.current.status" /></dd>
          <dt>Deposited</dt>
          <dd>{{ $filters.formatDateTime(aipStore.current.createdAt) }}</dd>
        </dl>
      </div>
    </div>
  </div>
</template>
