<script setup lang="ts">
import useEventListener from "@/composables/useEventListener";
import { useStorageStore } from "@/stores/storage";
import Modal from "bootstrap/js/dist/modal";
import { ref, onMounted } from "vue";
import { closeDialog } from "vue3-promise-dialog";

const props = defineProps({
  currentLocation: { type: String, required: false },
});

const el = ref<HTMLElement | null>(null);
const modal = ref<Modal | null>(null);

const storageStore = useStorageStore();
storageStore.fetchLocations();

onMounted(() => {
  if (!el.value) return;
  modal.value = new Modal(el.value);
  modal.value.show();
});

let data: string | null = null;

useEventListener(el, "hidden.bs.modal", (e) => {
  closeDialog(data);
});

const onChoose = (locationName: string) => {
  if (locationName === props.currentLocation) return;
  data = locationName;
  modal.value?.hide();
};
</script>

<template>
  <div class="modal" tabindex="-1" ref="el">
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Choose location</h5>
        </div>
        <div class="modal-body">
          <table class="table table-rounded table-hover table-sm table-linked">
            <thead>
              <tr>
                <th>Location name</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(item, index) in storageStore.locations"
                @click="onChoose(item.name)"
                :class="[item.name == props.currentLocation ? 'disabled' : '']"
              >
                <td>{{ item.name }}</td>
                <td><span class="badge bg-success">READY</span></td>
              </tr>
            </tbody>
          </table>
          <small class="text-muted" v-if="props.currentLocation">
            The current location is {{ props.currentLocation }}.
          </small>
        </div>
        <div class="modal-footer">
          <button
            type="button"
            class="btn btn-secondary"
            data-bs-dismiss="modal"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.disabled {
  cursor: not-allowed;
  background-color: #f5f5f5;
  font-weight: bold;
}
</style>
