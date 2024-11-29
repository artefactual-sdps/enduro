<script setup lang="ts">
import type { api } from "@/client";
import PackageReviewAlert from "@/components/PackageReviewAlert.vue";
import type {
  EnduroPackagePreservationTask,
  EnduroPackagePreservationTaskStatusEnum,
} from "@/openapi-generator";
import StatusBadge from "@/components/StatusBadge.vue";
import { useAuthStore } from "@/stores/auth";
import Collapse from "bootstrap/js/dist/collapse";
import { onMounted, watch, ref, toRefs } from "vue";
import IconCircleChevronDown from "~icons/akar-icons/circle-chevron-down";
import IconCircleChevronUp from "~icons/akar-icons/circle-chevron-up";

const authStore = useAuthStore();

const props = defineProps<{
  action: api.EnduroPackagePreservationAction;
  index: number;
  toggleAll: boolean | null;
}>();

const { action, index, toggleAll } = toRefs(props);

const emit = defineEmits<{
  (e: "update:toggleAll", value: null): void;
}>();

let shown = ref<boolean>(false);
const el = ref<HTMLElement | null>(null);
let col: Collapse | null = null;

onMounted(() => {
  if (!el.value) return;
  col = new Collapse(el.value, { toggle: false });
});

const toggle = () => {
  shown.value ? hide() : show();
};

const show = () => {
  col?.show();
  emit("update:toggleAll", null);
  shown.value = true;
};

const hide = () => {
  col?.hide();
  emit("update:toggleAll", null);
  shown.value = false;
};

watch(toggleAll, () => {
  if (toggleAll.value === null) return;
  toggleAll.value ? show() : hide();
});

let expandCounter = ref<number>(0);
watch(expandCounter, () => show());

function isComplete(task: EnduroPackagePreservationTask) {
  return task.status == "done" || task.status == "error";
}
</script>

<template>
  <div>
    <hr />
    <div class="mb-3">
      <div class="d-flex">
        <h3 class="h4">
          {{ $filters.getPreservationActionLabel(action.type) }}
          <StatusBadge :status="action.status" />
        </h3>
        <button
          class="btn btn-sm btn-link text-primary text-decoration-none ms-auto p-0"
          type="button"
          aria-expanded="false"
          :aria-controls="'preservation-actions-table-' + index"
          v-if="action.tasks"
          @click="toggle"
        >
          <span v-if="shown">
            <IconCircleChevronUp style="font-size: 2em" aria-hidden="true" />
            <span class="visually-hidden"
              >Collapse preservation tasks table</span
            >
          </span>
          <span v-else>
            <IconCircleChevronDown style="font-size: 2em" aria-hidden="true" />
            <span class="visually-hidden">Expand preservation tasks table</span>
          </span>
        </button>
      </div>
      <span v-if="action.completedAt">
        Completed
        {{ $filters.formatDateTime(action.completedAt) }}
        (took
        {{ $filters.formatDuration(action.startedAt, action.completedAt) }})
      </span>
      <span v-else>
        Started {{ $filters.formatDateTime(action.startedAt) }}
      </span>
    </div>

    <!--
    <PackageReviewAlert
      v-model:expandCounter="expandCounter"
      v-if="
        action.type ==
          api.EnduroPackagePreservationActionTypeEnum.CreateAip &&
        action.status ==
          api.EnduroPackagePreservationActionStatusEnum.Pending
      "
    />
    -->
    <PackageReviewAlert
      v-model:expandCounter="expandCounter"
      v-if="authStore.checkAttributes(['package:review'])"
    />

    <div
      ref="el"
      :id="'preservation-actions-' + index"
      class="collapse mb-3"
      v-if="action.tasks"
    >
      <div
        v-for="(task, index) in action.tasks.slice().reverse()"
        :key="action.id"
        class="mb-2 card"
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
                    Started: {{ $filters.formatDateTime(task.startedAt) }}
                  </span>
                  <span
                    v-if="
                      isComplete(task) &&
                      $filters.formatDateTime(task.completedAt)
                    "
                  >
                    Completed: {{ $filters.formatDateTime(task.completedAt) }}
                  </span>
                </div>
              </div>
              <div class="d-flex flex-row gap-4">
                <div class="flex-grow-1 line-break">{{ task.note }}</div>
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
</template>
