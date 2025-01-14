<script setup lang="ts">
import { api } from "@/client";
import PackageDetailsCard from "@/components/PackageDetailsCard.vue";
import PackageLocationCard from "@/components/PackageLocationCard.vue";
import PreservationActionCollapse from "@/components/PreservationActionCollapse.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import UUID from "@/components/UUID.vue";
import { useAuthStore } from "@/stores/auth";
import { usePackageStore } from "@/stores/package";
import { computed, ref } from "vue";

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
        <h2>AIP creation details</h2>
        <dl>
          <dt>Name</dt>
          <dd>{{ packageStore.current.name }}</dd>
          <dt>AIP UUID</dt>
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
      <div class="col-md-6">
        <PackageLocationCard />
        <PackageDetailsCard />
      </div>
    </div>

    <div v-if="authStore.checkAttributes(['package:listActions'])">
      <div class="d-flex">
        <h2 class="mb-0">Preservation actions</h2>
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
