<script setup lang="ts">
import Modal from "bootstrap/js/dist/modal";
import { onMounted, ref } from "vue";
import { closeDialog } from "vue3-promise-dialog";

import useEventListener from "@/composables/useEventListener";
import { useAipStore } from "@/stores/aip";

const aipStore = useAipStore();

const el = ref<HTMLElement | null>(null);
const modal = ref<Modal | null>(null);
const reason = ref<string>("");
const submit = ref<boolean>(false);
const titleId = "aip-deletion-request-dialog-title";
const bodyId = "aip-deletion-request-dialog-body";

onMounted(() => {
  if (!el.value) return;
  modal.value = new Modal(el.value);
  modal.value.show();
});

useEventListener(el, "hidden.bs.modal", () => {
  if (submit.value) {
    closeDialog(reason.value);
  } else {
    closeDialog(null);
  }
  submit.value = false;
});

const request = () => {
  submit.value = true;
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
          <h1 :id="titleId" class="modal-title fs-5">Delete AIP</h1>
        </div>
        <form @submit.prevent="request">
          <div :id="bodyId" class="modal-body">
            <p>
              This will initiate a <strong>deletion request</strong> for the
              following AIP and any related replicas.
            </p>
            <p>
              <strong>{{ aipStore.current?.name }}</strong>
            </p>
            <p>
              You MUST provide a reason for the AIP deletion. The request will
              be reviewed before deletion.
            </p>
            <div>
              <label for="reason" class="form-label">Reason:</label>
              <textarea
                id="reason"
                v-model="reason"
                class="form-control"
                rows="3"
                required
                minlength="10"
                maxlength="500"
              />
              <div class="form-text text-end">
                Reason must be between 10 and 500 characters.
              </div>
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
            <button type="submit" class="btn btn-danger">
              Request deletion
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
