<script setup lang="ts">
import { ref } from "vue";

import useBootstrapModal from "@/composables/useBootstrapModal";
import { useAipStore } from "@/stores/aip";

const emit = defineEmits<{
  resolve: [reason: string | null];
}>();

const aipStore = useAipStore();

const reason = ref<string>("");
const submit = ref<boolean>(false);
const titleId = "aip-deletion-request-dialog-title";
const bodyId = "aip-deletion-request-dialog-body";

const { element: el, hide } = useBootstrapModal(() => {
  emit("resolve", submit.value ? reason.value : null);
  submit.value = false;
});

const request = () => {
  submit.value = true;
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
