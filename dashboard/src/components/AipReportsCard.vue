<script setup lang="ts">
import { ref } from "vue";

import { api, client, getPath } from "@/client";
import { logError } from "@/helpers/logs";
import { ResponseError } from "@/openapi-generator";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";

const aipStore = useAipStore();
const authStore = useAuthStore();
const deletionRequestID = ref("");
const error = ref("");

const downloadDeletionReport = async (uuid: string) => {
  if (uuid === "") {
    return;
  }

  try {
    await client.storage.storageDownloadDeletionReportRequest({
      uuid: uuid,
    });
    window.open(
      getPath() + `/storage/deletion-reports/${uuid}/download`,
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

// Fetch the deletion report key if the user has permission.
if (
  authStore.checkAttributes(["storage:deletion_report:download"]) &&
  aipStore.current &&
  aipStore.isDeleted
) {
  client.storage
    .storageListDeletionRequests({
      uuid: aipStore.current.uuid,
      status: api.StorageListDeletionRequestsStatusEnum.Approved,
    })
    .then((resp) => {
      if (resp.length === 0 || !resp[0].reportKey) {
        return;
      }
      deletionRequestID.value = resp[0].uuid;
    })
    .catch((e) => {
      logError(e, "Fetch deletion request");
    });
}
</script>

<template>
  <div v-if="deletionRequestID" class="card mb-3">
    <div class="card-body">
      <h4 class="card-title">Reports</h4>
      <div v-if="error" class="alert alert-warning" role="alert">
        {{ error }}
      </div>
      <div class="d-flex flex-wrap gap-2">
        <button
          type="button"
          class="btn btn-primary btn-sm"
          @click="downloadDeletionReport(deletionRequestID)"
        >
          Deletion report
        </button>
      </div>
    </div>
  </div>
</template>
