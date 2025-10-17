<script setup lang="ts">
import { api } from "@/client";
import StatusBadge from "@/components/StatusBadge.vue";
import IconLink from "~icons/bi/box-arrow-up-right";

const statuses = [
  {
    status: api.EnduroIngestSipTaskStatusEnum.Done,
    description: "The task has completed successfully.",
  },
  {
    status: api.EnduroIngestSipTaskStatusEnum.Failed,
    description:
      "The related package has failed to meet this task's policy-defined criteria.",
  },
  {
    status: api.EnduroIngestSipTaskStatusEnum.InProgress,
    description: "The task is still processing.",
  },
  {
    status: api.EnduroIngestSipTaskStatusEnum.Pending,
    description: "The task is awaiting a user decision.",
  },
  {
    status: api.EnduroIngestSipTaskStatusEnum.Error,
    description:
      "The task has encountered a system error it could not resolve.",
  },
];

const { show = false } = defineProps<{
  show?: boolean;
}>();

const emit = defineEmits<{
  (e: "update:show", value: boolean): void;
}>();
</script>

<template>
  <Transition>
    <div v-show="show" id="workflow-help">
      <div class="card bg-light">
        <div class="card-body">
          <div class="d-flex mb-3">
            <div class="flex-grow-1">
              <p id="workflow-task-definition">
                A <b>task</b> is an operation performed on a file, directory, or
                package in the context of a workflow. In Enduro, a
                <b>workflow</b> is a series of tasks performed on a package to
                support the ingest and long-term preservation.
              </p>
              <p>
                Click on a workflow listed below to expand it and see more
                information on individual tasks run as part of the workflow.
              </p>
            </div>
            <div class="justify-content-end">
              <button
                id="workflow-help-close"
                type="button"
                class="btn-close align-middle"
                aria-label="Close"
                @click="() => emit('update:show', false)"
              />
            </div>
          </div>
          <span class="h5">Task status legend</span>
          <div id="task-status-legend" class="container-fluid border p-2 mb-3">
            <div
              v-for="(item, index) in statuses"
              :key="item.status"
              class="row"
            >
              <div class="col col-md-2 py-2 text-end">
                <StatusBadge :status="item.status" type="workflow" />
              </div>
              <div :id="`badge-${index}-desc`" class="col col-md-10 py-2">
                {{ item.description }}
              </div>
            </div>
          </div>
          <div class="text-end">
            <a
              href="https://enduro.readthedocs.io/user-manual/usage/#view-tasks-in-enduro"
              target="_new"
              >Learn more <IconLink alt="" aria-hidden="true"
            /></a>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.v-enter-active,
.v-leave-active {
  transition: opacity 0.1s ease;
}

.v-enter-from,
.v-leave-to {
  opacity: 0;
}
</style>
