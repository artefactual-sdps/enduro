<script setup lang="ts">
import { ref } from "vue";

import StatusBadge from "@/components/StatusBadge.vue";
import type { EnduroPackagePreservationTask } from "@/openapi-generator";

class Card {
  isOpen: boolean;
  note: string;
  more: string;

  constructor(task: EnduroPackagePreservationTask) {
    this.isOpen = false;

    if (task.note?.includes("\n")) {
      const [firstLine, ...remainingLines] = task.note.split("\n");
      this.note = firstLine;
      this.more = remainingLines.join("\n");
    } else {
      this.note = task.note ? task.note : "";
      this.more = "";
    }
  }

  toggle() {
    if (!this.more) {
      return;
    }

    this.isOpen = !this.isOpen;
  }
}

const isComplete = (task: EnduroPackagePreservationTask) => {
  return task.status == "done" || task.status == "error";
};

const props = defineProps<{
  index: number;
  task: EnduroPackagePreservationTask;
}>();

const card = ref(new Card(props.task));
</script>

<template>
  <div class="card-body">
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
          <div class="me-3">
            <span
              v-if="
                !isComplete(task) && $filters.formatDateTime(task.startedAt)
              "
            >
              Started:
              {{ $filters.formatDateTime(task.startedAt) }}
            </span>
            <span
              v-if="
                isComplete(task) && $filters.formatDateTime(task.completedAt)
              "
            >
              Completed:
              {{ $filters.formatDateTime(task.completedAt) }}
            </span>
          </div>
        </div>
        <div class="flex-grow-1">
          <span :id="'pt-note-' + index">{{ card.note }}</span>
          <span v-if="card.more">
            <span v-show="!card.isOpen">... </span>
            <Transition name="fade">
              <p
                v-show="card.isOpen"
                :id="'pt-note-' + index + '-more'"
                class="line-break"
              >
                {{ card.more }}
              </p>
            </Transition>
            <a
              :aria-controls="'pt-note-' + index + '-more'"
              aria-label="Toggle display of additional notes"
              @click.prevent="card.toggle"
              href="#"
            >
              Show {{ card.isOpen ? "less" : "more" }}
            </a>
          </span>
        </div>
      </div>
      <div class="d-flex pt-1">
        <StatusBadge :status="task.status" />
      </div>
    </div>
  </div>
</template>
