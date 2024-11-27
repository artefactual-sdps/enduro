<script setup lang="ts">
import type { api } from "@/client";
import PackageReviewAlert from "@/components/PackageReviewAlert.vue";
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
      :id="'preservation-actions-table-' + index"
      class="collapse table-responsive mb-3"
    >
      <table class="table table-bordered table-sm mb-0" v-if="action.tasks">
        <thead>
          <tr>
            <th scope="col">Task #</th>
            <th scope="col">Name</th>
            <th scope="col">Start</th>
            <th scope="col">End</th>
            <th scope="col">Outcome</th>
            <th scope="col">Notes</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(task, index) in action.tasks.slice().reverse()"
            :key="action.id"
          >
            <td>{{ action.tasks.length - index }}</td>
            <td>{{ task.name }}</td>
            <td>{{ $filters.formatDateTime(task.startedAt) }}</td>
            <td>{{ $filters.formatDateTime(task.completedAt) }}</td>
            <td>
              <StatusBadge :status="task.status" />
            </td>
            <td class="line-break">{{ task.note }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
