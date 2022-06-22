<script setup lang="ts">
import { usePackageStore } from "../stores/package";

const packageStore = usePackageStore();

const location = "aip-review";
</script>

<template>
  <div class="alert alert-info" role="alert" v-if="packageStore.isPending">
    <h4 class="alert-heading">Task: Review AIP</h4>
    Your AIP has been created and is currently stored in
    <strong>{{ location }}</strong
    >. Please review the output and decide if you would like to keep the AIP or
    reject it.
    <hr />
    <div class="d-flex flex-wrap gap-2">
      <button
        class="btn btn-danger"
        type="button"
        @click="packageStore.reject()"
      >
        Reject
      </button>
      <button
        class="btn btn-success"
        type="button"
        @click="packageStore.confirm()"
      >
        Confirm
      </button>
    </div>
  </div>
</template>
