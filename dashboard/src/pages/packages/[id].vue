<script setup lang="ts">
import PackagePendingAlert from "@/components/PackagePendingAlert.vue";
import { usePackageStore } from "@/stores/package";
import { useRoute } from "vue-router";

const route = useRoute();
const packageStore = usePackageStore();

packageStore.fetchCurrent(route.params.id.toString());
</script>

<template>
  <div v-if="packageStore.current">
    <PackagePendingAlert />

    <div class="container-xxl px-0">
      <!-- Breadcrumb -->
      <div>
        <nav aria-label="breadcrumb">
          <ol class="breadcrumb">
            <li class="breadcrumb-item">
              <router-link :to="{ name: 'packages' }">Packages</router-link>
            </li>
            <li class="breadcrumb-item active" aria-current="page">
              {{ packageStore.current.name }}
            </li>
          </ol>
        </nav>
      </div>

      <!-- Navigation tabs -->
      <div>
        <ul class="nav nav-tabs">
          <li class="nav-item">
            <router-link
              class="nav-link"
              exact-active-class="active"
              :to="{
                name: 'packages-id',
                params: { id: packageStore.current.id },
              }"
              >Overview</router-link
            >
          </li>
          <li class="nav-item">
            <router-link
              class="nav-link"
              exact-active-class="active"
              :to="{
                name: 'packages-id-workflow',
                params: { id: packageStore.current.id },
              }"
              >Workflow</router-link
            >
          </li>
        </ul>
      </div>
      <div class="pt-3">
        <router-view></router-view>
      </div>
    </div>
  </div>
</template>
