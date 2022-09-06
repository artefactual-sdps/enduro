<script setup lang="ts">
import { api } from "@/client";
import PackageReviewAlert from "@/components/PackageReviewAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import Collapse from "bootstrap/js/dist/collapse";
import { onMounted, watch } from "vue";
import IconCircleChevronDown from "~icons/akar-icons/circle-chevron-down";
import IconCircleChevronUp from "~icons/akar-icons/circle-chevron-up";

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
let col: Collapse | null = null;

onMounted(() => {
  if (!el) return;
  col = new Collapse(el, { toggle: false });
});

const toggle = () => {
  shown ? hide() : show();
};

const show = () => {
  col?.show();
  emit("update:toggleAll", null);
  shown = true;
};

const hide = () => {
  col?.hide();
  emit("update:toggleAll", null);
  shown = false;
};

watch($$(toggleAll), () => {
  if (toggleAll === null) return;
  toggleAll ? show() : hide();
});

let expandCounter = $ref<number>(0);
watch($$(expandCounter), () => show());
</script>

<template>
  <div class="mb-3 border-bottom container-xxl">
    <div class="mb-3 row">
      <div :class="shown ? 'col-6' : 'col-12'">
        <div>
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
                <IconCircleChevronUp
                  style="font-size: 2em; transform: rotate(270deg)"
                  aria-hidden="true"
                />
                <span class="visually-hidden"
                  >Collapse preservation tasks table</span
                >
              </span>
              <span v-else>
                <IconCircleChevronDown
                  style="font-size: 2em; transform: rotate(270deg)"
                  aria-hidden="true"
                />
                <span class="visually-hidden"
                  >Expand preservation tasks table</span
                >
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
      </div>

      <div :class="shown ? 'col-6' : 'd-none'">
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
                  <StatusBadge :status="task.status" />
                </td>
                <td>{{ task.note }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
