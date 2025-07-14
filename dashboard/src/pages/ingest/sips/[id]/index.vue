<script setup lang="ts">
import Tooltip from "bootstrap/js/dist/tooltip";
import { computed, onMounted, ref } from "vue";

import SipRelatedPackages from "@/components/SipRelatedPackages.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import UUID from "@/components/UUID.vue";
import WorkflowCollapse from "@/components/WorkflowCollapse.vue";
import WorkflowHelp from "@/components/WorkflowHelp.vue";
import uploader from "@/composables/sipUploader";
import { useAuthStore } from "@/stores/auth";
import { useSipStore } from "@/stores/sip";
import IconHelp from "~icons/clarity/help-solid?height=0.8em&width=0.8em";

const authStore = useAuthStore();
const sipStore = useSipStore();

const el = ref<HTMLElement | null>(null);
const showHelp = ref(false);

const toggleHelp = () => {
  showHelp.value = !showHelp.value;
  if (tooltip) tooltip.hide();
};

const createAipWorkflow = computed(
  () => sipStore.currentWorkflows?.workflows?.[0],
);

let tooltip: Tooltip | null = null;
onMounted(() => {
  if (el.value) tooltip = new Tooltip(el.value);
});
</script>

<template>
  <div v-if="sipStore.current">
    <div class="row">
      <div class="col-md-6">
        <h2>SIP details</h2>
        <dl>
          <dt>Name</dt>
          <dd>{{ sipStore.current.name }}</dd>
          <dt>UUID</dt>
          <dd><UUID :id="sipStore.current.uuid" /></dd>
          <dt>Status</dt>
          <dd>
            <StatusBadge :status="sipStore.current.status" type="package" />
          </dd>
          <dt>Uploaded by</dt>
          <dd>
            {{ uploader(sipStore.current) }}
          </dd>
          <dt>Started</dt>
          <dd>{{ $filters.formatDateTime(createAipWorkflow?.startedAt) }}</dd>
          <dt v-if="createAipWorkflow?.completedAt">Completed</dt>
          <dd v-if="createAipWorkflow?.completedAt">
            {{ $filters.formatDateTime(createAipWorkflow.completedAt) }}
            <div class="pt-2">
              (took
              {{
                $filters.formatDuration(
                  createAipWorkflow.startedAt,
                  createAipWorkflow.completedAt,
                )
              }})
            </div>
          </dd>
        </dl>
      </div>
      <div class="col-md-6">
        <SipRelatedPackages />
      </div>
    </div>

    <div
      v-if="
        sipStore.currentWorkflows?.workflows?.length &&
        authStore.checkAttributes(['ingest:sips:workflows:list'])
      "
    >
      <div>
        <h2 class="mb-0">
          Ingest workflow details
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
          v-for="(workflow, index) in sipStore.currentWorkflows?.workflows"
          v-bind:key="workflow.uuid"
        />
      </div>
    </div>
  </div>
</template>
