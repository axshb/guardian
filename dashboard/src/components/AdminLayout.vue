<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import Sheet from '@/components/ui/Sheet.vue';
import Button from '@/components/ui/Button.vue';
import { Image, Shield, Palette, LogOut, Menu as MenuIcon } from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const mobileOpen = ref(false);

const navItems = [
  { name: 'assets', href: '/assets', icon: Image },
  { name: 'referrers', href: '/referrers', icon: Shield },
  { name: 'settings', href: '/settings', icon: Palette },
];

function isActive(href: string) {
  return route.path === href || route.path.startsWith(`${href}/`);
}

async function handleLogout() {
  await auth.logout();
  router.push('/');
}

async function navigateTo(href: string) {
  mobileOpen.value = false;
  await router.push(href);
}

onMounted(async () => {
  try {
    const res = await fetch('/api/config');
    const cfg = await res.json();
    if (cfg.theme) document.documentElement.dataset.theme = cfg.theme;
  } catch {
    // ignore
  }
});
</script>

<template>
  <div class="min-h-screen bg-background text-foreground flex">
    <!-- desktop sidebar -->
    <aside class="hidden lg:flex w-60 border-r border-border fixed h-screen bg-background z-20 flex-col">
      <div class="px-5 py-6 mb-2">
        <h1 class="text-xl font-semibold">guardian</h1>
        <div class="h-px w-8 bg-primary mt-3" />
      </div>
      <nav class="flex-1 space-y-0.5 px-3">
        <button
          v-for="item in navItems"
          :key="item.href"
          @click="navigateTo(item.href)"
          class="flex items-center gap-3 px-3 py-2.5 text-sm rounded-md transition-colors w-full text-left"
          :class="isActive(item.href)
            ? 'bg-muted text-foreground font-medium'
            : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'"
        >
          <component :is="item.icon" :size="16" />
          <span>{{ item.name }}</span>
        </button>
      </nav>
      <div class="px-3 py-4 border-t border-border">
        <button
          @click="handleLogout"
          class="flex items-center gap-3 px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground hover:bg-muted/50 rounded-md transition-colors w-full text-left"
        >
          <LogOut :size="16" />
          <span>logout</span>
        </button>
      </div>
    </aside>

    <!-- mobile header -->
    <div class="lg:hidden fixed top-0 left-0 right-0 h-14 border-b border-border bg-background z-30 flex items-center px-4 gap-3">
      <Button variant="ghost" size="sm" @click="mobileOpen = true">
        <MenuIcon :size="18" />
      </Button>
      <span class="text-sm font-medium">guardian</span>
    </div>

    <Sheet :open="mobileOpen" @update:open="(v) => mobileOpen = v" side="left" class="w-60">
      <div class="px-5 py-6 mb-2">
        <h1 class="text-xl font-semibold">guardian</h1>
        <div class="h-px w-8 bg-primary mt-3" />
      </div>
      <nav class="flex-1 space-y-0.5 px-3">
        <button
          v-for="item in navItems"
          :key="item.href"
          @click="navigateTo(item.href)"
          class="flex items-center gap-3 px-3 py-2.5 text-sm rounded-md transition-colors w-full text-left"
          :class="isActive(item.href)
            ? 'bg-muted text-foreground font-medium'
            : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'"
        >
          <component :is="item.icon" :size="16" />
          <span>{{ item.name }}</span>
        </button>
      </nav>
      <div class="px-3 py-4 border-t border-border">
        <button
          @click="handleLogout"
          class="flex items-center gap-3 px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground hover:bg-muted/50 rounded-md transition-colors w-full text-left"
        >
          <LogOut :size="16" />
          <span>logout</span>
        </button>
      </div>
    </Sheet>

    <!-- content area -->
    <main class="flex-1 lg:ml-60 pt-14 lg:pt-0">
      <div class="p-6 lg:p-8 max-w-6xl mx-auto">
        <router-view />
      </div>
    </main>
  </div>
</template>
