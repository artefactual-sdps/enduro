<script lang="ts">
type PackageEnum =
  | api.EnduroIngestSipStatusEnum
  | api.EnduroIngestBatchStatusEnum
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
import type { Component } from "vue";

import IconValidated from "~icons/clarity/checkbox-list-line";
import IconCanceled from "~icons/clarity/cursor-hand-open-line";
import IconError from "~icons/clarity/flame-line";
import IconQueued from "~icons/clarity/hourglass-line";
import IconFailed from "~icons/clarity/remove-line";
import IconIngested from "~icons/clarity/success-standard-line";
import IconProcessing from "~icons/clarity/sync-line";
import IconPending from "~icons/clarity/warning-standard-line";

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
    "text-success-emphasis",
    "bg-success-subtle",
    "border border-2 border-success",
  ],
  stored: [
    "text-success-emphasis",
    "bg-success-subtle",
    "border border-2 border-success",
  ],
  deleted: [
    "text-danger-emphasis",
    "bg-danger-subtle",
    "border border-2 border-danger",
  ],
  failed: [
    "text-danger-emphasis",
    "bg-danger-subtle",
    "border border-2 border-danger",
  ],
  error: [
    "text-danger-emphasis",
    "bg-danger-subtle",
    "border border-2 border-danger",
  ],
  queued: [
    "text-secondary-emphasis",
    "bg-secondary-subtle",
    "border border-2 border-secondary",
  ],
  processing: [
    "text-info-emphasis",
    "bg-info-subtle",
    "border border-2 border-info",
  ],
  pending: [
    "text-warning-emphasis",
    "bg-warning-subtle",
    "border border-2 border-warning",
  ],
  unspecified: [
    "text-body-emphasis",
    "bg-dark-subtle",
    "border border-2 border-dark",
  ],
  validated: [
    "text-blue-emphasis",
    "bg-blue-subtle",
    "border border-2 border-blue",
  ],
  canceled: [
    "status-canceled",
    "text-body-emphasis",
    "bg-dark-subtle",
    "border border-2",
  ],
};

const packageIcon: Record<PackageEnum, Component | undefined> = {
  ingested: IconIngested,
  stored: IconIngested,
  deleted: IconFailed,
  failed: IconFailed,
  error: IconError,
  queued: IconQueued,
  processing: IconProcessing,
  pending: IconPending,
  unspecified: undefined,
  validated: IconValidated,
  canceled: IconCanceled,
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

const icon = computed<Component | undefined>(() => {
  if (props.type !== "package") {
    return undefined;
  }

  return packageIcon[props.status as PackageEnum];
});
</script>

<template>
  <span>
    <span
      :class="[
        'badge',
        'd-inline-flex',
        'align-items-center',
        'gap-2',
        colorClass,
      ]"
    >
      <component :is="icon" v-if="icon" aria-hidden="true" />
      {{ props.status.toUpperCase() }}
      <div
        v-if="props.status == api.EnduroIngestSipWorkflowStatusEnum.InProgress"
        class="spinner-border status-badge-spinner text-black"
        role="progress"
        aria-hidden="true"
      />
    </span>
    <span v-if="props.note" class="badge text-body fw-normal"
      >({{ props.note }})</span
    >
  </span>
</template>

<style scoped>
.status-badge-spinner {
  --bs-spinner-width: 1em;
  --bs-spinner-height: 1em;
  --bs-spinner-border-width: 0.2em;
}

.status-canceled {
  --bs-border-color: currentColor;
}
</style>
