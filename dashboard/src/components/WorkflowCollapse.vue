<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, toRefs, watch } from "vue";

import { api } from "@/client";
import AipDeletionReviewAlert from "@/components/AipDeletionReviewAlert.vue";
import SipDecisionAlert from "@/components/SipDecisionAlert.vue";
import SipReviewAlert from "@/components/SipReviewAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import Task from "@/components/Task.vue";
import { useAuthStore } from "@/stores/auth";
import { useSipStore } from "@/stores/sip";

const authStore = useAuthStore();
const sipStore = useSipStore();
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
const workflowHeader = ref<HTMLElement | null>(null);
const workflowHeaderHeight = ref("4.25rem");
const workflowItemStyle = computed(() => ({
  "--workflow-header-height": workflowHeaderHeight.value,
}));
let workflowHeaderObserver: ResizeObserver | undefined;

const updateWorkflowHeaderHeight = () => {
  if (!workflowHeader.value) return;
  workflowHeaderHeight.value = `${workflowHeader.value.getBoundingClientRect().height}px`;
};

onMounted(() => {
  if (!workflowHeader.value || typeof ResizeObserver === "undefined") {
    return;
  }

  updateWorkflowHeaderHeight();
  workflowHeaderObserver = new ResizeObserver(updateWorkflowHeaderHeight);
  workflowHeaderObserver.observe(workflowHeader.value);
});

onBeforeUnmount(() => workflowHeaderObserver?.disconnect());

watch(
  () => workflow.value.status,
  (status, previousStatus) => {
    if (previousStatus === undefined || status === previousStatus) return;
    if (!api.instanceOfEnduroIngestSipWorkflow(workflow.value)) return;
    if (!sipStore.current?.uuid) return;
    if (!authStore.checkAttributes(["ingest:sips:decision"])) return;

    // This is a background refresh after a monitor event. The store handles
    // the error state, so this catch only prevents duplicate global reporting.
    void sipStore
      .fetchCurrentDecision(sipStore.current.uuid)
      .catch(() => undefined);
  },
);

const showSipDecisionAlert = computed(() => {
  return (
    api.instanceOfEnduroIngestSipWorkflow(workflow.value) &&
    workflow.value.status == api.EnduroIngestSipWorkflowStatusEnum.Pending &&
    !!sipStore.currentDecision
  );
});

const showSipReviewAlert = computed(() => {
  return (
    workflow.value.type ==
      api.EnduroIngestSipWorkflowTypeEnum.CreateAndReviewAip &&
    workflow.value.status == api.EnduroIngestSipWorkflowStatusEnum.Pending &&
    !sipStore.currentDecision
  );
});

const showAipDeletionReviewAlert = computed(() => {
  return (
    workflow.value.type == api.EnduroStorageAipWorkflowTypeEnum.DeleteAip &&
    workflow.value.status == api.EnduroStorageAipWorkflowStatusEnum.Pending
  );
});

const showTasks = computed(() => {
  if (!workflow.value.tasks) {
    return false;
  }

  // Show tasks if there is only one workflow.
  if (props.of === 1) {
    return true;
  }

  // Show tasks if the workflow is active or requires attention.
  if (
    api.instanceOfEnduroIngestSipWorkflow(workflow.value) &&
    (workflow.value.status ===
      api.EnduroIngestSipWorkflowStatusEnum.InProgress ||
      workflow.value.status === api.EnduroIngestSipWorkflowStatusEnum.Pending)
  ) {
    return true;
  }
  if (
    api.instanceOfEnduroStorageAipWorkflow(workflow.value) &&
    (workflow.value.status ===
      api.EnduroStorageAipWorkflowStatusEnum.InProgress ||
      workflow.value.status === api.EnduroStorageAipWorkflowStatusEnum.Pending)
  ) {
    return true;
  }

  // Show tasks if a user decision is required.
  if (
    showSipDecisionAlert.value ||
    showSipReviewAlert.value ||
    showAipDeletionReviewAlert.value
  ) {
    return true;
  }

  return false;
});
</script>

