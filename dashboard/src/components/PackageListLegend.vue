<script setup lang="ts">
import { api } from "@/client";
import StatusBadge from "@/components/StatusBadge.vue";
import { computed } from "vue";

const { modelValue } = defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
}>();

const dismiss = () => {
  emit("update:modelValue", false);
};

const show = computed(() => modelValue);

const items = [
  {
    status: api.PackageShowResponseBodyStatusEnum.Done,
    description: "The current workflow or task has completed without errors.",
  },
  {
    status: api.PackageShowResponseBodyStatusEnum.Error,
    description:
      "The current workflow has encountered an error it could not resolve and failed.",
  },
  {
    status: api.PackageShowResponseBodyStatusEnum.InProgress,
    description: "The current workflow is still processing.",
  },
  {
    status: api.PackageShowResponseBodyStatusEnum.Queued,
    description:
      "The current workflow is waiting for an available worker to begin.",
  },
  {
    status: api.PackageShowResponseBodyStatusEnum.Pending,
    description: "The current workflow is awaiting a user decision.",
  },
];
</script>

<template>
  <Transition>
    <div class="alert alert-secondary alert-dismissible" v-if="show">
      <div class="container-fluid">
        <div class="row" v-for="(item, index) in items">
          <div class="col-12 col-md-2 py-2">
            <StatusBadge
              :status="item.status"
              :aria-describedby="`badge-${index}-desc`"
            />
          </div>
          <div class="col-12 col-md-10 py-2" :id="`badge-${index}-desc`">
            {{ item.description }}
          </div>
        </div>
      </div>

      <button
        type="button"
        class="btn-close"
        @click="dismiss"
        aria-label="Close"
      ></button>
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
