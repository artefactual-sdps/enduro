<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import IconLogin from "~icons/clarity/login-line";
</script>

<template>
  <div class="flex-grow-1 d-flex align-items-center justify-content-center">
    <button
      class="btn btn-lg btn-primary d-flex align-items-center"
      type="button"
      @click="useAuthStore().signinRedirect()"
    >
      <IconLogin aria-hidden="true" class="me-2" style="font-size: 25px" />
      Sign in
    </button>
  </div>
</template>
