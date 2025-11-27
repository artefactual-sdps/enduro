<script setup lang="ts">
import { ref } from "vue";

import { client, getPath } from "@/client";
import { ResponseError } from "@/openapi-generator";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";

const aipStore = useAipStore();
const authStore = useAuthStore();
const error = ref("");

const downloadDeletionReport = async () => {
  if (!aipStore.current?.uuid) {
    return;
  }

  try {
    await client.storage.storageAipDeletionReportRequest({
      uuid: aipStore.current.uuid,
    });
    window.open(
      getPath() + `/storage/aips/${aipStore.current.uuid}/deletion-report`,
      "_blank",
    );
  } catch (err) {
    // Try to parse the error and save it for 5 seconds. It will
    // display an alert including the error message.
    let errorMsg = "Unexpected error downloading deletion report";
    if (err instanceof ResponseError) {
      const body = await err.response.json();
      if (body.message) {
        errorMsg = body.message;
      }
    }
    error.value = errorMsg;
    setTimeout(() => (error.value = ""), 5000);
  }
};
</script>

<template>
  <div
    v-if="
      aipStore.current?.deletionReportKey &&
      authStore.checkAttributes(['storage:aips:deletion:report'])
    "
    class="card mb-3"
  >
    <div class="card-body">
      <h4 class="card-title">Reports</h4>
      <div v-if="error" class="alert alert-warning" role="alert">
        {{ error }}
      </div>
      <div class="d-flex flex-wrap gap-2">
        <button
          type="button"
          class="btn btn-primary btn-sm"
          @click="downloadDeletionReport"
        >
          Deletion report
        </button>
      </div>
    </div>
  </div>
</template>
