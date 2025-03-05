<script setup lang="ts">
import { ref, watch } from "vue";

import { storageServiceDownloadURL } from "@/client";
import UUID from "@/components/UUID.vue";
import { openLocationDialog } from "@/dialogs";
import { useAuthStore } from "@/stores/auth";
import { useIngestStore } from "@/stores/ingest";

const authStore = useAuthStore();
const ingestStore = useIngestStore();

let failed = ref<boolean | null>(null);

const choose = async () => {
  failed.value = false;
  const locationId = await openLocationDialog(
    ingestStore.currentSip?.locationId,
  );
  if (!locationId) return;
  const error = await ingestStore.move(locationId);
  if (error) {
    failed.value = true;
  }
};

const download = () => {
  if (!ingestStore.currentSip?.aipId) return;
  const url = storageServiceDownloadURL(ingestStore.currentSip.aipId);
  window.open(url, "_blank");
};

watch(ingestStore.ui.download, () => download());
</script>

<template>
  <div class="card mb-3">
    <div class="card-body">
      <div v-if="failed" class="alert alert-danger" role="alert">
        Move operation failed, try again!
      </div>
      <div v-if="ingestStore.isMoving" class="alert alert-info" role="alert">
        The AIP is being moved into a new location.
      </div>
      <h4 class="card-title">Location</h4>
      <p class="card-text">
        <span v-if="ingestStore.isRejected">AIP rejected.</span>
        <span v-else-if="!ingestStore.currentSip?.locationId"
          >Not available yet.</span
        >
        <span v-else><UUID :id="ingestStore.currentSip.locationId" /></span>
      </p>
      <div class="d-flex flex-wrap gap-2">
        <button
          v-if="
            !ingestStore.isRejected &&
            authStore.checkAttributes(['ingest:sips:move'])
          "
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
        <button
          v-if="
            ingestStore.currentSip?.aipId?.length &&
            authStore.checkAttributes(['storage:aips:download'])
          "
          type="button"
          class="btn btn-primary btn-sm"
          @click="download"
          :disabled="ingestStore.isMoving"
        >
          Download
        </button>
      </div>
    </div>
  </div>
</template>
