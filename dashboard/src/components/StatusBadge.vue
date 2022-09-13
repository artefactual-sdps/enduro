<script setup lang="ts">
import type { api } from "@/client";
import { computed } from "vue";

const props = defineProps<{
  status:
    | api.EnduroStoredPackageStatusEnum
    | api.EnduroPackagePreservationActionStatusEnum
    | api.EnduroPackagePreservationTaskStatusEnum;
  note?: string;
}>();

const classes: {
  [key in
    | api.EnduroStoredPackageStatusEnum
    | api.EnduroPackagePreservationActionStatusEnum
    | api.EnduroPackagePreservationTaskStatusEnum]: string;
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
