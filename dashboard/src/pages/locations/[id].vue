<script setup lang="ts">
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import Tabs from "@/components/Tabs.vue";
import { useStorageStore } from "@/stores/storage";
import { useAsyncState } from "@vueuse/core";
import { useRoute } from "vue-router";
import RawIconBundleLine from "~icons/clarity/bundle-line?raw&font-size=20px";
import RawIconDetailsLine from "~icons/clarity/details-line?raw&font-size=20px";
import IconRackServerLine from "~icons/clarity/rack-server-line";

const route = useRoute();
const storageStore = useStorageStore();

const { execute, error } = useAsyncState(
  storageStore.fetchCurrent(route.params.id.toString()),
  null
);

const tabs = [
  {
    icon: RawIconDetailsLine,
    text: "Summary",
    route: { name: "locations-id", params: { id: route.params.id } },
  },
  {
    icon: RawIconBundleLine,
    text: "Packages",
    route: { name: "locations-id-packages", params: { id: route.params.id } },
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

    <Tabs :tabs="tabs" />

    <router-view></router-view>
  </div>
</template>
