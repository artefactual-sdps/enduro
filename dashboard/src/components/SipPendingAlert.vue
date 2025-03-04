<script setup lang="ts">
import { useIngestStore } from "@/stores/ingest";

const ingestStore = useIngestStore();
</script>

<template>
  <div
    v-if="ingestStore.isPending"
    class="alert alert-warning text-center"
    role="alert"
  >
    This SIP is mid-workflow, and is currently awaiting user decision(s).
  </div>
</template>
