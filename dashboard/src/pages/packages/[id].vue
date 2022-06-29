<script setup lang="ts">
import PackagePendingAlert from "@/components/PackagePendingAlert.vue";
import { usePackageStore } from "@/stores/package";
import { useRoute } from "vue-router";

const route = useRoute();
const packageStore = usePackageStore();

await packageStore.fetchCurrent(route.params.id.toString());
</script>

<template>
  <div v-if="packageStore.current">
    <!-- Navigation bar -->
    <div class="container-fluid pt-3 packages-navbar">
      <div class="container-xxl px-0">
        <div class="col">
          <PackagePendingAlert />
        </div>
        <!-- Breadcrumb -->
        <div class="col">
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
        <div class="col">
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
      </div>
    </div>

    <div class="container-xxl px-3">
      <div class="pt-3">
        <router-view></router-view>
      </div>
    </div>
  </div>
</template>
