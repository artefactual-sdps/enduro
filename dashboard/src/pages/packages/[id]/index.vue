<script setup lang="ts">
import { storageServiceDownloadURL } from "../../../client";
import PackageReviewAlert from "../../../components/PackageReviewAlert.vue";
import PackageStatusBadge from "../../../components/PackageStatusBadge.vue";
import { usePackageStore } from "../../../stores/package";
import "bootstrap/js/dist/collapse";

const packageStore = usePackageStore();

const download = () => {
  if (!packageStore.current?.aipId) return;
  const url = storageServiceDownloadURL(packageStore.current.aipId);
  window.open(url, "_blank");
};
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
          <dd>{{ packageStore.current.startedAt }}</dd>
        </dl>
      </div>
      <div class="col-md-6">
        <div class="card mb-3">
          <div class="card-body">
            <h5 class="card-title">Location</h5>
            <p class="card-text">
              <a href="#">aip-review</a>
            </p>
            <div class="">
              <a href="#" class="btn btn-primary btn-sm"
                >Choose storage location</a
              >
            </div>
          </div>
        </div>
        <div class="card mb-3">
          <div class="card-body">
            <h5 class="card-title">Package details</h5>
            <dl>
              <dt>Original objects</dt>
              <dd>14</dd>
              <dt>Package size</dt>
              <dd>1.45 GB</dd>
              <dt>Last workflow outcome</dt>
              <dd>
                <PackageStatusBadge
                  :status="packageStore.current.status"
                  :note="'Create and Review AIP'"
                />
              </dd>
            </dl>
            <div class="d-flex flex-wrap gap-2">
              <button class="btn btn-secondary btn-sm disabled">
                View metadata summary
              </button>
              <button
                class="btn btn-primary btn-sm"
                type="button"
                @click="download"
              >
                Download
              </button>
            </div>
          </div>
        </div>
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
      v-if="
        packageStore.current_preservation_actions &&
        packageStore.current_preservation_actions.actions
      "
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

<style scoped></style>
