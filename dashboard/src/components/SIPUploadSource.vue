<script setup lang="ts">
import { useAsyncState, useInfiniteScroll } from "@vueuse/core";
import { ref } from "vue";
import { useRouter } from "vue-router/auto";

import { api, client } from "@/client";
import { humanFileSize } from "@/composables/format";
import { useAuthStore } from "@/stores/auth";
import IconBundle from "~icons/clarity/bundle-line";

const authStore = useAuthStore();
const router = useRouter();

// Hardcoded source ID that must match the one in the backend configuration.
// TODO: Fetch SIP sources from the backend.
const sourceId = "e6ddb29a-66d1-480e-82eb-fcfef1c825c5";
const items = ref<api.EnduroIngestSipsourceObject[]>([]);
const selectedSips = ref<string[]>([]);
const nextCursor = ref<string | undefined>(undefined);
const listContainer = ref<HTMLElement>();
const errorMessage = ref<string | null>(null);

// At this point, the user must have at least one of the permissions
// to create SIP or batches. Determine whether to default to batch ingest.
// The switch will be disabled if the user doesn't have both permissions.
const isBatch = ref(!authStore.checkAttributes(["ingest:sips:create"]));
const batchID = ref("");

const { execute, isLoading } = useAsyncState(
  async (cursor?: string | undefined) => {
    // Clear previous error.
    errorMessage.value = null;
    try {
      const page = await client.ingest.ingestListSipSourceObjects({
        uuid: sourceId,
        cursor,
      });
      items.value = [...items.value, ...page.objects];
      nextCursor.value = page.next;
    } catch (error) {
      console.error("Failed to load SIPs:", error);
      errorMessage.value = "Failed to load SIPs.";
      items.value = [];
    }
  },
  null,
);

useInfiniteScroll(
  listContainer,
  async () => {
    await execute(0, nextCursor.value);
  },
  {
    distance: 100,
    canLoadMore: () => !!nextCursor.value && !isLoading.value,
  },
);

const startIngest = async () => {
  if (!selectedSips.value.length) return;

  // Clear previous error.
  errorMessage.value = null;

  try {
    if (isBatch.value) {
      const addBatchRequestBody: api.AddBatchRequestBody = {
        keys: selectedSips.value,
        sourceId,
      };
      const identifier = batchID.value.trim();
      if (identifier) addBatchRequestBody.identifier = identifier;
      const { uuid } = await client.ingest.ingestAddBatch({
        addBatchRequestBody,
      });
      router.push({ name: "/ingest/batches/[id]/", params: { id: uuid } });
    } else {
      // Group request promises.
      const ingestPromises = selectedSips.value.map((key) => {
        return client.ingest.ingestAddSip({ key, sourceId });
      });
      await Promise.all(ingestPromises);
      router.push({ path: "/ingest/sips" });
    }
  } catch (error) {
    console.error("Failed to start ingest:", error);
    errorMessage.value = "Failed to start ingest.";
  }
};

const clickSip = (key: string) => {
  if (selectedSips.value.includes(key)) {
    selectedSips.value = selectedSips.value.filter((i) => i !== key);
  } else {
    selectedSips.value.push(key);
  }
};
</script>

<template>
  <div
    v-if="errorMessage"
    class="alert alert-danger alert-dismissible mb-3"
    role="alert"
  >
    {{ errorMessage }}
    <button
      type="button"
      class="btn-close"
      aria-label="Close"
      @click="errorMessage = null"
    />
  </div>

  <h2 class="mb-3">1. Select SIPs to Ingest</h2>
  <div class="mb-3 table">
    <div class="form-text d-flex gap-2 justify-content-end">
      <span>Selected SIPs: {{ selectedSips.length }}</span>
      <span aria-hidden="true">•</span>
      <a href="#" @click.prevent="selectedSips = items.map((item) => item.key)"
        >Select all</a
      >
      <span aria-hidden="true">•</span>
      <a href="#" @click.prevent="selectedSips = []">Clear selections</a>
    </div>
    <div class="table-responsive overflow-auto">
      <table ref="listContainer" class="table table-hover mb-1">
        <thead>
          <tr class="sticky-top">
            <th scope="col">&nbsp;</th>
            <th scope="col">Name</th>
            <th scope="col" class="d-none d-sm-table-cell">Size</th>
            <th scope="col" class="d-none d-sm-table-cell">Deposited</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="items.length === 0 && !isLoading">
            <td colspan="4">No SIPs found</td>
          </tr>
          <tr
            v-for="item in items"
            v-else
            :key="item.key"
            :class="selectedSips.includes(item.key) ? 'table-primary' : ''"
            role="button"
            @click="clickSip(item.key)"
          >
            <td>
              <input
                :id="'cb-' + item.key"
                v-model="selectedSips"
                class="form-check-input"
                type="checkbox"
                :value="item.key"
              />
            </td>
            <td>
              <span class="d-none d-sm-inline-block"
                ><IconBundle aria-hidden="true"
              /></span>
              {{ item.key }}
            </td>
            <td class="d-none d-sm-table-cell">
              {{ item.size ? `${humanFileSize(item.size, 1)}` : "" }}
            </td>
            <td class="d-none d-sm-table-cell">
              {{ $filters.formatDateTime(item.modTime) }}
            </td>
          </tr>
          <tr v-if="isLoading">
            <td colspan="4" class="text-center">
              Loading...
              <div
                class="spinner-border spinner-border-sm text-muted"
                role="status"
              />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-if="!isBatch" class="form-text text-end">
      Each SIP uploaded will be ingested separately in its own workflow.
    </div>
  </div>

  <div>
    <h2 class="mb-3">2. Configure Ingest</h2>
    <div class="form-check form-switch mb-3">
      <input
        id="batch-switch"
        v-model="isBatch"
        class="form-check-input"
        type="checkbox"
        role="switch"
        :disabled="
          !authStore.checkAttributes([
            'ingest:sips:create',
            'ingest:batches:create',
          ])
        "
      />
      <label class="form-check-label" for="batch-switch">
        Treat this ingest as a batch?
      </label>
    </div>
    <div v-if="isBatch" class="mb-3">
      <label class="form-label" for="batch-id"
        >Enter custom batch identifier (optional)</label
      >
      <input
        id="batch-id"
        v-model="batchID"
        type="text"
        class="form-control"
        placeholder="Batch identifier"
      />
    </div>
  </div>

  <h2 class="mb-3">3. Launch Ingest</h2>
  <button
    class="btn btn-primary"
    :disabled="selectedSips.length === 0"
    @click="startIngest"
  >
    Start Ingest
  </button>
</template>

<style scoped>
/* Remove top border from selected items when the previous item is also selected. */
.list-group-item.border-primary + .list-group-item.border-primary {
  border-top: none !important;
}

.table-responsive {
  max-height: 500px;
}
</style>
