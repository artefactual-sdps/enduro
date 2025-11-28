<script setup lang="ts">
import { ref } from "vue";

import { api, client } from "@/client";
import { logError } from "@/helpers/logs";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";

const aipStore = useAipStore();
const authStore = useAuthStore();
const reportKey = ref("");

const downloadDeletionReport = async (reportKey: string) => {
  if (reportKey === "") {
    return;
  }
  // TODO: download deletion report.
};

// Fetch the deletion report key if the user has permission.
if (
  authStore.checkAttributes(["storage:aips:deletion:list"]) &&
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
      reportKey.value = resp[0].reportKey;
    })
    .catch((e) => {
      logError(e, "Fetch deletion request");
    });
}
</script>

<template>
  <div v-if="reportKey" class="card mb-3">
    <div class="card-body">
      <h4 class="card-title">Reports</h4>
      <div class="d-flex flex-wrap gap-2">
        <button
          type="button"
          class="btn btn-primary btn-sm"
          @click="downloadDeletionReport(reportKey)"
        >
          Deletion report
        </button>
      </div>
    </div>
  </div>
</template>
