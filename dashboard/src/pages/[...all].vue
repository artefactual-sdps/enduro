<script setup lang="ts"></script>

<template>
  <div class="container-xxl pt-3">
    <div class="alert alert-warning" role="alert">
      <h4 class="alert-heading">Page not found!</h4>
      <p>We can't find the page you're looking for.</p>
      <hr />
      <router-link class="btn btn-warning" :to="{ name: 'index' }"
        >Take me home</router-link
      >
    </div>
  </div>
</template>
