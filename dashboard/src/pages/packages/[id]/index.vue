<script setup lang="ts">
import PackageDetailsCard from "@/components/PackageDetailsCard.vue";
import PackageLocationCard from "@/components/PackageLocationCard.vue";
import PreservationActionCollapse from "@/components/PreservationActionCollapse.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import { usePackageStore } from "@/stores/package";

const packageStore = usePackageStore();

let toggleAll = $ref<boolean | null>(false);
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
          <dd>{{ packageStore.current.aipId }}</dd>
          <dt>Workflow status</dt>
          <dd>
            <StatusBadge
              :status="packageStore.current.status"
              :note="'Create and Review AIP'"
            />
          </dd>
          <dt>Started</dt>
          <dd>{{ $filters.formatDateTime(packageStore.current.startedAt) }}</dd>
          <span v-if="packageStore.current.completedAt">
            <dt>Completed</dt>
            <dd>
              {{ $filters.formatDateTime(packageStore.current.completedAt) }}
              <div class="pt-2">
                (took
                {{
                  $filters.formatDuration(
                    packageStore.current.startedAt,
                    packageStore.current.completedAt
                  )
                }})
              </div>
            </dd>
          </span>
        </dl>
      </div>
      <div class="col-md-6">
        <PackageLocationCard />
        <PackageDetailsCard />
      </div>
    </div>

    <div class="d-flex">
      <h2 class="mb-0">Preservation actions</h2>
      <div
        class="align-self-end ms-auto d-flex"
        v-if="packageStore.current_preservation_actions?.actions"
      >
        <button
          class="btn btn-sm btn-link text-decoration-none p-0"
          type="button"
          @click="toggleAll = true"
        >
          Expand all
        </button>
        <span class="px-1 link-primary">|</span>
        <button
          class="btn btn-sm btn-link text-decoration-none p-0"
          type="button"
          @click="toggleAll = false"
        >
          Collapse all
        </button>
      </div>
    </div>

    <PreservationActionCollapse
      :action="action"
      :index="index"
      v-model:toggleAll="toggleAll"
      v-for="(action, index) in packageStore.current_preservation_actions
        ?.actions"
    />
  </div>
</template>
