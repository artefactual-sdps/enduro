<script setup lang="ts">
import { ref } from "vue";

import SafeHtml from "@/components/SafeHtml.vue";
import useBootstrapModal from "@/composables/useBootstrapModal";
import type { BatchReviewConfirmDialogProps } from "@/dialogs/batchReviewConfirm";

defineProps<BatchReviewConfirmDialogProps>();

const emit = defineEmits<{
  resolve: [confirmed: boolean];
}>();

const confirmed = ref(false);
const titleId = "batch-review-confirm-dialog-title";
const bodyId = "batch-review-confirm-dialog-body";

const { element: el, hide } = useBootstrapModal(() => {
  emit("resolve", confirmed.value);
  confirmed.value = false;
});

const confirm = (value: boolean) => {
  confirmed.value = value;
  hide();
};
</script>

<template>
  <div
    ref="el"
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
            @click="confirm(false)"
          ></button>
        </div>
        <div :id="bodyId" class="modal-body">
          <SafeHtml :html="bodyHtml" />
        </div>
        <div class="modal-footer">
          <button
            type="button"
            :class="['btn', confirmClass]"
            @click="confirm(true)"
          >
            Yes
          </button>
          <button
            type="button"
            class="btn btn-secondary"
            @click="confirm(false)"
          >
            No
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