<template>
  <div
    class="accordion-item workflow-accordion-item mb-2"
    :style="workflowItemStyle"
  >
    <h4
      :id="'wf' + index + '-heading'"
      ref="workflowHeader"
      class="accordion-header workflow-sticky-header"
    >
      <button
        ref="wfBtn"
        :class="[
          'accordion-button',
          'workflow-accordion-button',
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
        <div class="workflow-summary">
          <div class="workflow-summary-title">
            <span>{{ $filters.getWorkflowLabel(workflow.type) }}</span>
            <StatusBadge :status="workflow.status" type="workflow" />
          </div>
          <div class="workflow-summary-meta">
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
          <div v-if="workflow.tasks" class="workflow-summary-count">
            {{ tasks.length }} {{ tasks.length === 1 ? "task" : "tasks" }}
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
      <SipDecisionAlert
        v-if="
          showSipDecisionAlert &&
          authStore.checkAttributes(['ingest:sips:decision'])
        "
      />
      <SipReviewAlert
        v-if="
          showSipReviewAlert &&
          authStore.checkAttributes(['ingest:sips:review'])
        "
      />
      <AipDeletionReviewAlert
        v-if="
          showAipDeletionReviewAlert &&
          (authStore.checkAttributes(['storage:aips:deletion:review']) ||
            authStore.checkAttributes(['storage:aips:deletion:request']))
        "
        :note="workflow.tasks?.[0]?.note || ''"
      />
      <div class="accordion-body workflow-task-table">
        <div class="workflow-task-list-header" aria-hidden="true">
          <span class="text-center">#</span>
          <span>Task</span>
          <span>Time</span>
          <span class="text-end">Status</span>
        </div>
        <ul class="workflow-task-list">
          <li
            v-for="(task, idx) of tasks"
            :id="'wf' + index + '-task' + (tasks.length - idx)"
            :key="task.uuid"
            class="workflow-task-list-item"
          >
            <Task :index="tasks.length - idx" :task="task" />
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.accordion-item.workflow-accordion-item {
  border: 1px solid var(--bs-border-color);
  border-radius: var(--bs-accordion-border-radius);
}

.workflow-sticky-header {
  position: sticky;
  top: $header-height;
  z-index: $zindex-sticky - 1;
  background: var(--bs-body-bg);
}

.workflow-accordion-item .collapsing {
  transition: none;
}

.workflow-accordion-button {
  padding-block: 0.75rem;
  border: 0;
}

.workflow-accordion-button:hover,
.workflow-accordion-button:focus {
  background-color: var(--bs-body-bg);
}

.workflow-accordion-button:focus {
  box-shadow: none;
}

.workflow-accordion-button:focus-visible {
  outline: 2px solid var(--bs-primary);
  outline-offset: -2px;
}

.workflow-accordion-button:not(.collapsed) {
  box-shadow: none;
}

.workflow-summary {
  display: grid;
  flex: 1;
  grid-template-areas:
    "title count"
    "meta count";
  grid-template-columns: minmax(0, 1fr) max-content;
  gap: 0.125rem 1rem;
  min-width: 0;
  margin-right: 0.75rem;
  text-align: left;
}

.workflow-summary-title {
  display: flex;
  grid-area: title;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
  font-size: 1.125rem;
  font-weight: 600;
}

.workflow-summary-meta,
.workflow-summary-count {
  color: var(--bs-secondary-color);
  font-size: 0.875rem;
  font-weight: 400;
}

.workflow-summary-meta {
  grid-area: meta;
}

.workflow-summary-count {
  grid-area: count;
  align-self: center;
  white-space: nowrap;
}

.workflow-task-table {
  --workflow-task-columns: 2.5rem minmax(0, 1fr) 12rem 7.25rem;

  padding: 0;
}

.workflow-task-list-header {
  position: sticky;
  top: calc(#{$header-height} + var(--workflow-header-height));
  z-index: $zindex-sticky - 2;
  display: grid;
  grid-template-columns: var(--workflow-task-columns);
  gap: 0.75rem;
  padding: 0.375rem 0.75rem;
  color: var(--bs-secondary-color);
  background: var(--bs-tertiary-bg);
  border-top: 1px solid var(--bs-border-color);
  border-bottom: 1px solid var(--bs-border-color);
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.025em;
  text-transform: uppercase;
}

.workflow-task-list {
  padding: 0;
  margin: 0;
  overflow: hidden;
  list-style: none;
}

.workflow-task-list-item {
  background: var(--bs-tertiary-bg);
}

.workflow-task-list-item:nth-child(even) {
  background: var(--bs-body-bg);
}

.workflow-task-list-item + .workflow-task-list-item {
  border-top: 1px solid var(--bs-border-color);
}

@media (max-width: 991.98px) {
  .workflow-task-list-header {
    display: none;
  }

  .workflow-task-list {
    border-top: 1px solid var(--bs-border-color);
  }
}

@media (max-width: 575.98px) {
  .workflow-summary {
    grid-template-areas:
      "title title"
      "meta count";
  }
}
</style>
