<script setup lang="ts">
import { computed } from "vue";
const props = defineProps({
  offset: {
    type: Number,
    default: 0,
  },
  limit: {
    type: Number,
    default: 20,
  },
  total: {
    type: Number,
    default: 0,
  },
});

const last = computed(() => {
  // Calculate the last result shown on the page.
  return Math.min(props.offset + props.limit, props.total);
});
</script>

<template>
  <template v-if="props.total === 0"> No results found </template>
  <template v-else-if="props.total === 1">
    Found {{ props.total }} result
  </template>
  <template v-else-if="props.total <= props.limit">
    Found {{ props.total }} results
  </template>
  <template v-else>
    Showing {{ props.offset + 1 }} - {{ last }} of {{ props.total }}
    results
  </template>
</template>
