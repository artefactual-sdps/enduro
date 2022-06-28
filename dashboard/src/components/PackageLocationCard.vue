<script setup lang="ts">
import { openPackageLocationDialog } from "@/dialogs";
import { useStorageStore } from "@/stores/storage";

const storageStore = useStorageStore();

const choose = async () => {
  const location = await openPackageLocationDialog(
    storageStore.package?.location
  );
  console.log(location);
  // TODO: packageStore.current.move
};
</script>

<template>
  <div class="card mb-3">
    <div class="card-body">
      <h5 class="card-title">Location</h5>
      <p class="card-text">
        <span v-if="!storageStore.package?.location">Not available yet.</span>
        <span v-else>{{ storageStore.package.location }}</span>
      </p>
      <div v-if="storageStore.package?.location">
        <button type="button" class="btn btn-primary btn-sm" @click="choose">
          Choose storage location
        </button>
      </div>
    </div>
  </div>
</template>
