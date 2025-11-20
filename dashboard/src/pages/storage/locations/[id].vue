<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import { useRoute, useRouter } from "vue-router/auto";

import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import Tabs from "@/components/Tabs.vue";
import { useAuthStore } from "@/stores/auth";
import { useLocationStore } from "@/stores/location";
import IconAIPs from "~icons/clarity/bundle-line?font-size=20px";
import IconDetails from "~icons/clarity/details-line?font-size=20px";
import IconLocations from "~icons/octicon/server-24";

const route = useRoute("/storage/locations/[id]");
const router = useRouter();
const authStore = useAuthStore();
const locationStore = useLocationStore();

const { execute, error } = useAsyncState(
  locationStore.fetchCurrent(route.params.id.toString()),
  null,
);

const tabs = [
  {
    icon: IconDetails,
    text: "Summary",
    route: router.resolve({
      name: "/storage/locations/[id]/",
      params: { id: route.params.id },
    }),
    show: authStore.checkAttributes(["storage:locations:read"]),
  },
  {
    icon: IconAIPs,
    text: "AIPs",
    route: router.resolve({
      name: "/storage/locations/[id]/aips",
      params: { id: route.params.id },
    }),
    show: authStore.checkAttributes(["storage:locations:aips:list"]),
  },
];
</script>

<template>
  <div class="container-xxl">
    <PageLoadingAlert v-if="error" :execute="execute" :error="error" />

    <h1 v-if="locationStore.current" class="d-flex mb-3">
      <IconLocations class="me-3 text-dark" />{{ locationStore.current.name }}
    </h1>

    <Tabs :tabs="tabs" param="id" />

    <RouterView />
  </div>
</template>
