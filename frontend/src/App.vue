<script setup lang="ts">
import { onMounted } from "vue";
import AppShell from "@/components/shell/AppShell.vue";
import Toast from "@/components/Toast.vue";
import { useAuthStore } from "@/stores/auth";

onMounted(() => useAuthStore().initialize());
</script>

<template>
  <AppShell>
    <RouterView />
  </AppShell>
  <Toast />
</template>
