<script setup lang="ts">
import { computed } from "vue";

import type { api } from "@/client";

const props = defineProps<{
  status:
    | api.EnduroIngestSipStatusEnum
    | api.EnduroIngestSipPreservationActionStatusEnum
    | api.EnduroIngestSipPreservationTaskStatusEnum;
  note?: string;
}>();

const classes: {
  [key in
    | api.EnduroIngestSipStatusEnum
    | api.EnduroIngestSipPreservationActionStatusEnum
    | api.EnduroIngestSipPreservationTaskStatusEnum]: string;
} = {
  new: "text-bg-dark",
  "in progress": "text-bg-secondary",
  done: "text-bg-success",
  error: "text-bg-danger",
  unknown: "text-bg-dark",
  queued: "text-bg-info",
  pending: "text-bg-warning",
  abandoned: "text-bg-dark",
  unspecified: "text-bg-dark",
};

const colorClass = computed(() => {
  return classes[props.status];
});
</script>

<template>
  <span>
    <span :class="['badge', colorClass]">
      {{ props.status.toUpperCase() }}
    </span>
    <span v-if="props.note" class="badge text-dark fw-normal"
      >({{ props.note }})</span
    >
  </span>
</template>
