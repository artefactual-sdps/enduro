<script setup lang="ts">
import { computed } from "vue";

import StatusBadge from "@/components/StatusBadge.vue";
import type { StatusEnum } from "@/components/StatusBadge.vue";

export type LegendItem = {
  status: StatusEnum;
  description: string;
};

const { show = false } = defineProps<{
  show?: boolean;
  items: LegendItem[];
}>();

const emit = defineEmits<{
  (e: "update:show", value: boolean): void;
}>();

const dismiss = () => {
  emit("update:show", false);
};

const visible = computed(() => show);
</script>

<template>
  <Transition>
    <div v-if="visible" class="alert alert-secondary alert-dismissible">
      <div class="container-fluid">
        <div v-for="(item, index) in items" :key="item.status" class="row">
          <div class="col-12 col-md-2 py-2 text-end">
            <StatusBadge
              :status="item.status"
              type="package"
              :aria-describedby="`badge-${index}-desc`"
            />
          </div>
          <div :id="`badge-${index}-desc`" class="col-12 col-md-10 py-2">
            {{ item.description }}
          </div>
        </div>
      </div>

      <button
        type="button"
        class="btn-close"
        aria-label="Close"
        @click="dismiss"
      />
    </div>
  </Transition>
</template>

<style scoped>
.v-enter-active,
.v-leave-active {
  transition: opacity 0.1s ease;
}

.v-enter-from,
.v-leave-to {
  opacity: 0;
}
</style>
