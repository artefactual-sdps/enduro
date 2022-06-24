<script setup lang="ts">
import { openPackageLocationDialog } from "@/dialogs";
import { usePackageStore } from "@/stores/package";

const packageStore = usePackageStore();

// TODO: packageStore.current.location
const location = undefined;

const choose = async () => {
  await openPackageLocationDialog();
  // TODO: packageStore.current.move
};
</script>

<template>
  <div class="card mb-3" v-if="packageStore.current">
    <div class="card-body">
      <h5 class="card-title">Location</h5>
      <p class="card-text">
        <span v-if="location">{{ location }}</span>
        <span v-else>Not available yet.</span>
      </p>
      <div v-if="location || true">
        <button type="button" class="btn btn-primary btn-sm" @click="choose">
          Choose storage location
        </button>
      </div>
    </div>
  </div>
</template>
