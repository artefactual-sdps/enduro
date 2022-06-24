<script setup lang="ts">
import { client, api } from "@/client";
import PackageStatusBadge from "@/components/PackageStatusBadge.vue";
import { onMounted, reactive } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();
const items: Array<api.EnduroStoredPackageResponseBody> = reactive([]);

onMounted(() => {
  client.package.packageList().then((resp) => {
    Object.assign(items, resp.items);
  });
});

const openPackage = (id: number) => {
  router.push({ name: "packages-id", params: { id } });
};
</script>

<template>
  <div class="container-xxl px-0">
    <h2>Packages</h2>
    <table class="table table-bordered table-hover">
      <thead>
        <tr>
          <th scope="col">ID</th>
          <th scope="col">Name</th>
          <th scope="col">UUID</th>
          <th scope="col">Started</th>
          <th scope="col">Location</th>
          <th scope="col">Status</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="pkg in items" :key="pkg.id" @click="openPackage(pkg.id)">
          <td scope="row">{{ pkg.id }}</td>
          <td>
            <router-link
              :to="{ name: 'packages-id', params: { id: pkg.id } }"
              >{{ pkg.name }}</router-link
            >
          </td>
          <td class="font-monospace">{{ pkg.aipId }}</td>
          <td>{{ $filters.formatDateTime(pkg.startedAt) }}</td>
          <td></td>
          <td>
            <PackageStatusBadge :status="pkg.status" />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
tbody tr {
  cursor: pointer;
}
</style>
