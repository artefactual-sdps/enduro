<script lang="ts">
type PackageEnum =
  | api.EnduroIngestSipStatusEnum
  | api.EnduroStorageAipStatusEnum;
type WorkflowEnum =
  | api.EnduroIngestSipWorkflowStatusEnum
  | api.EnduroStorageAipWorkflowStatusEnum;
type TaskEnum =
  | api.EnduroIngestSipTaskStatusEnum
  | api.EnduroStorageAipTaskStatusEnum;

export type StatusEnum = PackageEnum | WorkflowEnum | TaskEnum;
</script>

<script setup lang="ts">
import { computed } from "vue";

import { api } from "@/client";

type BadgeType = "package" | "workflow";
type BadgeStyle = string[];

const props = defineProps<{
  status: StatusEnum;
  type: BadgeType;
  note?: string;
}>();

const packageStyle: {
  [key in PackageEnum]: BadgeStyle;
} = {
  ingested: [
    "text-dark",
    "bg-success-subtle",
    "border border-2 border-success",
  ],
  stored: ["text-dark", "bg-success-subtle", "border border-2 border-success"],
  deleted: ["text-dark", "bg-danger-subtle", "border border-2 border-danger"],
  failed: ["text-dark", "bg-danger-subtle", "border border-2 border-danger"],
  error: ["text-dark", "bg-danger-subtle", "border border-2 border-danger"],
  queued: [
    "text-dark",
    "bg-secondary-subtle",
    "border border-2 border-secondary",
  ],
  processing: ["text-dark", "bg-info-subtle", "border border-2 border-info"],
  pending: ["text-dark", "bg-warning-subtle", "border border-2 border-warning"],
  unspecified: ["text-dark", "bg-dark-subtle", "border border-2 border-dark"],
};

const workflowStyle: {
  [key in WorkflowEnum | TaskEnum]: BadgeStyle;
} = {
  done: ["text-bg-success"],
  failed: ["text-bg-danger"],
  error: ["text-bg-danger"],
  queued: ["text-bg-secondary"],
  "in progress": ["text-bg-info"],
  pending: ["text-bg-warning"],
  canceled: ["text-bg-dark"],
  unspecified: ["text-bg-dark"],
};

const colorClass = computed(() => {
  function getBadgeStyle(type: BadgeType): BadgeStyle {
    switch (type) {
      case "package":
        return packageStyle[props.status as PackageEnum];
      case "workflow":
        return workflowStyle[props.status as WorkflowEnum];
    }
  }

  return getBadgeStyle(props.type).join(" ");
});
</script>

<template>
  <span>
    <span :class="['badge', colorClass]">
      {{ props.status.toUpperCase() }}
      <div
        v-if="props.status == api.EnduroIngestSipWorkflowStatusEnum.InProgress"
        class="spinner-border spinner-border-sm text-black"
        role="progress"
        aria-hidden="true"
      />
    </span>
    <span v-if="props.note" class="badge text-dark fw-normal"
      >({{ props.note }})</span
    >
  </span>
</template>
