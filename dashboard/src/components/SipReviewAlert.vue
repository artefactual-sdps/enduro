<script setup lang="ts">
import { openLocationDialog } from "@/dialogs";
import { useIngestStore } from "@/stores/ingest";

let { expandCounter } = defineProps<{
  expandCounter: number;
}>();

const emit = defineEmits<{
  (e: "update:expandCounter", value: number): void;
}>();

const ingestStore = useIngestStore();

const confirm = async () => {
  const locationId = await openLocationDialog();
  if (!locationId) return;
  ingestStore.confirm(locationId);
};
</script>

<template>
  <div class="alert alert-info" role="alert" v-if="ingestStore.isPending">
    <h4 class="alert-heading">Task: Review AIP</h4>
    <p>
      Please review the output and decide if you would like to keep the AIP or
      reject it.
    </p>
    <p class="mb-1">Links:</p>
    <ul>
      <li>
        <a
          href="#"
          @click.prevent="emit('update:expandCounter', expandCounter + 1)"
          >Expand</a
        >
        the task details below
      </li>
      <li>View a summary of the preservation metadata created</li>
      <li>
        <a href="#" @click.prevent="ingestStore.ui.download.request"
          >Download</a
        >
        a local copy of the AIP for inspection
      </li>
    </ul>
    <hr />
    <div class="d-flex flex-wrap gap-2">
      <button
        class="btn btn-danger"
        type="button"
        @click="ingestStore.reject()"
      >
        Reject
      </button>
      <button class="btn btn-success" type="button" @click="confirm">
        Confirm
      </button>
    </div>
  </div>
</template>
