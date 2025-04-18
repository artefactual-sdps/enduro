<script setup lang="ts">
import { computed } from "vue";

import { api } from "@/client";
import StatusBadge from "@/components/StatusBadge.vue";

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
    status: api.EnduroIngestSipStatusEnum.Error,
    description:
      "The SIP workflow encountered a system error and ingest was aborted.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Failed,
    description:
      "The SIP has failed to failed to meet the policy-defined criteria for ingest, halting the workflow.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Queued,
    description:
      "The SIP is about to be part of an active workflow and is awaiting processing.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Processing,
    description:
      "The SIP is currently part of an active workflow and is undergoing processing.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Pending,
    description: "The SIP is part of a workflow awaiting a user decision.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Ingested,
    description: "The SIP has successfully completed all ingest processing.",
  },
];
</script>

<template>
  <Transition>
    <div class="alert alert-secondary alert-dismissible" v-if="show">
      <div class="container-fluid">
        <div class="row" v-for="(item, index) in items" :key="item.status">
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
