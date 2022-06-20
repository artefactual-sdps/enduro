<script setup lang="ts">
import { usePackageStore } from "../stores/package";

const store = usePackageStore();
const location = "aip-review";
</script>

<template>
  <div class="alert alert-info" role="alert">
    <h4 class="alert-heading">Task: Review AIP</h4>
    Your AIP has been created and is currently stored in
    <strong>{{ location }}</strong
    >. Please review the output and decide if you would like to keep the AIP or
    reject it.
    <hr />
    <div class="d-flex flex-wrap gap-2">
      <button class="btn btn-danger" type="button" @click="store.reject()">
        Reject
      </button>
      <button class="btn btn-success" type="button" @click="store.confirm()">
        Confirm
      </button>
    </div>
  </div>
</template>
