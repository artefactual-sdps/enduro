<script setup lang="ts">
import { computed } from "vue";

import type { api } from "@/client";

const props = defineProps<{
  status:
    | api.EnduroIngestSipStatusEnum
    | api.EnduroIngestSipWorkflowStatusEnum
    | api.EnduroIngestSipTaskStatusEnum
    | api.EnduroStorageAipStatusEnum
    | api.EnduroStorageAipWorkflowStatusEnum
    | api.EnduroStorageAipTaskStatusEnum;
  note?: string;
}>();

const classes: {
  [key in
    | api.EnduroIngestSipStatusEnum
    | api.EnduroIngestSipWorkflowStatusEnum
    | api.EnduroIngestSipTaskStatusEnum
    | api.EnduroStorageAipStatusEnum
    | api.EnduroStorageAipWorkflowStatusEnum
    | api.EnduroStorageAipTaskStatusEnum]: string;
} = {
  "in progress": "text-bg-info",
  done: "text-bg-success",
  error: "text-bg-danger",
  queued: "text-bg-secondary",
  pending: "text-bg-warning",
  canceled: "text-bg-dark",
  unspecified: "text-bg-dark",
  stored: "text-bg-success",
  deleted: "text-bg-danger",
  processing: "text-bg-info",
  failed: "text-bg-danger",
  ingested: "text-bg-success",
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
