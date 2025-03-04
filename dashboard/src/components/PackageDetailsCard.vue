<script setup lang="ts">
import { computed, watch } from "vue";

import { storageServiceDownloadURL } from "@/client";
import StatusBadge from "@/components/StatusBadge.vue";
import { useAuthStore } from "@/stores/auth";
import { useIngestStore } from "@/stores/ingest";

const authStore = useAuthStore();
const ingestStore = useIngestStore();

const download = () => {
  if (!ingestStore.currentSip?.aipId) return;
  const url = storageServiceDownloadURL(ingestStore.currentSip.aipId);
  window.open(url, "_blank");
};

const stored = computed(() => {
  return ingestStore.currentSip?.aipId?.length;
});

watch(ingestStore.ui.download, () => download());
</script>

<template>
  <div class="card mb-3" v-if="ingestStore.currentSip">
    <div class="card-body">
      <h4 class="card-title">Package details</h4>
      <dl>
        <dt>Original objects</dt>
        <dd>N/A</dd>
        <dt>Package size</dt>
        <dd>N/A</dd>
        <dt>Last workflow status</dt>
        <dd>
          <StatusBadge
            v-if="ingestStore.currentPreservationActions?.actions"
            :status="ingestStore.currentPreservationActions?.actions[0].status"
            :note="
              $filters.getPreservationActionLabel(
                ingestStore.currentPreservationActions?.actions[0].type,
              )
            "
          />
        </dd>
      </dl>
      <div class="d-flex flex-wrap gap-2">
        <button
          v-if="authStore.checkAttributes(['storage:aips:download'])"
          :class="{
            btn: true,
            'btn-primary': true,
            'btn-sm': true,
            disabled: !stored,
          }"
          type="button"
          @click="download"
        >
          Download
        </button>
      </div>
    </div>
  </div>
</template>
