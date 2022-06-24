<script setup lang="ts">
import { usePackageStore } from "@/stores/package";

const packageStore = usePackageStore();
</script>

<template>
  <div
    v-if="packageStore.isPending"
    class="alert alert-warning text-center"
    role="alert"
  >
    This package is mid-workflow, and is currently awaiting user decision(s).
  </div>
</template>
