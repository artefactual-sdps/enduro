<script setup lang="ts">
import { openPackageLocationDialog } from "@/dialogs";
import { usePackageStore } from "@/stores/package";

const packageStore = usePackageStore();

let failed = $ref<boolean | null>(null);

const choose = async () => {
  failed = false;
  const locationName = await openPackageLocationDialog(
    packageStore.current?.location
  );
  if (!locationName) return;
  const error = await packageStore.move(locationName);
  if (error) {
    failed = true;
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
      <h5 class="card-title">Location</h5>
      <p class="card-text">
        <span v-if="packageStore.isRejected">Package rejected.</span>
        <span v-else-if="!packageStore.current?.location"
          >Not available yet.</span
        >
        <span v-else>{{ packageStore.current.location }}</span>
      </p>
      <div class="actions" v-if="!packageStore.isRejected">
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
      </div>
    </div>
  </div>
</template>
