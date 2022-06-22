<script setup lang="ts">
import { usePackageStore } from "../stores/package";

const packageStore = usePackageStore();
</script>

<template>
  <div class="row" v-if="packageStore.isPending">
    <div class="col-12">
      <div class="alert alert-warning text-center mb-0 m-3" role="alert">
        This package is mid-workflow, and is currently awaiting user
        decision(s).
      </div>
    </div>
  </div>
</template>
