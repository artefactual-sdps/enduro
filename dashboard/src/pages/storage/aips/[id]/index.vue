<script setup lang="ts">
import AipLocationCard from "@/components/AipLocationCard.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import UUID from "@/components/UUID.vue";
import WorkflowCollapse from "@/components/WorkflowCollapse.vue";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";
import IconLink from "~icons/bi/box-arrow-up-right";
import IconHelp from "~icons/clarity/help-solid?height=0.8em&width=0.8em";

const aipStore = useAipStore();
const authStore = useAuthStore();
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
          <dd><StatusBadge :status="aipStore.current.status" /></dd>
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
      <div class="d-flex">
        <h2 class="mb-0">
          Workflows
          <a
            id="workflowHelpToggle"
            data-bs-toggle="collapse"
            href="#workflowHelp"
            role="button"
            aria-expanded="false"
            aria-controls="workflowHelp"
            aria-label="Show workflows help"
            ><IconHelp alt="help"
          /></a>
        </h2>
      </div>
      <div
        class="collapse"
        id="workflowHelp"
        aria-labelledby="workflowHelpToggle"
      >
        <div class="card card-body flex flex-column bg-light">
          <div>
            <p>
              A <strong>workflow</strong> is composed of one or more
              <strong>tasks</strong> performed on a SIP/AIP to support
              preservation.
            </p>
            <p>
              Click on a workflow listed below to expand it and see more
              information on individual tasks run as part of the workflow.
            </p>
          </div>
          <div class="align-self-end">
            <a
              href="https://github.com/artefactual-sdps/enduro/blob/main/docs/src/user-manual/usage.md#view-tasks-in-enduro"
              target="_new"
              >Learn more <IconLink alt="" aria-hidden="true"
            /></a>
          </div>
        </div>
      </div>

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
