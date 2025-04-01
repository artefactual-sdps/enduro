<script setup lang="ts">
import { ref, watch } from "vue";
import { openDialog } from "vue3-promise-dialog";

import { storageServiceDownloadURL } from "@/client";
import AipDeletionRequestDialog from "@/components/AipDeletionRequestDialog.vue";
import LocationDialog from "@/components/LocationDialog.vue";
import UUID from "@/components/UUID.vue";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";

const authStore = useAuthStore();
const aipStore = useAipStore();
const failed = ref(false);

const choose = async () => {
  failed.value = false;
  const locationId = await openDialog(LocationDialog, {
    currentLocationId: aipStore.current?.locationId,
  });
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

const requestDeletion = async () => {
  if (!aipStore.current) return;
  const reason = await openDialog(AipDeletionRequestDialog);
  if (!reason) return;
  // TODO: Handle error.
  await aipStore.requestDeletion(reason);
};
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
        <span v-else-if="aipStore.isDeleted">AIP deleted.</span>
        <span v-else-if="!aipStore.current?.locationId"
          >Not available yet.</span
        >
        <span v-else><UUID :id="aipStore.current.locationId" /></span>
      </p>
      <div class="d-flex flex-wrap gap-2">
        <button
          v-if="authStore.checkAttributes(['storage:aips:download'])"
          type="button"
          class="btn btn-primary btn-sm"
          @click="download"
          :disabled="!aipStore.isStored && !aipStore.isPending"
        >
          Download
        </button>
        <button
          v-if="
            false && // TODO: Enable this also based on location type and available locations.
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
        <button
          v-if="authStore.checkAttributes(['storage:aips:deletion:request'])"
          type="button"
          class="btn btn-primary btn-sm"
          @click="requestDeletion"
          :disabled="!aipStore.isStored"
        >
          Delete
        </button>
      </div>
    </div>
  </div>
</template>
