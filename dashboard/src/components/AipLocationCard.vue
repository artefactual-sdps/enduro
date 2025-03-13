<script setup lang="ts">
import { ref, watch } from "vue";

import { storageServiceDownloadURL } from "@/client";
import UUID from "@/components/UUID.vue";
import { openLocationDialog } from "@/dialogs";
import { useAuthStore } from "@/stores/auth";
import { useSipStore } from "@/stores/sip";

const authStore = useAuthStore();
const sipStore = useSipStore();

let failed = ref<boolean>(false);

const choose = async () => {
  failed.value = false;
  const locationId = await openLocationDialog(sipStore.current?.locationId);
  if (!locationId) return;
  const error = await sipStore.move(locationId);
  if (error) {
    failed.value = true;
  }
};

const download = () => {
  if (!sipStore.current?.aipId) return;
  const url = storageServiceDownloadURL(sipStore.current.aipId);
  window.open(url, "_blank");
};

watch(sipStore.ui.download, () => download());
</script>

<template>
  <div class="card mb-3">
    <div class="card-body">
      <div v-if="failed" class="alert alert-danger" role="alert">
        Move operation failed, try again!
      </div>
      <div v-if="sipStore.isMoving" class="alert alert-info" role="alert">
        The AIP is being moved into a new location.
      </div>
      <h4 class="card-title">Location</h4>
      <p class="card-text">
        <span v-if="sipStore.isRejected">AIP rejected.</span>
        <span v-else-if="!sipStore.current?.locationId"
          >Not available yet.</span
        >
        <span v-else><UUID :id="sipStore.current.locationId" /></span>
      </p>
      <div class="d-flex flex-wrap gap-2">
        <button
          v-if="
            !sipStore.isRejected &&
            authStore.checkAttributes(['ingest:sips:move'])
          "
          type="button"
          class="btn btn-primary btn-sm"
          @click="choose"
          :disabled="!sipStore.isMovable"
        >
          <template v-if="sipStore.isMoving">
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
            sipStore.current?.aipId?.length &&
            authStore.checkAttributes(['storage:aips:download'])
          "
          type="button"
          class="btn btn-primary btn-sm"
          @click="download"
          :disabled="sipStore.isMoving"
        >
          Download
        </button>
      </div>
    </div>
  </div>
</template>
