<script setup lang="ts">
import { useSipStore } from "@/stores/sip";

const sipStore = useSipStore();
</script>

<template>
  <div
    v-if="sipStore.isPending"
    class="alert alert-warning text-center"
    role="alert"
  >
    This SIP is mid-workflow, and is currently awaiting user decision(s).
  </div>
</template>
