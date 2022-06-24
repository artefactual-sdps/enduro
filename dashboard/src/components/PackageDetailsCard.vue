<script setup lang="ts">
import { storageServiceDownloadURL } from "@/client";
import PackageStatusBadge from "@/components/PackageStatusBadge.vue";
import { usePackageStore } from "@/stores/package";

const packageStore = usePackageStore();

const download = () => {
  if (!packageStore.current?.aipId) return;
  const url = storageServiceDownloadURL(packageStore.current.aipId);
  window.open(url, "_blank");
};
</script>

<template>
  <div class="card mb-3" v-if="packageStore.current">
    <div class="card-body">
      <h5 class="card-title">Package details</h5>
      <dl>
        <dt>Original objects</dt>
        <dd>N/A</dd>
        <dt>Package size</dt>
        <dd>N/A</dd>
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
        <button class="btn btn-primary btn-sm" type="button" @click="download">
          Download
        </button>
      </div>
    </div>
  </div>
</template>
