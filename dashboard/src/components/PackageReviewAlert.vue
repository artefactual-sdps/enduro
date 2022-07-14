<script setup lang="ts">
import { openPackageLocationDialog } from "@/dialogs";
import { usePackageStore } from "@/stores/package";

let { expandCounter } = defineProps<{
  expandCounter: number;
}>();

const emit = defineEmits<{
  (e: "update:expandCounter", value: number): void;
}>();

const packageStore = usePackageStore();

const confirm = async () => {
  const locationName = await openPackageLocationDialog();
  if (!locationName) return;
  packageStore.confirm(locationName);
};
</script>

<template>
  <div class="alert alert-info" role="alert" v-if="packageStore.isPending">
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
        <a href="#" @click.prevent="packageStore.ui.download.request"
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
        @click="packageStore.reject()"
      >
        Reject
      </button>
      <button class="btn btn-success" type="button" @click="confirm">
        Confirm
      </button>
    </div>
  </div>
</template>
