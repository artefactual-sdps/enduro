<script setup lang="ts">
import { useAsyncState, useInfiniteScroll } from "@vueuse/core";
import { ref } from "vue";
import { useRouter } from "vue-router/auto";

import { api, client } from "@/client";
import IconBundle from "~icons/clarity/bundle-line";

const router = useRouter();

// Hardcoded source ID that must match the one in the backend configuration.
// TODO: Fetch SIP sources from the backend.
const sourceId = "e6ddb29a-66d1-480e-82eb-fcfef1c825c5";
const items = ref<api.EnduroIngestSipsourceObject[]>([]);
const checkedItems = ref<string[]>([]);
const nextCursor = ref<string | undefined>(undefined);
const listContainer = ref<HTMLElement>();
const errorMessage = ref<string | null>(null);

const { execute, isLoading } = useAsyncState(
  async (cursor: string | undefined) => {
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
  if (!checkedItems.value.length) return;

  // Clear previous error.
  errorMessage.value = null;

  try {
    // Group request promises.
    const ingestPromises = checkedItems.value.map((key) => {
      return client.ingest.ingestAddSip({ key, sourceId });
    });
    await Promise.all(ingestPromises);
    router.push({ path: "/ingest/sips" });
  } catch (error) {
    console.error("Failed to start ingest:", error);
    errorMessage.value = "Failed to start ingest.";
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
  <div class="mb-3">
    <div class="form-text d-flex gap-2 justify-content-end">
      <span>Selected SIPs: {{ checkedItems.length }}</span>
      <span>•</span>
      <a href="#" @click.prevent="checkedItems = items.map((item) => item.key)"
        >Select all</a
      >
      <span>•</span>
      <a href="#" @click.prevent="checkedItems = []">Clear selections</a>
    </div>
    <ul
      ref="listContainer"
      class="list-group list-group-flush overflow-auto border"
      style="max-height: 500px"
    >
      <li v-if="items.length === 0 && !isLoading" class="list-group-item p-0">
        <div class="p-2">No SIPs found</div>
      </li>
      <li
        v-for="item in items"
        :key="item.key"
        class="list-group-item list-group-item-action p-0"
        :class="
          checkedItems.includes(item.key)
            ? 'list-group-item-primary border border-primary'
            : ''
        "
      >
        <label class="form-check-label d-flex gap-2 align-items-center p-2">
          <input
            class="form-check-input mt-0 mx-1"
            type="checkbox"
            v-model="checkedItems"
            :value="item.key"
            :id="'cb-' + item.key"
          />
          <IconBundle aria-hidden="true" />
          {{ item.key }}
        </label>
      </li>
      <li v-if="isLoading" class="list-group-item text-center p-3">
        <div class="spinner-border text-muted" role="status">
          <span class="visually-hidden">Loading...</span>
        </div>
      </li>
    </ul>
    <div class="form-text text-end">
      Each SIP uploaded will be ingested separately in its own workflow.
    </div>
  </div>

  <h2 class="mb-3">2. Launch Ingest</h2>
  <button
    class="btn btn-primary"
    @click="startIngest"
    :disabled="checkedItems.length === 0"
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
