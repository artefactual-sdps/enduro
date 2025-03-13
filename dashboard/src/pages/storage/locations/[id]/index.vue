<script setup lang="ts">
import UUID from "@/components/UUID.vue";
import { useLocationStore } from "@/stores/location";

const locationStore = useLocationStore();
</script>

<template>
  <div v-if="locationStore.current">
    <div class="row">
      <div class="col-md-6">
        <h2>Location details</h2>
        <dl>
          <dt>Name</dt>
          <dd>{{ locationStore.current.name }}</dd>
          <dt>UUID</dt>
          <dd><UUID :id="locationStore.current.uuid" /></dd>
          <dt>Source</dt>
          <dd>
            {{ $filters.getLocationSourceLabel(locationStore.current.source) }}
          </dd>
          <dt>Purpose</dt>
          <dd>
            {{
              $filters.getLocationPurposeLabel(locationStore.current.purpose)
            }}
          </dd>
          <template v-if="locationStore.current.description">
            <dt>Description</dt>
            <dd>{{ locationStore.current.description }}</dd>
          </template>
        </dl>
      </div>
    </div>
  </div>
</template>
