<script setup lang="ts">
import { computed, ref, toRefs } from "vue";

import type { api } from "@/client";
import PreservationTask from "@/components/PreservationTask.vue";
import SipReviewAlert from "@/components/SipReviewAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import { useAuthStore } from "@/stores/auth";

const authStore = useAuthStore();
const tasks = computed<api.EnduroIngestSipPreservationTask[]>(() => {
  if (!props.action.tasks) {
    return [];
  }

  // Show the last task first.
  return props.action.tasks.slice().reverse();
});

const props = defineProps<{
  action: api.EnduroIngestSipPreservationAction;
  index: number;
}>();

const { action, index } = toRefs(props);

let expandCounter = ref<number>(0);
</script>

<template>
  <div class="accordion-item border-0 mb-2">
    <h4 class="accordion-header" :id="'pa-heading-' + index">
      <button
        v-if="action.tasks"
        class="accordion-button collapsed"
        type="button"
        data-bs-toggle="collapse"
        :data-bs-target="'#pa-body-' + index"
        aria-expanded="false"
        :aria-controls="'pa-body-' + index"
      >
        <div class="d-flex flex-column">
          <div class="h4">
            {{ $filters.getPreservationActionLabel(action.type) }}
            <StatusBadge :status="action.status" />
          </div>
          <div>
            <span v-if="action.completedAt">
              Completed
              {{ $filters.formatDateTime(action.completedAt) }}
              (took
              {{
                $filters.formatDuration(action.startedAt, action.completedAt)
              }})
            </span>
            <span v-else>
              Started {{ $filters.formatDateTime(action.startedAt) }}
            </span>
          </div>
        </div>
      </button>
    </h4>
    <div
      v-if="action.tasks"
      :id="'pa-body-' + index"
      class="accordion-collapse collapse"
      :aria-labelledby="'pa-heading-' + index"
      data-bs-parent="#preservation-actions"
    >
      <SipReviewAlert
        v-model:expandCounter="expandCounter"
        v-if="authStore.checkAttributes(['ingest:sips:review'])"
      />
      <ul class="accordion-body d-flex flex-column gap-1">
        <li
          v-for="(task, index) of tasks"
          :id="'prestask-' + (tasks.length - index)"
          :key="task.id"
          class="mb-2 card bg-light"
        >
          <PreservationTask :index="tasks.length - index" :task="task" />
        </li>
      </ul>
    </div>
  </div>
</template>
