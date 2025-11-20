<script setup lang="ts">
import { computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router/auto";

import SIPUploadLocal from "@/components/SIPUploadLocal.vue";
import SIPUploadSource from "@/components/SIPUploadSource.vue";
import Tabs from "@/components/Tabs.vue";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import IconUpload from "~icons/clarity/backup-restore-line?font-size=20px";
import IconAdd from "~icons/clarity/plus-circle-line";
import IconCloudUpload from "~icons/clarity/upload-cloud-line?font-size=20px";

import "@uppy/core/dist/style.css";
import "@uppy/dashboard/dist/style.css";
import "@uppy/progress-bar/dist/style.css";

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const layoutStore = useLayoutStore();

layoutStore.updateBreadcrumb([{ text: "Ingest" }, { text: "Upload SIPs" }]);

watch(
  () => route.query.from,
  (from) => {
    if (from === undefined || from === "") {
      const from = authStore.checkAttributes(["ingest:sips:upload"])
        ? "local"
        : "source";
      router.replace({ path: "/ingest/upload", query: { from } });
    }
  },
  { immediate: true },
);

const tabs = computed(() => [
  {
    icon: IconUpload,
    text: "Local upload",
    route: router.resolve({
      name: "/ingest/upload/",
      query: { ...route.query, from: "local" },
    }),
    show: authStore.checkAttributes(["ingest:sips:upload"]),
  },
  {
    icon: IconCloudUpload,
    text: "Select from source",
    route: router.resolve({
      name: "/ingest/upload/",
      query: { ...route.query, from: "source" },
    }),
    show: authStore.checkAttributes([
      "ingest:sipsources:objects:list",
      "ingest:sips:create",
    ]),
  },
]);
</script>

<template>
  <div class="container-xxl">
    <h1 class="d-flex mb-3"><IconAdd class="me-3 text-dark" />Upload SIPs</h1>

    <div class="mb-3">
      <Tabs :tabs="tabs" param="from" />
    </div>

    <div class="mb-3">
      <SIPUploadLocal
        v-if="
          authStore.checkAttributes(['ingest:sips:upload']) &&
          route.query.from === 'local'
        "
      />
      <SIPUploadSource
        v-if="
          authStore.checkAttributes([
            'ingest:sipsources:objects:list',
            'ingest:sips:create',
          ]) && route.query.from === 'source'
        "
      />
    </div>
  </div>
</template>
