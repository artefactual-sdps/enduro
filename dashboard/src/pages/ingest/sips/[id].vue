<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import { useRoute, useRouter } from "vue-router/auto";

import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import SipPendingAlert from "@/components/SipPendingAlert.vue";
import Tabs from "@/components/Tabs.vue";
import { useAuthStore } from "@/stores/auth";
import { useIngestStore } from "@/stores/ingest";
import IconDetails from "~icons/clarity/details-line?raw&font-size=20px";
import IconSIPs from "~icons/octicon/package-dependencies-24";

const route = useRoute("/ingest/sips/[id]");
const router = useRouter();
const authStore = useAuthStore();
const ingestStore = useIngestStore();

const { execute, error } = useAsyncState(
  ingestStore.fetchCurrentSip(route.params.id.toString()),
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

    <SipPendingAlert v-if="ingestStore.currentSip" />

    <h1 class="d-flex mb-3" v-if="ingestStore.currentSip">
      <IconSIPs class="me-3 text-dark" />{{ ingestStore.currentSip.name }}
    </h1>

    <Tabs :tabs="tabs" param="id" />

    <router-view></router-view>
  </div>
</template>
