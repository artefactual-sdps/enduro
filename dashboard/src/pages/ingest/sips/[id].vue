<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import { useRoute, useRouter } from "vue-router/auto";

import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import SipPendingAlert from "@/components/SipPendingAlert.vue";
import Tabs from "@/components/Tabs.vue";
import { useAuthStore } from "@/stores/auth";
import { useSipStore } from "@/stores/sip";
import IconDetails from "~icons/clarity/details-line?font-size=20px";
import IconSIPs from "~icons/octicon/package-dependencies-24";

const route = useRoute("/ingest/sips/[id]");
const router = useRouter();
const authStore = useAuthStore();
const sipStore = useSipStore();

const { execute, error } = useAsyncState(
  sipStore.fetchCurrent(route.params.id.toString()).then(() => {
    if (
      sipStore.current?.uuid &&
      authStore.checkAttributes(["ingest:sips:workflows:list"])
    ) {
      return sipStore.fetchCurrentWorkflows(sipStore.current.uuid);
    }
  }),
  null,
);

const tabs = [
  {
    icon: IconDetails,
    text: "Summary",
    route: router.resolve({
      name: "/ingest/sips/[id]/",
      params: { id: route.params.id },
    }),
    show: authStore.checkAttributes(["ingest:sips:read"]),
  },
];
</script>

<template>
  <div class="container-xxl">
    <PageLoadingAlert v-if="error" :execute="execute" :error="error" />

    <SipPendingAlert v-if="sipStore.current" />

    <h1 v-if="sipStore.current" class="d-flex mb-3">
      <IconSIPs class="me-3 text-dark" />{{ sipStore.current.name }}
    </h1>

    <Tabs :tabs="tabs" param="id" />

    <RouterView />
  </div>
</template>
