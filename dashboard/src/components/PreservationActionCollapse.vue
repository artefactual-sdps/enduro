<script setup lang="ts">
import { ref, toRefs } from "vue";

import type { api } from "@/client";
import PackageReviewAlert from "@/components/PackageReviewAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import type { EnduroPackagePreservationTask } from "@/openapi-generator";
import { useAuthStore } from "@/stores/auth";

const authStore = useAuthStore();

const props = defineProps<{
  action: api.EnduroPackagePreservationAction;
  index: number;
}>();

const { action, index } = toRefs(props);

let expandCounter = ref<number>(0);

function isComplete(task: EnduroPackagePreservationTask) {
  return task.status == "done" || task.status == "error";
}
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
      class="accordion-collapse collapse bg-light"
      :aria-labelledby="'pa-heading-' + index"
      data-bs-parent="#preservation-actions"
    >
      <div class="accordion-body d-flex flex-column gap-1">
        <PackageReviewAlert
          v-model:expandCounter="expandCounter"
          v-if="authStore.checkAttributes(['package:review'])"
        />
        <div
          v-for="(task, index) in action.tasks.slice().reverse()"
          :key="action.id"
          class="card"
        >
          <div class="card-body">
            <div class="d-flex flex-row align-start gap-3">
              <div class="fd-flex">
                <span
                  class="fs-6 badge rounded-pill border border-primary text-primary"
                >
                  {{ action.tasks.length - index }}
                </span>
              </div>
              <div
                class="d-flex flex-column flex-grow-1 align-content-stretch min-w-0"
              >
                <div class="d-flex flex-wrap pt-1">
                  <div class="me-auto text-truncate fw-bold">
                    {{ task.name }}
                  </div>
                  <div class="me-3">
                    <span
                      v-if="
                        !isComplete(task) &&
                        $filters.formatDateTime(task.startedAt)
                      "
                    >
                      Started:
                      {{ $filters.formatDateTime(task.startedAt) }}
                    </span>
                    <span
                      v-if="
                        isComplete(task) &&
                        $filters.formatDateTime(task.completedAt)
                      "
                    >
                      Completed:
                      {{ $filters.formatDateTime(task.completedAt) }}
                    </span>
                  </div>
                </div>
                <div class="d-flex flex-row gap-4">
                  <div class="flex-grow-1 line-break">
                    {{ task.note }}
                  </div>
                </div>
              </div>
              <div class="d-flex pt-1">
                <StatusBadge :status="task.status" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
