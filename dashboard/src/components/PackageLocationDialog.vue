<script setup lang="ts">
import { useStorageStore } from "@/stores/storage";
import Modal from "bootstrap/js/dist/modal";
import { ref, onMounted } from "vue";
import { closeDialog as cd } from "vue3-promise-dialog";

const el = ref<HTMLElement | null>(null);
const modal = ref<Modal | null>(null);
const storageStore = useStorageStore();
storageStore.fetchLocations();

onMounted(() => {
  if (!el.value) return;
  modal.value = new Modal(el.value);
  modal.value.show();
});

const closeDialog = (data?: any) => {
  modal.value?.hide();
  cd(data ? data : null);
};

const onChoose = (locationName: string) => {
  closeDialog(locationName);
};

const onClose = () => {
  closeDialog();
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
          <table class="table table-rounded table-hover table-sm">
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
              >
                <td>{{ item.name }}</td>
                <td><span class="badge bg-success">READY</span></td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="onClose">
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
tbody tr {
  cursor: pointer;
}
</style>
