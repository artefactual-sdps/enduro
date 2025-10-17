<script setup lang="ts">
import Modal from "bootstrap/js/dist/modal";
import { onMounted, ref } from "vue";
import { closeDialog } from "vue3-promise-dialog";

import useEventListener from "@/composables/useEventListener";
import { useLocationStore } from "@/stores/location";

const props = defineProps({
  currentLocationId: { type: String, required: false, default: undefined },
});

const locationStore = useLocationStore();
locationStore.fetchLocations();

const el = ref<HTMLElement | null>(null);
const modal = ref<Modal | null>(null);

onMounted(() => {
  if (!el.value) return;
  modal.value = new Modal(el.value);
  modal.value.show();
});

let data: string | null = null;

useEventListener(el, "hidden.bs.modal", () => {
  closeDialog(data);
});

const onChoose = (locationId: string) => {
  if (locationId === props.currentLocationId) return;
  data = locationId;
  modal.value?.hide();
};
</script>

<template>
  <div ref="el" class="modal" tabindex="-1">
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Choose location</h5>
        </div>
        <div class="modal-body">
          <div class="table-responsive mb-3">
            <table class="table table-sm mb-0">
              <thead>
                <tr>
                  <th>Location name</th>
                  <th>Status</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in locationStore.locations"
                  :key="item.uuid"
                  :class="[
                    item.uuid == props.currentLocationId ? 'current' : '',
                  ]"
                >
                  <td>{{ item.name }}</td>
                  <td>
                    <span class="badge bg-success">READY</span>
                  </td>
                  <td class="text-end">
                    <button
                      v-if="item.uuid != props.currentLocationId"
                      class="btn btn-sm btn-primary"
                      @click="onChoose(item.uuid)"
                    >
                      Move
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <small v-if="props.currentLocationId" class="text-muted">
            The current location is {{ props.currentLocationId }}.
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
.current {
  background-color: #f5f5f5;
}

table .btn {
  font-size: 0.75rem;
  font-weight: bold;
  padding: 0.125rem 0.25rem;
}
</style>
