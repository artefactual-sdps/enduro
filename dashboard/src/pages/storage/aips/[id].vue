<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import { useRoute, useRouter } from "vue-router/auto";

import AipPendingAlert from "@/components/AipPendingAlert.vue";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import Tabs from "@/components/Tabs.vue";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";
import IconAIPs from "~icons/clarity/bundle-line";
import IconDetails from "~icons/clarity/details-line?font-size=20px";

const route = useRoute("/storage/aips/[id]");
const router = useRouter();
const authStore = useAuthStore();
const aipStore = useAipStore();

const { execute, error } = useAsyncState(
  aipStore.fetchCurrent(route.params.id.toString()).catch((err) => {
    error.value = err;
  }),
  null,
);

const tabs = [
  {
    icon: IconDetails,
    text: "Summary",
    route: router.resolve({
      name: "/storage/aips/[id]/",
      params: { id: route.params.id },
    }),
    show: authStore.checkAttributes(["storage:aips:read"]),
  },
];
</script>

<template>
  <div class="container-xxl">
    <PageLoadingAlert v-if="error" :execute="execute" :error="error" />

    <template v-if="aipStore.current">
      <AipPendingAlert />

      <h1 class="d-flex mb-3">
        <IconAIPs class="me-3 text-dark" />{{ aipStore.current.name }}
      </h1>

      <Tabs :tabs="tabs" param="id" />

      <RouterView />
    </template>
  </div>
</template>
