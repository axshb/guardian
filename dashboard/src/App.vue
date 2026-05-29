<script setup lang="ts">
import { onMounted } from 'vue';
import { useAuthStore } from './stores/auth';

const auth = useAuthStore();

onMounted(async () => {
  auth.checkAuth();
  try {
    const res = await fetch('/api/config');
    const cfg = await res.json();
    if (cfg.theme) {
      document.documentElement.dataset.theme = cfg.theme;
    }
  } catch {
    // ignore
  }
});
</script>

<template>
  <router-view />
  <ToastContainer />
</template>
