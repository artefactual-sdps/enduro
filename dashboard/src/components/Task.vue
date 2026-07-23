<script setup lang="ts">
import { computed, ref } from "vue";

import EmailLinkedText from "@/components/EmailLinkedText.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import { formatDateTime } from "@/composables/format";
import type {
  EnduroIngestSipTask,
  EnduroStorageAipTask,
} from "@/openapi-generator";

const props = defineProps<{
  index: number;
  task: EnduroIngestSipTask | EnduroStorageAipTask;
}>();

const idPrefix = computed(() => `pt-${props.task.uuid}`);
const startedAt = computed(() => formatDateTime(props.task.startedAt));
const completedAt = computed(() => formatDateTime(props.task.completedAt));
const isOpen = ref(false);

const noteData = computed(() => {
  const taskNote = props.task.note;

  if (taskNote?.includes("\n")) {
    const [firstLine, ...remainingLines] = taskNote.split("\n");
    return {
      note: firstLine,
      more: remainingLines.join("\n"),
    };
  } else {
    return {
      note: taskNote ? taskNote : "",
      more: "",
    };
  }
});

const toggle = () => {
  if (!noteData.value.more) {
    return;
  }
  isOpen.value = !isOpen.value;
};
</script>

<template>
  <div :id="`${idPrefix}-body`" class="card-body">
    <div class="d-flex flex-row align-items-start gap-3">
      <div class="d-flex">
        <span
          class="fs-6 badge rounded-pill border border-primary text-primary"
        >
          {{ index }}
        </span>
      </div>
      <div
        class="d-flex flex-column flex-grow-1 align-content-stretch min-w-0 gap-2"
      >
        <div class="d-flex flex-wrap pt-1">
          <div class="me-auto text-truncate fw-bold">
            {{ task.name }}
          </div>
          <div :id="`${idPrefix}-time`" class="me-3">
            <span v-if="completedAt">
              Completed:
              {{ completedAt }}
            </span>
            <span v-else-if="startedAt">
              Started:
              {{ startedAt }}
            </span>
          </div>
        </div>
        <div class="flex-grow-1">
          <span :id="`${idPrefix}-note`">
            <EmailLinkedText :text="noteData.note" />
          </span>
          <span v-if="noteData.more">
            <span v-show="!isOpen">... </span>
            <Transition name="fade">
              <p
                v-show="isOpen"
                :id="`${idPrefix}-note-more`"
                class="line-break"
              >
                <EmailLinkedText :text="noteData.more" />
              </p>
            </Transition>
            <button
              :id="`${idPrefix}-note-toggle`"
              class="btn btn-link p-0 align-baseline"
              type="button"
              :aria-controls="`${idPrefix}-note-more`"
              :aria-expanded="isOpen ? 'true' : 'false'"
              aria-label="Toggle display of additional notes"
              @click="toggle"
            >
              {{ isOpen ? "Show less" : "Show more" }}
            </button>
          </span>
        </div>
      </div>
      <div class="d-flex pt-1">
        <StatusBadge :status="task.status" type="workflow" />
      </div>
    </div>
  </div>
</template>
