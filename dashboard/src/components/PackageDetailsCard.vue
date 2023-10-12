<script setup lang="ts">
import { storageServiceDownloadURL } from "@/client";
import StatusBadge from "@/components/StatusBadge.vue";
import { usePackageStore } from "@/stores/package";
import { computed, watch } from "vue";

const packageStore = usePackageStore();

const download = () => {
  if (!packageStore.current?.aipId) return;
  const url = storageServiceDownloadURL(packageStore.current.aipId);
  window.open(url, "_blank");
};

const stored = computed(() => {
  return packageStore.current?.aipId?.length;
});

watch(packageStore.ui.download, () => download());
</script>

<template>
  <div class="card mb-3" v-if="packageStore.current">
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
            v-if="packageStore.current_preservation_actions?.actions"
            :status="
              packageStore.current_preservation_actions?.actions[0].status
            "
            :note="
              $filters.getPreservationActionLabel(
                packageStore.current_preservation_actions?.actions[0].type,
              )
            "
          />
        </dd>
      </dl>
      <div class="d-flex flex-wrap gap-2">
        <button class="btn btn-secondary btn-sm disabled">
          View metadata summary
        </button>
        <button
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
