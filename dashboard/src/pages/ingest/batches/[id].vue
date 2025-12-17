<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import { useRoute, useRouter } from "vue-router/auto";

import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import Tabs from "@/components/Tabs.vue";
import { useAuthStore } from "@/stores/auth";
import { useBatchStore } from "@/stores/batch";
import IconDetails from "~icons/clarity/details-line?font-size=20px";
import IconBatches from "~icons/clarity/layers-line";

const route = useRoute("/ingest/batches/[id]");
const router = useRouter();
const authStore = useAuthStore();
const batchStore = useBatchStore();

const { execute, error } = useAsyncState(
  batchStore.fetchCurrent(route.params.id.toString()),
  null,
);

const tabs = [
  {
    icon: IconDetails,
    text: "Summary",
    route: router.resolve({
      name: "/ingest/batches/[id]/",
      params: { id: route.params.id },
    }),
    show: authStore.checkAttributes(["ingest:batches:read"]),
  },
];
</script>

<template>
  <div class="container-xxl">
    <PageLoadingAlert v-if="error" :execute="execute" :error="error" />

    <template v-if="batchStore.current">
      <h1 class="d-flex mb-3">
        <IconBatches class="me-3 text-dark" />{{
          batchStore.current.identifier
        }}
      </h1>

      <Tabs :tabs="tabs" param="id" />

      <RouterView />
    </template>
  </div>
</template>
