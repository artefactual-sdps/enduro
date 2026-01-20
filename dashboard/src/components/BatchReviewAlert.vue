<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useBatchStore } from "@/stores/batch";

const authStore = useAuthStore();
const batchStore = useBatchStore();
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
        class="btn btn-success"
        type="button"
        @click="batchStore.reviewBatch(true)"
      >
        Continue
      </button>
      <button
        class="btn btn-danger"
        type="button"
        @click="batchStore.reviewBatch(false)"
      >
        Cancel
      </button>
    </div>
  </div>
</template>
