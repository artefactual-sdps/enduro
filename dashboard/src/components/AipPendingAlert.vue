<script setup lang="ts">
import { useAipStore } from "@/stores/aip";

const aipStore = useAipStore();
</script>

<template>
  <div
    v-if="aipStore.isPending"
    class="alert alert-warning text-center"
    role="alert"
  >
    This AIP is mid-workflow, and is currently awaiting user decision(s).
  </div>
</template>
