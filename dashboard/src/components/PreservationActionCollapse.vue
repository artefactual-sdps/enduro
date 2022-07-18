<script setup lang="ts">
import IconCircleChevronDown from "~icons/akar-icons/circle-chevron-down";
import IconCircleChevronUp from "~icons/akar-icons/circle-chevron-up";
import PackageReviewAlert from "@/components/PackageReviewAlert.vue";
import { api } from "@/client";
import { onMounted, watch } from "vue";
import useEventListener from "@/composables/useEventListener";
import Collapse from "bootstrap/js/dist/collapse";

const { action, index, toggleAll } = defineProps<{
  action: api.EnduroPackagePreservationActionResponseBody;
  index: number;
  toggleAll: boolean | null;
}>();

const emit = defineEmits<{
  (e: "update:toggleAll", value: null): void;
}>();

let shown = $ref<boolean>(false);

const el = $ref<HTMLElement | null>(null);
useEventListener($$(el), "shown.bs.collapse", () => {
  shown = true;
  emit("update:toggleAll", null);
});
useEventListener($$(el), "hidden.bs.collapse", () => {
  shown = false;
  emit("update:toggleAll", null);
});

let col = <Collapse | null>null;
onMounted(() => {
  if (!el) return;
  col = new Collapse(el, { toggle: false });
});

watch($$(toggleAll), () => {
  if (toggleAll === null) return;
  toggleAll ? col?.show() : col?.hide();
});

let expandCounter = $ref<number>(0);
watch($$(expandCounter), () => col?.show());

const getPreservationActionLabel = (value: api.EnduroPackagePreservationActionResponseBodyTypeEnum) => {
  switch (value) {
    case api.EnduroPackagePreservationActionResponseBodyTypeEnum.CreateAip:
      return "Create AIP";
    case api.EnduroPackagePreservationActionResponseBodyTypeEnum.MovePackage:
      return "Move package";
    default:
      return value;
  }
};
</script>

<template>
  <div>
    <hr />
    <div class="mb-3">
      <div class="d-flex">
        <h3 class="h4">
          {{ getPreservationActionLabel(action.type) }}
          <span
            class="badge"
            :class="$filters.formatPreservationActionStatus(action.status)"
            >{{ action.status }}</span
          >
        </h3>
        <button
          class="btn btn-sm btn-link text-decoration-none ms-auto p-0"
          type="button"
          data-bs-toggle="collapse"
          :data-bs-target="'#preservation-actions-table-' + index"
          aria-expanded="false"
          :aria-controls="'preservation-actions-table-' + index"
          v-if="action.tasks"
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
          api.EnduroPackagePreservationActionResponseBodyTypeEnum.CreateAip &&
        action.status ==
          api.EnduroPackagePreservationActionResponseBodyStatusEnum.Pending
      "
    />
    -->
    <PackageReviewAlert v-model:expandCounter="expandCounter" />

    <div ref="el" :id="'preservation-actions-table-' + index" class="collapse">
      <table class="table table-bordered table-sm" v-if="action.tasks">
        <thead>
          <tr>
            <th scope="col">Task #</th>
            <th scope="col">Name</th>
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
            <td>
              <span
                class="badge"
                :class="$filters.formatPreservationTaskStatus(task.status)"
                >{{ task.status }}</span
              >
            </td>
            <td>{{ task.note }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
