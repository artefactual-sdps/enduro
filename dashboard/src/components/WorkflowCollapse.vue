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
  of: number;
}>();

const { workflow, index } = toRefs(props);

let expandCounter = ref<number>(0);

const showSipReviewAlert = () => {
  return (
    workflow.value.type ==
      api.EnduroIngestSipWorkflowTypeEnum.CreateAndReviewAip &&
    workflow.value.status == api.EnduroIngestSipWorkflowStatusEnum.Pending
  );
};

const showAipDeletionReviewAlert = () => {
  return (
    workflow.value.type == api.EnduroStorageAipWorkflowTypeEnum.DeleteAip &&
    workflow.value.status == api.EnduroStorageAipWorkflowStatusEnum.Pending
  );
};

const showTasks = computed(() => {
  if (!workflow.value.tasks) {
    return false;
  }

  // Show tasks if there is only one workflow.
  if (props.of === 1) {
    return true;
  }

  // Show tasks if the workflow is "in progress".
  if (
    api.instanceOfEnduroIngestSipWorkflow(workflow.value) &&
    workflow.value.status === api.EnduroIngestSipWorkflowStatusEnum.InProgress
  ) {
    return true;
  }
  if (
    api.instanceOfEnduroStorageAipWorkflow(workflow.value) &&
    workflow.value.status === api.EnduroStorageAipWorkflowStatusEnum.InProgress
  ) {
    return true;
  }

  // Show tasks if a user decision is required.
  if (showSipReviewAlert() || showAipDeletionReviewAlert()) {
    return true;
  }

  return false;
});
</script>

<template>
  <div class="accordion-item border-0 mb-2">
    <h4 class="accordion-header" :id="'wf' + index + '-heading'">
      <button
        ref="wfBtn"
        v-if="workflow.tasks"
        :class="[
          'accordion-button',
          {
            collapsed: !showTasks,
          },
        ]"
        type="button"
        data-bs-toggle="collapse"
        :data-bs-target="'#wf' + index + '-tasks'"
        :aria-expanded="showTasks ? 'true' : 'false'"
        :aria-controls="'wf' + index + '-tasks'"
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
      :id="'wf' + index + '-tasks'"
      :class="[
        'accordion-collapse',
        'collapse',
        {
          show: showTasks,
        },
      ]"
      :aria-labelledby="'wf' + index + '-heading'"
      data-bs-parent="#workflows"
    >
      <SipReviewAlert
        v-model:expandCounter="expandCounter"
        v-if="
          showSipReviewAlert() &&
          authStore.checkAttributes(['ingest:sips:review'])
        "
      />
      <AipDeletionReviewAlert
        v-if="
          showAipDeletionReviewAlert() &&
          (authStore.checkAttributes(['storage:aips:deletion:review']) ||
            authStore.checkAttributes(['storage:aips:deletion:request']))
        "
        :note="workflow.tasks?.[0]?.note || ''"
      />
      <ul class="accordion-body d-flex flex-column gap-1">
        <li
          v-for="(task, idx) of tasks"
          :id="'wf' + index + '-task' + (tasks.length - idx)"
          :key="idx"
          class="mb-2 card bg-light"
        >
          <Task :index="tasks.length - idx" :task="task" />
        </li>
      </ul>
    </div>
  </div>
</template>
