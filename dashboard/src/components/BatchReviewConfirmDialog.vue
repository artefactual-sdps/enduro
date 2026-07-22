<script setup lang="ts">
import SafeHtml from "@/components/SafeHtml.vue";
import useDialog from "@/dialogs/useDialog";

defineProps<{
  heading: string;
  bodyHtml: string;
  confirmClass: "btn-primary" | "btn-danger";
}>();

const emit = defineEmits<{
  resolve: [confirmed: boolean];
}>();

const titleId = "batch-review-confirm-dialog-title";
const bodyId = "batch-review-confirm-dialog-body";

const { element, close } = useDialog(emit, false);
</script>

<template>
  <div
    ref="element"
    class="modal"
    tabindex="-1"
    role="dialog"
    aria-modal="true"
    :aria-labelledby="titleId"
    :aria-describedby="bodyId"
  >
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h1 :id="titleId" class="modal-title fs-5">{{ heading }}</h1>
          <button
            type="button"
            class="btn-close"
            aria-label="Close"
            @click="close(false)"
          ></button>
        </div>
        <div :id="bodyId" class="modal-body">
          <SafeHtml :html="bodyHtml" />
        </div>
        <div class="modal-footer">
          <button
            type="button"
            :class="['btn', confirmClass]"
            @click="close(true)"
          >
            Yes
          </button>
          <button type="button" class="btn btn-secondary" @click="close(false)">
            No
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
