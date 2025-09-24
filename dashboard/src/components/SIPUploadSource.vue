<script setup lang="ts">
import { useAsyncState, useInfiniteScroll } from "@vueuse/core";
import { ref } from "vue";
import { useRouter } from "vue-router/auto";

import { api, client } from "@/client";
import { humanFileSize } from "@/composables/format";
import IconBundle from "~icons/clarity/bundle-line";

const router = useRouter();

// Hardcoded source ID that must match the one in the backend configuration.
// TODO: Fetch SIP sources from the backend.
const sourceId = "e6ddb29a-66d1-480e-82eb-fcfef1c825c5";
const items = ref<api.EnduroIngestSipsourceObject[]>([]);
const selectedSips = ref<string[]>([]);
const nextCursor = ref<string | undefined>(undefined);
const listContainer = ref<HTMLElement>();
const errorMessage = ref<string | null>(null);

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
    // Group request promises.
    const ingestPromises = selectedSips.value.map((key) => {
      return client.ingest.ingestAddSip({ key, sourceId });
    });
    await Promise.all(ingestPromises);
    router.push({ path: "/ingest/sips" });
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
      @click="errorMessage = null"
      aria-label="Close"
    ></button>
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
    <div class="table-responsive overflow-auto" style="max-height: 500px">
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
            v-else
            v-for="item in items"
            :key="item.key"
            :class="selectedSips.includes(item.key) ? 'table-primary' : ''"
            role="button"
            @click="clickSip(item.key)"
          >
            <td>
              <input
                class="form-check-input"
                type="checkbox"
                v-model="selectedSips"
                :value="item.key"
                :id="'cb-' + item.key"
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
              ></div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="form-text text-end">
      Each SIP uploaded will be ingested separately in its own workflow.
    </div>
  </div>

  <h2 class="mb-3">2. Launch Ingest</h2>
  <button
    class="btn btn-primary"
    @click="startIngest"
    :disabled="selectedSips.length === 0"
  >
    Start Ingest
  </button>
</template>

<style scoped>
/* Remove top border from selected items when the previous item is also selected. */
.list-group-item.border-primary + .list-group-item.border-primary {
  border-top: none !important;
}
</style>
