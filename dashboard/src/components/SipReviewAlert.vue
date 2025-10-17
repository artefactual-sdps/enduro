<script setup lang="ts">
import { openDialog } from "vue3-promise-dialog";

import LocationDialog from "@/components/LocationDialog.vue";
import { useAipStore } from "@/stores/aip";
import { useSipStore } from "@/stores/sip";

let { expandCounter } = defineProps<{
  expandCounter: number;
}>();

const emit = defineEmits<{
  (e: "update:expandCounter", value: number): void;
}>();

const aipStore = useAipStore();
const sipStore = useSipStore();

if (sipStore.current?.aipUuid) {
  aipStore.fetchCurrent(sipStore.current.aipUuid);
}

const confirm = async () => {
  const locationId = await openDialog(LocationDialog);
  if (!locationId) return;
  sipStore.confirm(locationId);
};
</script>

<template>
  <div v-if="sipStore.isPending" class="alert alert-info" role="alert">
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
        <a href="#" @click.prevent="aipStore.download">Download</a>
        a local copy of the AIP for inspection
      </li>
    </ul>
    <hr />
    <div class="d-flex flex-wrap gap-2">
      <button class="btn btn-danger" type="button" @click="sipStore.reject()">
        Reject
      </button>
      <button class="btn btn-success" type="button" @click="confirm">
        Confirm
      </button>
    </div>
  </div>
</template>
