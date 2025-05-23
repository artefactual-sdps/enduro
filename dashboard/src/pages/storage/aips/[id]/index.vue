<script setup lang="ts">
import Tooltip from "bootstrap/js/dist/tooltip";
import { onMounted, ref } from "vue";

import AipLocationCard from "@/components/AipLocationCard.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import UUID from "@/components/UUID.vue";
import WorkflowCollapse from "@/components/WorkflowCollapse.vue";
import WorkflowHelp from "@/components/WorkflowHelp.vue";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";
import IconHelp from "~icons/clarity/help-solid?height=0.8em&width=0.8em";

const aipStore = useAipStore();
const authStore = useAuthStore();

const el = ref<HTMLElement | null>(null);
let tooltip: Tooltip | null = null;

const showHelp = ref(false);
const toggleHelp = () => {
  showHelp.value = !showHelp.value;
  if (tooltip) tooltip.hide();
};

onMounted(() => {
  if (el.value) tooltip = new Tooltip(el.value);
});
</script>

<template>
  <div v-if="aipStore.current">
    <div class="row">
      <div class="col-md-6">
        <h2>AIP details</h2>
        <dl>
          <dt>Name</dt>
          <dd>{{ aipStore.current.name }}</dd>
          <dt>UUID</dt>
          <dd><UUID :id="aipStore.current.uuid" /></dd>
          <dt>Status</dt>
          <dd>
            <StatusBadge :status="aipStore.current.status" type="package" />
          </dd>
          <dt>Deposited</dt>
          <dd>{{ $filters.formatDateTime(aipStore.current.createdAt) }}</dd>
        </dl>
      </div>
      <div class="col-md-6">
        <AipLocationCard />
      </div>
    </div>
    <div
      v-if="
        aipStore.currentWorkflows?.workflows?.length &&
        authStore.checkAttributes(['storage:aips:workflows:list'])
      "
    >
      <div>
        <h2 class="mb-0">
          Workflows
          <a
            ref="el"
            id="workflowHelpToggle"
            href="#workflowHelp"
            role="button"
            aria-expanded="false"
            aria-controls="workflowHelp"
            data-bs-toggle="tooltip"
            data-bs-title="Toggle help"
            @click="toggleHelp"
            ><IconHelp alt="help"
          /></a>
        </h2>
      </div>
      <WorkflowHelp
        :show="showHelp"
        @update:show="(value) => (showHelp = value)"
      />

      <hr />

      <div class="accordion mb-2" id="workflows">
        <WorkflowCollapse
          :workflow="workflow"
          :index="index"
          v-for="(workflow, index) in aipStore.currentWorkflows?.workflows"
          v-bind:key="workflow.uuid"
        />
      </div>
    </div>
  </div>
</template>
