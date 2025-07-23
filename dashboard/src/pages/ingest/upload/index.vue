<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router/auto";

import SIPUploadCloud from "@/components/SIPUploadCloud.vue";
import SIPUploadLocal from "@/components/SIPUploadLocal.vue";
import Tabs from "@/components/Tabs.vue";
import { useLayoutStore } from "@/stores/layout";
import IconUpload from "~icons/clarity/backup-restore-line?raw&font-size=20px";
import IconAdd from "~icons/clarity/plus-circle-line";
import IconCloudUpload from "~icons/clarity/upload-cloud-line?raw&font-size=20px";

import "@uppy/core/dist/style.css";
import "@uppy/dashboard/dist/style.css";
import "@uppy/progress-bar/dist/style.css";

const route = useRoute();
const router = useRouter();
const layoutStore = useLayoutStore();

layoutStore.updateBreadcrumb([{ text: "Ingest" }, { text: "Upload SIPs" }]);

onMounted(() => {
  if (route.query.source === undefined || route.query.source === "") {
    router.replace({ path: "/ingest/upload", query: { source: "local" } });
  }
});

const tabs = computed(() => [
  {
    icon: IconUpload,
    text: "Local upload",
    route: router.resolve({
      name: "/ingest/upload/",
      query: { ...route.query, source: "local" },
    }),
    show: true,
  },
  {
    icon: IconCloudUpload,
    text: "Cloud upload",
    route: router.resolve({
      name: "/ingest/upload/",
      query: { ...route.query, source: "cloud" },
    }),
    show: true,
  },
]);
</script>

<template>
  <div class="container-xxl">
    <h1 class="d-flex mb-3"><IconAdd class="me-3 text-dark" />Upload SIPs</h1>

    <div class="mb-3">
      <Tabs :tabs="tabs" param="source" />
    </div>

    <div class="mb-3">
      <SIPUploadLocal v-if="route.query.source === 'local'" />
      <SIPUploadCloud v-if="route.query.source === 'cloud'" />
    </div>
  </div>
</template>
