<script setup lang="ts">
import { openPackageLocationDialog } from "@/dialogs";
import { usePackageStore } from "@/stores/package";

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
