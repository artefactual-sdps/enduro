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
  <div :id="`${idPrefix}-body`" class="workflow-task">
    <div class="workflow-task-number">
      <span class="badge rounded-pill border border-primary text-primary">
        {{ index }}
      </span>
    </div>
    <div class="workflow-task-name">
      {{ task.name }}
    </div>
    <div :id="`${idPrefix}-time`" class="workflow-task-time">
      <span v-if="completedAt">
        <span class="workflow-task-time-label">Ended </span>
        <span class="workflow-task-time-value">
          {{ completedAt }}
        </span>
      </span>
      <span v-else-if="startedAt">
        <span class="workflow-task-time-label">Started </span>
        <span class="workflow-task-time-value">
          {{ startedAt }}
        </span>
      </span>
      <span v-else class="workflow-task-time-empty" aria-label="No timestamp">
        &mdash;
      </span>
    </div>
    <div v-if="noteData.note || noteData.more" class="workflow-task-note">
      <span :id="`${idPrefix}-note`">
        <EmailLinkedText :text="noteData.note" />
      </span>
      <span v-if="noteData.more">
        <span v-show="!isOpen">... </span>
        <Transition name="fade">
          <p
            v-show="isOpen"
            :id="`${idPrefix}-note-more`"
            class="line-break mb-1"
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
    <div class="workflow-task-status">
      <StatusBadge :status="task.status" type="workflow" />
    </div>
  </div>
</template>

<style scoped>
.workflow-task {
  display: grid;
  grid-template-areas:
    "number name time status"
    ". note note status";
  grid-template-columns: var(
    --workflow-task-columns,
    2.5rem minmax(0, 1fr) 12rem 7.25rem
  );
  align-items: start;
  gap: 0.125rem 0.75rem;
  min-width: 0;
  padding: 0.5rem 0.75rem;
}

.workflow-task-number {
  display: flex;
  grid-area: number;
  justify-content: center;
}

.workflow-task-number .badge {
  min-width: 2rem;
}

.workflow-task-name {
  grid-area: name;
  min-width: 0;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.workflow-task-time {
  grid-area: time;
  color: var(--bs-secondary-color);
  font-size: 0.875rem;
  white-space: nowrap;
}

.workflow-task-note {
  grid-area: note;
  min-width: 0;
  color: var(--bs-secondary-color);
  font-size: 0.875rem;
  overflow-wrap: anywhere;
}

.workflow-task-status {
  grid-area: status;
  justify-self: end;
}

@media (max-width: 991.98px) {
  .workflow-task {
    grid-template-areas:
      "number name status"
      ". time time"
      ". note note";
    grid-template-columns: 2.5rem minmax(0, 1fr) max-content;
  }

  .workflow-task-time {
    white-space: normal;
  }
}
</style>
