<script setup lang="ts">
import { computed, ref, toRefs } from "vue";

import { api } from "@/client";
import AipDeletionReviewAlert from "@/components/AipDeletionReviewAlert.vue";
import SipReviewAlert from "@/components/SipReviewAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import Task from "@/components/Task.vue";
import { useAuthStore } from "@/stores/auth";

const authStore = useAuthStore();
const tasks = computed<api.EnduroIngestSipTask[] | api.EnduroStorageAipTask[]>(
  () => {
    if (!props.workflow.tasks) {
      return [];
    }

    // Show the last task first.
    return props.workflow.tasks.slice().reverse();
  },
);

const props = defineProps<{
  workflow: api.EnduroIngestSipWorkflow | api.EnduroStorageAipWorkflow;
  index: number;
}>();

const { workflow, index } = toRefs(props);

let expandCounter = ref<number>(0);

const showSipReviewAlert = (
  workflow: api.EnduroIngestSipWorkflow | api.EnduroStorageAipWorkflow,
) => {
  return (
    workflow.type == api.EnduroIngestSipWorkflowTypeEnum.CreateAndReviewAip &&
    workflow.status == api.EnduroIngestSipWorkflowStatusEnum.Pending
  );
};

const showAipDeletionReviewAlert = (
  workflow: api.EnduroIngestSipWorkflow | api.EnduroStorageAipWorkflow,
) => {
  return (
    workflow.type == api.EnduroStorageAipWorkflowTypeEnum.DeleteAip &&
    workflow.status == api.EnduroStorageAipWorkflowStatusEnum.Pending
  );
};
</script>

<template>
  <div class="accordion-item border-0 mb-2">
    <h4 class="accordion-header" :id="'w-heading-' + index">
      <button
        v-if="workflow.tasks"
        class="accordion-button collapsed"
        type="button"
        data-bs-toggle="collapse"
        :data-bs-target="'#w-body-' + index"
        aria-expanded="false"
        :aria-controls="'w-body-' + index"
      >
        <div class="d-flex flex-column">
          <div class="h4">
            {{ $filters.getWorkflowLabel(workflow.type) }}
            <StatusBadge :status="workflow.status" type="workflow" />
          </div>
          <div>
            <span v-if="workflow.completedAt">
              Completed
              {{ $filters.formatDateTime(workflow.completedAt) }}
              (took
              {{
                $filters.formatDuration(
                  workflow.startedAt,
                  workflow.completedAt,
                )
              }})
            </span>
            <span v-else>
              Started {{ $filters.formatDateTime(workflow.startedAt) }}
            </span>
          </div>
        </div>
      </button>
    </h4>
    <div
      v-if="workflow.tasks"
      :id="'w-body-' + index"
      class="accordion-collapse collapse"
      :aria-labelledby="'w-heading-' + index"
      data-bs-parent="#workflows"
    >
      <SipReviewAlert
        v-model:expandCounter="expandCounter"
        v-if="
          showSipReviewAlert(workflow) &&
          authStore.checkAttributes(['ingest:sips:review'])
        "
      />
      <AipDeletionReviewAlert
        v-if="
          showAipDeletionReviewAlert(workflow) &&
          (authStore.checkAttributes(['storage:aips:deletion:review']) ||
            authStore.checkAttributes(['storage:aips:deletion:request']))
        "
        :note="workflow.tasks?.[0]?.note || ''"
      />
      <ul class="accordion-body d-flex flex-column gap-1">
        <li
          v-for="(task, index) of tasks"
          :id="'task-' + (tasks.length - index)"
          :key="index"
          class="mb-2 card bg-light"
        >
          <Task :index="tasks.length - index" :task="task" />
        </li>
      </ul>
    </div>
  </div>
</template>
