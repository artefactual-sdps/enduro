<script setup lang="ts">
import { computed, ref } from "vue";

import EmailLinkedText from "@/components/EmailLinkedText.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import { formatDateTime } from "@/composables/format";
import type {
  EnduroIngestSipTask,
  EnduroStorageAipTask,
} from "@/openapi-generator";

const isComplete = (task: EnduroIngestSipTask | EnduroStorageAipTask) => {
  return task.status == "done" || task.status == "error";
};

const props = defineProps<{
  index: number;
  task: EnduroIngestSipTask | EnduroStorageAipTask;
}>();

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
  <div :id="'pt-' + props.index + '-body'" class="card-body">
    <div class="d-flex flex-row align-start gap-3">
      <div class="fd-flex">
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
          <div :id="'pt-' + index + '-time'" class="me-3">
            <span v-if="!isComplete(task) && formatDateTime(task.startedAt)">
              Started:
              {{ formatDateTime(task.startedAt) }}
            </span>
            <span v-if="isComplete(task) && formatDateTime(task.completedAt)">
              Completed:
              {{ formatDateTime(task.completedAt) }}
            </span>
          </div>
        </div>
        <div class="flex-grow-1">
          <span :id="'pt-' + index + '-note'">
            <EmailLinkedText :text="noteData.note" />
          </span>
          <span v-if="noteData.more">
            <span v-show="!isOpen">... </span>
            <Transition name="fade">
              <p
                v-show="isOpen"
                :id="'pt-' + index + '-note-more'"
                class="line-break"
              >
                <EmailLinkedText :text="noteData.more" />
              </p>
            </Transition>
            <a
              :id="'pt-' + index + '-note-toggle'"
              :aria-controls="'pt-' + index + '-note-more'"
              aria-label="Toggle display of additional notes"
              href="#"
              @click.prevent="toggle"
            >
              {{ isOpen ? "Show less" : "Show more" }}
            </a>
          </span>
        </div>
      </div>
      <div class="d-flex pt-1">
        <StatusBadge :status="task.status" type="workflow" />
      </div>
    </div>
  </div>
</template>
