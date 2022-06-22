<script setup lang="ts">
import { usePackageStore } from "../../../stores/package";

const packageStore = usePackageStore();
</script>

<template>
  <div class="row mt-3">
    <div class="col">
      <h2>Not done yet!</h2>
      <pre>{{ packageStore.current }}</pre>
    </div>
  </div>
</template>
