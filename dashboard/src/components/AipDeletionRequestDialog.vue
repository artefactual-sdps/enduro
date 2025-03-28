<script setup lang="ts">
import Modal from "bootstrap/js/dist/modal";
import { onMounted, ref } from "vue";
import { closeDialog } from "vue3-promise-dialog";

import useEventListener from "@/composables/useEventListener";
import { useAipStore } from "@/stores/aip";

const aipStore = useAipStore();

const el = ref<HTMLElement | null>(null);
const modal = ref<Modal | null>(null);

onMounted(() => {
  if (!el.value) return;
  modal.value = new Modal(el.value);
  modal.value.show();
});

let reason: string = "";

useEventListener(el, "hidden.bs.modal", () => {
  closeDialog(reason);
});

const request = () => {
  modal.value?.hide();
};
</script>

<template>
  <div class="modal" tabindex="-1" ref="el">
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Delete AIP</h5>
        </div>
        <div class="modal-body">
          <p>
            This will initiate a <strong>deletion request</strong> for the
            following AIP and any related replicas.
          </p>
          <p>
            <strong>{{ aipStore.current?.name }}</strong>
          </p>
          <p>
            You MUST provide a reason for the deletion. The request will be
            reviewed before deletion.
          </p>
          <div>
            <label for="reason" class="form-label"
              >Reason for AIP deletion</label
            >
            <textarea
              class="form-control"
              id="reason"
              rows="3"
              v-model="reason"
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button
            type="button"
            class="btn btn-secondary"
            data-bs-dismiss="modal"
          >
            Cancel
          </button>
          <button type="button" class="btn btn-danger" @click="request()">
            Request deletion
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
