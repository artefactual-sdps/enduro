<script setup lang="ts">
import { ref } from "vue";

import UUID from "@/components/UUID.vue";
import { openPackageLocationDialog } from "@/dialogs";
import { useAuthStore } from "@/stores/auth";
import { useIngestStore } from "@/stores/ingest";

const authStore = useAuthStore();
const ingestStore = useIngestStore();

let failed = ref<boolean | null>(null);

const choose = async () => {
  failed.value = false;
  const locationId = await openPackageLocationDialog(
    ingestStore.currentSip?.locationId,
  );
  if (!locationId) return;
  const error = await ingestStore.move(locationId);
  if (error) {
    failed.value = true;
  }
};
</script>

<template>
  <div class="card mb-3">
    <div class="card-body">
      <div v-if="failed" class="alert alert-danger" role="alert">
        Move operation failed, try again!
      </div>
      <div v-if="ingestStore.isMoving" class="alert alert-info" role="alert">
        The package is being moved into a new location.
      </div>
      <h4 class="card-title">Location</h4>
      <p class="card-text">
        <span v-if="ingestStore.isRejected">Package rejected.</span>
        <span v-else-if="!ingestStore.currentSip?.locationId"
          >Not available yet.</span
        >
        <span v-else><UUID :id="ingestStore.currentSip.locationId" /></span>
      </p>
      <div
        class="actions"
        v-if="
          !ingestStore.isRejected &&
          authStore.checkAttributes(['ingest:sips:move'])
        "
      >
        <button
          type="button"
          class="btn btn-primary btn-sm"
          @click="choose"
          :disabled="!ingestStore.isMovable"
        >
          <template v-if="ingestStore.isMoving">
            <span
              class="spinner-grow spinner-grow-sm me-2"
              role="status"
              aria-hidden="true"
            ></span>
            Moving...
          </template>
          <template v-else>Choose storage location</template>
        </button>
      </div>
    </div>
  </div>
</template>
