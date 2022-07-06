<script setup lang="ts">
import PackageDetailsCard from "@/components/PackageDetailsCard.vue";
import PackageLocationCard from "@/components/PackageLocationCard.vue";
import PackageReviewAlert from "@/components/PackageReviewAlert.vue";
import PackageStatusBadge from "@/components/PackageStatusBadge.vue";
import { usePackageStore } from "@/stores/package";
import "bootstrap/js/dist/collapse";

const packageStore = usePackageStore();
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
            <PackageStatusBadge
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
      <h2 class="flex-grow-1 mb-0">Preservation actions</h2>
      <button
        class="btn btn-sm btn-link text-decoration-none align-self-end"
        type="button"
        data-bs-toggle="collapse"
        data-bs-target="#preservation-actions-table"
        aria-expanded="false"
        aria-controls="preservation-actions-table"
        v-if="packageStore.current_preservation_actions?.actions"
      >
        Expand all | Collapse all
      </button>
    </div>

    <hr />

    <div class="mb-3">
      <h3>
        Create and Review AIP
        <PackageStatusBadge :status="packageStore.current.status" />
      </h3>
      <span v-if="packageStore.current.completedAt">
        Completed
        {{ $filters.formatDateTime(packageStore.current.completedAt) }}
        (took
        {{
          $filters.formatDuration(
            packageStore.current.startedAt,
            packageStore.current.completedAt
          )
        }})
      </span>
      <span v-else>
        Started {{ $filters.formatDateTime(packageStore.current.startedAt) }}
      </span>
    </div>

    <PackageReviewAlert />

    <table
      class="table table-bordered table-sm collapse"
      id="preservation-actions-table"
      v-if="packageStore.current_preservation_actions?.actions"
    >
      <thead>
        <tr>
          <th scope="col">Task #</th>
          <th scope="col">Name</th>
          <th scope="col">Outcome</th>
          <th scope="col">Notes</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(action, idx) in packageStore.current_preservation_actions
            .actions"
          :key="action.id"
        >
          <td>{{ idx + 1 }}</td>
          <td>{{ action.name }}</td>
          <td>
            <span
              class="badge"
              :class="$filters.formatPreservationActionStatus(action.status)"
              >{{ action.status }}</span
            >
          </td>
          <td>TODO: note goes here</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
