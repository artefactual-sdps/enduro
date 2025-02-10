<script setup lang="ts">
import UUID from "@/components/UUID.vue";
import { openPackageLocationDialog } from "@/dialogs";
import { useAuthStore } from "@/stores/auth";
import { usePackageStore } from "@/stores/package";
import { ref } from "vue";

const authStore = useAuthStore();
const packageStore = usePackageStore();

let failed = ref<boolean | null>(null);

const choose = async () => {
  failed.value = false;
  const locationId = await openPackageLocationDialog(
    packageStore.current?.locationId,
  );
  if (!locationId) return;
  const error = await packageStore.move(locationId);
  if (error) {
    failed.value = true;
  }
};
</script>

<template>
  <div class="card mb-3">
    <div class="card-body">
      <div v-if="failed" class="alert alert-danger" role="alert">
        Move operation failed, try again!
      </div>
      <div v-if="packageStore.isMoving" class="alert alert-info" role="alert">
        The package is being moved into a new location.
      </div>
      <h4 class="card-title">Location</h4>
      <p class="card-text">
        <span v-if="packageStore.isRejected">Package rejected.</span>
        <span v-else-if="!packageStore.current?.locationId"
          >Not available yet.</span
        >
        <span v-else><UUID :id="packageStore.current.locationId" /></span>
      </p>
      <div
        class="actions"
        v-if="
          !packageStore.isRejected &&
          authStore.checkAttributes(['package:move'])
        "
      >
        <div class="d-flex flex-wrap gap-2">
          <button
            type="button"
            class="btn btn-primary btn-sm"
            @click="choose"
            :disabled="!packageStore.isMovable"
          >
            <template v-if="packageStore.isMoving">
              <span
                class="spinner-grow spinner-grow-sm me-2"
                role="status"
                aria-hidden="true"
              ></span>
              Moving...
            </template>
            <template v-else>Choose storage location</template>
          </button>
          <button
            v-if="authStore.checkAttributes(['storage:package:download'])"
            :class="{
              btn: true,
              'btn-primary': true,
              'btn-sm': true,
            }"
            type="button"
          >
            Download
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
