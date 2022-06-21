<script setup lang="ts">
const props = defineProps({
  status: { type: String, required: true },
});
</script>

<template>
  <span class="badge text-bg-warning">{{ props.status.toUpperCase() }}</span>
</template>
