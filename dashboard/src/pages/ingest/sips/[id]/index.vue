<script setup lang="ts">
import { computed } from "vue";

import { api } from "@/client";
import AipLocationCard from "@/components/AipLocationCard.vue";
import PreservationActionCollapse from "@/components/PreservationActionCollapse.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import { useAuthStore } from "@/stores/auth";
import { useSipStore } from "@/stores/sip";
import IconLink from "~icons/bi/box-arrow-up-right";
import IconHelp from "~icons/clarity/help-solid?height=0.8em&width=0.8em";

const authStore = useAuthStore();
const sipStore = useSipStore();

const createAipWorkflow = computed(
  () =>
    sipStore.currentPreservationActions?.actions?.filter(
      (action) =>
        action.type ===
          api.EnduroIngestSipPreservationActionTypeEnum.CreateAip ||
        action.type ===
          api.EnduroIngestSipPreservationActionTypeEnum.CreateAndReviewAip,
    )[0],
);
</script>

<template>
  <div v-if="sipStore.current">
    <div class="row">
      <div class="col-md-6">
        <h2>SIP details</h2>
        <dl>
          <dt>Name</dt>
          <dd>{{ sipStore.current.name }}</dd>
          <dt>Status</dt>
          <dd><StatusBadge :status="sipStore.current.status" /></dd>
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
        <AipLocationCard />
      </div>
    </div>

    <div v-if="authStore.checkAttributes(['ingest:sips:actions:list'])">
      <div class="d-flex">
        <h2 class="mb-0">
          Preservation actions
          <a
            id="presActionHelpToggle"
            data-bs-toggle="collapse"
            href="#preservationActionHelp"
            role="button"
            aria-expanded="false"
            aria-controls="preservationActionHelp"
            aria-label="Show preservation action help"
            ><IconHelp alt="help"
          /></a>
        </h2>
      </div>
      <div
        class="collapse"
        id="preservationActionHelp"
        aria-labelledby="presActionHelpToggle"
      >
        <div class="card card-body flex flex-column bg-light">
          <div>
            <p>
              A preservation action is a <strong>workflow</strong> composed of
              one or more <strong>tasks</strong> performed on a SIP to support
              preservation.
            </p>
            <p>
              Click on a preservation action listed below to expand it and see
              more information on individual tasks run as part of the workflow.
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

      <div class="accordion mb-2" id="preservation-actions">
        <PreservationActionCollapse
          :action="action"
          :index="index"
          v-for="(action, index) in sipStore.currentPreservationActions
            ?.actions"
          v-bind:key="action.id"
        />
      </div>
    </div>
  </div>
</template>
