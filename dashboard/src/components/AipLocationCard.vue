<script setup lang="ts">
import { ref, watch } from "vue";

import { storageServiceDownloadURL } from "@/client";
import UUID from "@/components/UUID.vue";
import { openLocationDialog } from "@/dialogs";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";

const authStore = useAuthStore();
const aipStore = useAipStore();
const failed = ref(false);

const choose = async () => {
  failed.value = false;
  const locationId = await openLocationDialog(aipStore.current?.locationId);
  if (!locationId) return;
  const error = await aipStore.move(locationId);
  if (error) {
    failed.value = true;
  }
};

const download = () => {
  if (!aipStore.current?.uuid) return;
  const url = storageServiceDownloadURL(aipStore.current?.uuid);
  window.open(url, "_blank");
};

watch(aipStore.ui.download, () => download());
</script>

<template>
  <div class="card mb-3">
    <div class="card-body">
      <div v-if="failed" class="alert alert-danger" role="alert">
        Move operation failed, try again!
      </div>
      <div v-if="aipStore.isMoving" class="alert alert-info" role="alert">
        The AIP is being moved into a new location.
      </div>
      <h4 class="card-title">Location</h4>
      <p class="card-text">
        <span v-if="aipStore.isRejected">AIP rejected.</span>
        <span v-else-if="!aipStore.current?.locationId"
          >Not available yet.</span
        >
        <span v-else><UUID :id="aipStore.current.locationId" /></span>
      </p>
      <div class="d-flex flex-wrap gap-2">
        <button
          v-if="
            aipStore.isStored &&
            authStore.checkAttributes(['storage:aips:download'])
          "
          type="button"
          class="btn btn-primary btn-sm"
          @click="download"
          :disabled="aipStore.isMoving"
        >
          Download
        </button>
        <button
          v-if="
            aipStore.isStored &&
            authStore.checkAttributes(['storage:aips:move'])
          "
          type="button"
          class="btn btn-primary btn-sm"
          @click="choose"
          :disabled="!aipStore.isMovable"
        >
          <template v-if="aipStore.isMoving">
            <span
              class="spinner-grow spinner-grow-sm me-2"
              role="status"
              aria-hidden="true"
            ></span>
            Moving...
          </template>
          <template v-else>Move</template>
        </button>
      </div>
    </div>
  </div>
</template>
