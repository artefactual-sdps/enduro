<script setup lang="ts">
import { api } from "@/client";
import PackageDetailsCard from "@/components/PackageDetailsCard.vue";
import PackageLocationCard from "@/components/PackageLocationCard.vue";
import PreservationActionCollapse from "@/components/PreservationActionCollapse.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import UUID from "@/components/UUID.vue";
import { useAuthStore } from "@/stores/auth";
import { usePackageStore } from "@/stores/package";
import { computed } from "vue";
import IconLink from "~icons/bi/box-arrow-up-right";
import IconHelp from "~icons/clarity/help-solid?height=0.8em&width=0.8em";

const authStore = useAuthStore();
const packageStore = usePackageStore();

const createAipWorkflow = computed(
  () =>
    packageStore.current_preservation_actions?.actions?.filter(
      (action) =>
        action.type === api.EnduroPackagePreservationActionTypeEnum.CreateAip ||
        action.type ===
          api.EnduroPackagePreservationActionTypeEnum.CreateAndReviewAip,
    )[0],
);
</script>

<template>
  <div v-if="packageStore.current">
    <div class="row">
      <div class="col-md-6">
        <h2>SIP details</h2>
        <dl>
          <dt>Name</dt>
          <dd>{{ packageStore.current.name }}</dd>
          <dt>UUID</dt>
          <dd><UUID :id="packageStore.current.aipId" /></dd>
          <dt>Workflow status</dt>
          <dd>
            <StatusBadge
              v-if="createAipWorkflow"
              :status="createAipWorkflow.status"
              :note="
                $filters.getPreservationActionLabel(createAipWorkflow.type)
              "
            />
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
      <div class="col-md-6"></div>
    </div>

    <div v-if="authStore.checkAttributes(['package:listActions'])">
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
              one or more <strong>tasks</strong> performed on a package to
              support preservation.
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
          v-for="(action, index) in packageStore.current_preservation_actions
            ?.actions"
        />
      </div>
    </div>
  </div>
</template>
