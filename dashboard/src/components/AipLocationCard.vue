<script setup lang="ts">
import AipDeletionRequestDialog from "@/components/AipDeletionRequestDialog.vue";
import UUID from "@/components/UUID.vue";
import { openDialog } from "@/dialogs/dialog";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";

const authStore = useAuthStore();
const aipStore = useAipStore();

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
      <h4 class="card-title text-primary">Location</h4>
      <p class="card-text">
        <span v-if="aipStore.isDeleted">AIP deleted.</span>
        <span v-else-if="!aipStore.current?.locationUuid"
          >Not available yet.</span
        >
        <span v-else><UUID :id="aipStore.current.locationUuid" /></span>
      </p>
      <div v-if="!aipStore.isDeleted">
        <Transition mode="out-in">
          <div
            v-if="aipStore.downloadError"
            class="alert alert-danger text-center mb-0"
            role="alert"
          >
            {{ aipStore.downloadError }}
          </div>
          <div v-else class="d-flex flex-wrap gap-2">
            <button
              v-if="
                authStore.checkAttributes(['storage:aips:download']) &&
                (aipStore.isStored || aipStore.isPending)
              "
              type="button"
              class="btn btn-primary btn-sm"
              :disabled="aipStore.downloading"
              @click="aipStore.download()"
            >
              <template v-if="aipStore.downloading">
                <span
                  class="spinner-grow spinner-grow-sm me-2"
                  role="status"
                  aria-hidden="true"
                />
                Downloading...
              </template>
              <template v-else>Download</template>
            </button>
            <button
              v-if="
                authStore.checkAttributes(['storage:aips:deletion:request']) &&
                aipStore.isStored
              "
              type="button"
              class="btn btn-primary btn-sm"
              @click="requestDeletion"
            >
              Delete
            </button>
          </div>
        </Transition>
      </div>
    </div>
  </div>
</template>

<style scoped>
.v-enter-active,
.v-leave-active {
  transition: opacity 0.3s;
}
.v-enter-from,
.v-leave-to {
  opacity: 0;
}
.v-enter-to,
.v-leave-from {
  opacity: 1;
}
</style>
