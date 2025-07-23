<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import { onMounted, ref } from "vue";

import { client } from "@/client";
import IconBundle from "~icons/clarity/bundle-line";

interface SourceItem {
  key: string;
  selected: boolean;
  size?: number;
}

const items = ref<SourceItem[]>([]);
const checkedItems = ref<string[]>([]);

const { execute } = useAsyncState(() => {
  return client.ingest
    .ingestListSourceItems({ uuid: "e6ddb29a-66d1-480e-82eb-fcfef1c825c5" })
    .then((page) => {
      items.value = page.items.map((item) => ({
        key: item.key,
        selected: false,
        size: item.size,
      }));
    })
    .catch((error) => {
      console.error("Failed to fetch SIP list:", error);
      items.value = [];
    });
}, null);

onMounted(() => {
  execute();
});
</script>

<template>
  <h2 class="mb-3">1. Select SIPs to Ingest</h2>
  <div class="d-flex gap-2">
    <div class="form-group">
      <label for="source-select"
        >Selected SIPs: {{ checkedItems.length }}</label
      >
      <ul
        class="list-group border border-primary p-0 overflow-auto"
        multiple
        id="source-select"
        style="max-height: 500px; min-width: 300px"
      >
        <li
          v-for="item in items"
          :key="item.key"
          class="list-group-item"
          :class="
            checkedItems.includes(item.key) ? 'list-group-item-primary' : ''
          "
        >
          <input
            class="form-check-input me-1"
            type="checkbox"
            v-model="checkedItems"
            :value="item.key"
            :id="'cb-' + item.key"
          />
          <label
            class="form-check-label stretched-link"
            :for="'cb-' + item.key"
          >
            <IconBundle /> {{ item.key }}
          </label>
        </li>
      </ul>
      <note class="text-muted">
        SIPS will each be ingested individually in their own workflow.
      </note>
    </div>
  </div>

  <h2 class="mt-3">2. Launch Ingest</h2>
  <button class="btn btn-primary mt-2">Start Ingest</button>
</template>
