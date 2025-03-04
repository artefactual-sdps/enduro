<script setup lang="ts">
import UUID from "@/components/UUID.vue";
import { useStorageStore } from "@/stores/storage";

const storageStore = useStorageStore();
</script>

<template>
  <div v-if="storageStore.current">
    <div class="row">
      <div class="col-md-6">
        <dl>
          <dt>Location name</dt>
          <dd>{{ storageStore.current.name }}</dd>
          <dt>Source</dt>
          <dd>
            {{ $filters.getLocationSourceLabel(storageStore.current.source) }}
          </dd>
          <dt>UUID</dt>
          <dd>
            <UUID :id="storageStore.current.uuid" />
          </dd>
          <dt>Purpose</dt>
          <dd>
            {{ $filters.getLocationPurposeLabel(storageStore.current.purpose) }}
          </dd>
          <dt>Status</dt>
          <dd>TODO</dd>
          <dt>Capacity</dt>
          <dd>TODO</dd>
          <dt>AIPs</dt>
          <dd>TODO</dd>
          <template v-if="storageStore.current.description">
            <dt>Description</dt>
            <dd>{{ storageStore.current.description }}</dd>
          </template>
        </dl>
      </div>
    </div>
  </div>
</template>
