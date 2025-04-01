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
  <div class="modal" tabindex="-1" ref="el">
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Delete AIP</h5>
        </div>
        <form @submit.prevent="request">
          <div class="modal-body">
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
                class="form-control"
                id="reason"
                rows="3"
                v-model="reason"
                required
                minlength="10"
                maxlength="500"
              ></textarea>
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
