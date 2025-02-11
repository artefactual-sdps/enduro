<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import { useRoute, useRouter } from "vue-router/auto";

import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import Tabs from "@/components/Tabs.vue";
import { useAuthStore } from "@/stores/auth";
import { useStorageStore } from "@/stores/storage";
import RawIconBundleLine from "~icons/clarity/bundle-line?raw&font-size=20px";
import RawIconDetailsLine from "~icons/clarity/details-line?raw&font-size=20px";
import IconRackServerLine from "~icons/clarity/rack-server-line";

const route = useRoute("/locations/[id]");
const router = useRouter();
const authStore = useAuthStore();
const storageStore = useStorageStore();

const { execute, error } = useAsyncState(
  storageStore.fetchCurrent(route.params.id.toString()),
  null,
);

const tabs = [
  {
    icon: RawIconDetailsLine,
    text: "Summary",
    route: router.resolve({
      name: "/locations/[id]/",
      params: { id: route.params.id },
    }),
    show: authStore.checkAttributes(["storage:location:read"]),
  },
  {
    icon: RawIconBundleLine,
    text: "Packages",
    route: router.resolve({
      name: "/locations/[id]/packages",
      params: { id: route.params.id },
    }),
    show: authStore.checkAttributes(["storage:location:listPackages"]),
  },
];
</script>

<template>
  <div class="container-xxl">
    <PageLoadingAlert v-if="error" :execute="execute" :error="error" />

    <h1 class="d-flex mb-3" v-if="storageStore.current">
      <IconRackServerLine class="me-3 text-dark" />{{
        storageStore.current.name
      }}
    </h1>

    <Tabs :tabs="tabs" param="id" />

    <router-view></router-view>
  </div>
</template>
