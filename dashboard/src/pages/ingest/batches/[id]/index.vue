<script setup lang="ts">
import Tooltip from "bootstrap/js/dist/tooltip";
import { onMounted, ref } from "vue";

import { api } from "@/client";
import BatchReviewAlert from "@/components/BatchReviewAlert.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import StatusLegend from "@/components/StatusLegend.vue";
import UUID from "@/components/UUID.vue";
import uploader from "@/composables/uploader";
import { useAuthStore } from "@/stores/auth";
import { useBatchStore } from "@/stores/batch";
import IconInfo from "~icons/akar-icons/info-fill";

const authStore = useAuthStore();
const batchStore = useBatchStore();

const el = ref<HTMLElement | null>(null);
let tooltip: Tooltip | null = null;

const showLegend = ref(false);
const toggleLegend = () => {
  showLegend.value = !showLegend.value;
  tooltip?.hide();
};

const statuses = [
  {
    status: api.EnduroIngestSipStatusEnum.Ingested,
    description: "The SIP has successfully completed all ingest processing.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Failed,
    description:
      "The SIP has failed to meet the policy-defined criteria for ingest, halting the workflow.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Processing,
    description:
      "The SIP is currently part of an active workflow and is undergoing processing.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Pending,
    description: "The SIP is part of a workflow awaiting a user decision.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Queued,
    description:
      "The SIP is about to be part of an active workflow and is awaiting processing.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Error,
    description:
      "The SIP workflow encountered a system error and ingest was aborted.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Canceled,
    description:
      "The SIP workflow has been canceled as part of a failed batch.",
  },
  {
    status: api.EnduroIngestSipStatusEnum.Validated,
    description:
      "The SIP has passed validation and it is waiting for other SIPs in the batch to validate.",
  },
];

onMounted(() => {
  if (el.value) tooltip = new Tooltip(el.value);
});
</script>

<template>
  <div v-if="batchStore.current">
    <h2>Batch details</h2>
    <div class="row">
      <div class="col-md-6">
        <dl>
          <dt>Identifier</dt>
          <dd>{{ batchStore.current.identifier }}</dd>
          <dt>UUID</dt>
          <dd><UUID :id="batchStore.current.uuid" /></dd>
          <dt>SIPs</dt>
          <dd>{{ batchStore.current.sipsCount }}</dd>
          <dt>Status</dt>
          <dd>
            <StatusBadge :status="batchStore.current.status" type="package" />
          </dd>
        </dl>
      </div>
      <div class="col-md-6">
        <dl>
          <dt>Ingested by</dt>
          <dd>{{ uploader(batchStore.current) }}</dd>
          <template v-if="batchStore.current.startedAt">
            <dt>Started</dt>
            <dd>{{ $filters.formatDateTime(batchStore.current.startedAt) }}</dd>
          </template>
          <template v-if="batchStore.current.completedAt">
            <dt>Completed</dt>
            <dd>
              {{ $filters.formatDateTime(batchStore.current.completedAt) }}
              <div class="pt-2">
                (took
                {{
                  $filters.formatDuration(
                    batchStore.current.startedAt,
                    batchStore.current.completedAt,
                  )
                }})
              </div>
            </dd>
          </template>
        </dl>
      </div>
    </div>

    <BatchReviewAlert />

    <h2 class="mb-3">SIPs in Batch</h2>
    <StatusLegend
      :show="showLegend"
      :items="statuses"
      @update:show="(val) => (showLegend = val)"
    />
    <div class="table-responsive mb-3">
      <table class="table table-bordered mb-0">
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">Started</th>
            <th scope="col">
              <span class="d-flex gap-2">
                Status
                <button
                  ref="el"
                  class="btn btn-sm btn-link text-decoration-none ms-auto p-0"
                  type="button"
                  data-bs-toggle="tooltip"
                  data-bs-title="Toggle legend"
                  @click="toggleLegend"
                >
                  <IconInfo class="fs-6" aria-hidden="true" />
                  <span class="visually-hidden">Toggle SIP status legend</span>
                </button>
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="sip in batchStore.currentSips" :key="sip.uuid">
            <td>
              <RouterLink
                v-if="authStore.checkAttributes(['ingest:sips:read'])"
                :to="{ name: '/ingest/sips/[id]/', params: { id: sip.uuid } }"
              >
                {{ sip.name || sip.uuid }}
              </RouterLink>
              <span v-else>{{ sip.name || sip.uuid }}</span>
            </td>
            <td>{{ $filters.formatDateTime(sip.startedAt) }}</td>
            <td>
              <StatusBadge :status="sip.status" type="package" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
