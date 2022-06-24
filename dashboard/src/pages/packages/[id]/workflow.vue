<script setup lang="ts">
import { usePackageStore } from "@/stores/package";

const packageStore = usePackageStore();
</script>

<template>
  <div v-if="packageStore.current">
    <h2>Not done yet!</h2>
    <pre>{{ packageStore.current }}</pre>
  </div>
</template>
