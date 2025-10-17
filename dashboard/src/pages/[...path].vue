<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";

useLayoutStore().updateBreadcrumb([{ text: "Not found" }]);
</script>

<template>
  <div class="container-xxl">
    <div class="alert alert-warning" role="alert">
      <h4 class="alert-heading">Page not found!</h4>
      <p>We can't find the page you're looking for.</p>
      <hr />
      <RouterLink class="btn btn-warning" :to="{ name: '/' }">
        Take me home
      </RouterLink>
    </div>
  </div>
</template>
