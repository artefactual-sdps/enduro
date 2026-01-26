<script setup lang="ts">
import { openDialog } from "vue3-promise-dialog";

import BatchReviewConfirmDialog from "@/components/BatchReviewConfirmDialog.vue";
import { useAuthStore } from "@/stores/auth";
import { useBatchStore } from "@/stores/batch";
import IconContinue from "~icons/clarity/thumbs-up-line";
import IconCancel from "~icons/clarity/trash-line";

const authStore = useAuthStore();
const batchStore = useBatchStore();

const confirmCancel = async () => {
  const confirmed = await openDialog(BatchReviewConfirmDialog, {
    heading: "Cancel batch",
    bodyHtml:
      `<p>Are you sure you want to cancel batch <strong>${batchStore.current?.identifier}</strong>?</p>` +
      "<p>Clicking yes will mark this batch as CANCELED. Any SIPs that have already been ingested will remain in AIP storage and can be deleted manually.</p>",
    confirmClass: "btn-danger",
  });
  if (!confirmed) return;
  await batchStore.reviewBatch(false);
};

const confirmContinue = async () => {
  const confirmed = await openDialog(BatchReviewConfirmDialog, {
    heading: "Continue partial ingest",
    bodyHtml:
      `<p>Are you sure you want to continue processing batch <strong>${batchStore.current?.identifier}</strong>?</p>` +
      "<p>Clicking yes will mark this batch as INGESTED, even though some SIPs failed the ingest process.</p>",
    confirmClass: "btn-primary",
  });
  if (!confirmed) return;
  await batchStore.reviewBatch(true);
};
</script>

<template>
  <div v-if="batchStore.isPending" class="alert alert-info" role="alert">
    <h4 class="alert-heading">Review batch</h4>
    <p>
      Some SIPs in this batch were not fully ingested - you can click through to
      the SIP details page of any packages listed below to see more information.
    </p>
    <p>
      Choose whether to continue processing the batch without the uningested
      SIPs or to cancel the batch.
    </p>
    <div
      v-if="authStore.checkAttributes(['ingest:batches:review'])"
      class="d-flex flex-wrap gap-2"
    >
      <button
        class="btn btn-primary d-flex align-items-center gap-2"
        type="button"
        @click="confirmContinue"
      >
        <IconContinue aria-hidden="true" />
        Continue partial ingest
      </button>
      <button
        class="btn btn-danger d-flex align-items-center gap-2"
        type="button"
        @click="confirmCancel"
      >
        <IconCancel aria-hidden="true" />
        Cancel batch
      </button>
    </div>
  </div>
</template>
