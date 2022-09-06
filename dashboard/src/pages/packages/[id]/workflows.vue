<script setup lang="ts">
import PreservationActionCollapse from "@/components/PreservationActionCollapse.vue";
import { usePackageStore } from "@/stores/package";

const packageStore = usePackageStore();

let toggleAll = $ref<boolean | null>(false);
</script>

<template>
  <div v-if="packageStore.current">
    <div class="d-flex">
      <div
        class="align-self-end ms-auto d-flex"
        v-if="
          packageStore.current_preservation_actions?.actions &&
          packageStore.current_preservation_actions.actions.length > 1
        "
      >
        <button
          class="btn btn-sm btn-link p-0"
          type="button"
          @click="toggleAll = true"
        >
          Expand all
        </button>
        <span class="px-1">|</span>
        <button
          class="btn btn-sm btn-link p-0"
          type="button"
          @click="toggleAll = false"
        >
          Collapse all
        </button>
      </div>
    </div>

    <PreservationActionCollapse
      :action="action"
      :index="index"
      v-model:toggleAll="toggleAll"
      v-for="(action, index) in packageStore.current_preservation_actions
        ?.actions"
    />
  </div>
</template>
