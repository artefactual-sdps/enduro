<script setup lang="ts">
import Modal from "bootstrap/js/dist/modal";
import { onMounted, ref } from "vue";
import { closeDialog } from "vue3-promise-dialog";

import SafeHtml from "@/components/SafeHtml.vue";
import useEventListener from "@/composables/useEventListener";

defineProps<{
  heading: string;
  bodyHtml: string;
  confirmClass: "btn-primary" | "btn-danger";
}>();

const el = ref<HTMLElement | null>(null);
const modal = ref<Modal | null>(null);
const confirmed = ref(false);
const titleId = "batch-review-confirm-dialog-title";
const bodyId = "batch-review-confirm-dialog-body";

onMounted(() => {
  if (!el.value) return;
  modal.value = new Modal(el.value);
  modal.value.show();
});

useEventListener(el, "hidden.bs.modal", () => {
  closeDialog(confirmed.value);
  confirmed.value = false;
});

const confirm = (value: boolean) => {
  confirmed.value = value;
  modal.value?.hide();
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
