<script setup lang="ts">
import { openPackageLocationDialog } from "@/dialogs";
import { usePackageStore } from "@/stores/package";
import { useRoute } from "vue-router";

const route = useRoute();
const packageStore = usePackageStore();

const confirm = async () => {
  const locationId = await openPackageLocationDialog();
  if (!locationId) return;
  packageStore.confirm(locationId);
};
</script>

<template>
  <div class="alert alert-warning" role="alert" v-if="packageStore.isPending">
    <h4 class="alert-heading">Task: Review AIP</h4>
    <p>
      This package is mid-workflow. Please review the output and decide if you
      would like to keep the AIP or reject it.
    </p>
    <ul>
      <li>
        <router-link
          :to="{
            name: 'packages-id-workflows',
            params: { id: route.params.id },
          }"
          >Check</router-link
        >
        the list of actions associated with the workflow
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
