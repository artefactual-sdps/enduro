<script setup lang="ts">
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import { useStorageStore } from "@/stores/storage";
import { useAsyncState } from "@vueuse/core";
import { useRoute } from "vue-router";
import IconRackServerLine from "~icons/clarity/rack-server-line";

const route = useRoute();
const storageStore = useStorageStore();

const { execute, error } = useAsyncState(
  storageStore.fetchCurrent(route.params.id.toString()),
  null
);
</script>

<template>
  <div class="container-xxl">
    <PageLoadingAlert v-if="error" :execute="execute" :error="error" />

    <h1 class="d-flex mb-3" v-if="storageStore.current">
      <IconRackServerLine class="me-3 text-dark" />{{
        storageStore.current.name
      }}
    </h1>

    <!-- Navigation tabs -->
    <ul class="nav nav-tabs mb-3" v-if="storageStore.current">
      <li class="nav-item">
        <router-link
          class="nav-link"
          exact-active-class="active"
          :to="{
            name: 'locations-id',
            params: { id: storageStore.current.uuid },
          }"
          >Summary</router-link
        >
      </li>
      <li class="nav-item">
        <router-link
          class="nav-link"
          exact-active-class="active"
          :to="{
            name: 'locations-id-packages',
            params: { id: storageStore.current.uuid },
          }"
          >Packages</router-link
        >
      </li>
    </ul>
    <router-view></router-view>
  </div>
</template>
