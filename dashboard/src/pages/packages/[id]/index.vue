<script setup lang="ts">
import { api } from "@/client";
import PackageDetailsCard from "@/components/PackageDetailsCard.vue";
import PackageLocationCard from "@/components/PackageLocationCard.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import UUID from "@/components/UUID.vue";
import { usePackageStore } from "@/stores/package";

const packageStore = usePackageStore();

const createAipWorkflow = $computed(
  () =>
    packageStore.current_preservation_actions?.actions?.filter(
      (action) =>
        action.type ===
        api.EnduroPackagePreservationActionResponseBodyTypeEnum.CreateAip
    )[0]
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
                  createAipWorkflow.completedAt
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
  </div>
</template>
