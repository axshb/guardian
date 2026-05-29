<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import Card from '@/components/ui/Card.vue';
import Button from '@/components/ui/Button.vue';
import Input from '@/components/ui/Input.vue';
import Table from '@/components/ui/Table.vue';
import Badge from '@/components/ui/Badge.vue';
import Dialog from '@/components/ui/Dialog.vue';
import { useToast } from '@/composables/useToast';
import { Search } from 'lucide-vue-next';

interface ReferrerAlbum {
  album: string;
  images: { id: string; views: number; lastSeen: number }[];
}

interface ReferrerDetail {
  domain: string;
  banned: boolean;
  lastSeen: number;
  totalViews: number;
  albums: ReferrerAlbum[];
}

const toast = useToast();

const referrers = ref<ReferrerDetail[]>([]);
const redirectImage = ref('');
const search = ref('');
const page = ref(0);
const sortField = ref('lastSeen');
const sortOrder = ref(-1);
const detailDomain = ref<ReferrerDetail | null>(null);
const confirmAction = ref<{ domain: string; action: 'ban' | 'unban' | 'forget' } | null>(null);
const pending = ref<string | null>(null);

const PAGE_SIZE = 25;

onMounted(async () => {
  const res = await fetch('/api/referrers');
  const data = await res.json();
  referrers.value = data.domains || [];
  redirectImage.value = data.redirectImage || '';
});

const totalViews = computed(() => referrers.value.reduce((a, b) => a + b.totalViews, 0));
const bannedCount = computed(() => referrers.value.filter(r => r.banned).length);

const filtered = computed(() => {
  let rows = [...referrers.value];
  if (search.value) {
    rows = rows.filter(r => r.domain.toLowerCase().includes(search.value.toLowerCase()));
  }
  rows.sort((a, b) => {
    let cmp = 0;
    switch (sortField.value) {
      case 'domain': cmp = a.domain.localeCompare(b.domain); break;
      case 'status': cmp = Number(b.banned) - Number(a.banned); break;
      case 'views': cmp = a.totalViews - b.totalViews; break;
      case 'lastSeen': cmp = a.lastSeen - b.lastSeen; break;
      case 'albums': cmp = a.albums.length - b.albums.length; break;
    }
    return sortOrder.value === 1 ? cmp : -cmp;
  });
  return rows;
});

const paged = computed(() => {
  const start = page.value * PAGE_SIZE;
  return filtered.value.slice(start, start + PAGE_SIZE);
});

function onSort(e: any) {
  sortField.value = e.sortField;
  sortOrder.value = e.sortOrder;
  page.value = 0;
}

function fmtRelative(ts: number) {
  if (!ts) return '—';
  const diff = Date.now() / 1000 - ts;
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

async function handleAction(domain: string, action: 'ban' | 'unban' | 'forget') {
  pending.value = domain;
  try {
    await fetch(`/api/referrers/${action}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ domain }),
    });
    toast.add({ severity: 'success', summary: `${action === 'ban' ? 'Banned' : action === 'unban' ? 'Unbanned' : 'Forgot'} "${domain}"`, life: 3000 });
    const res = await fetch('/api/referrers');
    const data = await res.json();
    referrers.value = data.domains || [];
  } catch {
    toast.add({ severity: 'error', summary: `${action === 'ban' ? 'Ban' : action === 'unban' ? 'Unban' : 'Forget'} failed`, life: 3000 });
  } finally {
    pending.value = null;
    confirmAction.value = null;
  }
}

async function handleRedirectUpdate() {
  try {
    await fetch('/api/referrers/set-redirect', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url: redirectImage.value }),
    });
    toast.add({ severity: 'success', summary: 'Redirect image updated', life: 3000 });
  } catch {
    toast.add({ severity: 'error', summary: 'Update failed', life: 3000 });
  }
}
</script>

<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-2xl font-semibold">Referrers</h1>
      <p class="text-sm text-muted-foreground mt-1">Domain tracking & firewall.</p>
    </div>

    <!-- summary -->
    <div class="flex gap-3 text-xs">
      <Badge severity="info">{{ referrers.length }} Referrers</Badge>
      <Badge severity="danger">{{ bannedCount }} Banned</Badge>
      <Badge severity="success">{{ totalViews.toLocaleString() }} Views</Badge>
    </div>

    <!-- redirect image -->
    <Card class="p-6">
      <h3 class="text-sm font-medium mb-4">Redirect Image</h3>
      <div class="flex flex-col sm:flex-row gap-3">
        <Input v-model="redirectImage" placeholder="https://example.com/image.jpg" class="flex-1" />
        <Button @click="handleRedirectUpdate">Update</Button>
      </div>
      <img v-if="redirectImage" :src="redirectImage" alt="preview" class="max-h-32 w-auto object-contain opacity-60 mt-4 rounded-lg" />
    </Card>

    <!-- search -->
    <div class="flex items-center gap-3">
      <div class="relative flex-1 max-w-sm">
        <Search :size="14" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input v-model="search" @input="page = 0" placeholder="Filter domains..." class="pl-9 w-full" />
      </div>
      <span class="text-xs text-muted-foreground whitespace-nowrap">{{ filtered.length }} / {{ referrers.length }}</span>
    </div>

    <!-- table -->
    <Table class="border border-border rounded-lg">
      <thead>
        <tr class="border-b border-border">
          <th class="h-10 px-4 text-left align-middle font-medium text-muted-foreground text-xs uppercase tracking-wider cursor-pointer" @click="onSort({ sortField: 'domain', sortOrder: sortField === 'domain' ? -sortOrder : -1 })">Domain</th>
          <th class="h-10 px-4 text-left align-middle font-medium text-muted-foreground text-xs uppercase tracking-wider cursor-pointer" @click="onSort({ sortField: 'status', sortOrder: sortField === 'status' ? -sortOrder : -1 })">Status</th>
          <th class="h-10 px-4 text-right align-middle font-medium text-muted-foreground text-xs uppercase tracking-wider cursor-pointer" @click="onSort({ sortField: 'views', sortOrder: sortField === 'views' ? -sortOrder : -1 })">Views</th>
          <th class="h-10 px-4 text-right align-middle font-medium text-muted-foreground text-xs uppercase tracking-wider cursor-pointer" @click="onSort({ sortField: 'lastSeen', sortOrder: sortField === 'lastSeen' ? -sortOrder : -1 })">Last Seen</th>
          <th class="h-10 px-4 text-center align-middle font-medium text-muted-foreground text-xs uppercase tracking-wider cursor-pointer" @click="onSort({ sortField: 'albums', sortOrder: sortField === 'albums' ? -sortOrder : -1 })">Albums</th>
          <th class="h-10 px-4 text-right align-middle font-medium text-muted-foreground text-xs uppercase tracking-wider">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="data in paged" :key="data.domain" class="border-b border-border transition-colors hover:bg-muted/50">
          <td class="p-4">
            <span class="text-foreground font-medium cursor-pointer" @click="detailDomain = data">{{ data.domain }}</span>
          </td>
          <td class="p-4">
            <Badge :severity="data.banned ? 'danger' : 'success'">{{ data.banned ? 'Banned' : 'Active' }}</Badge>
          </td>
          <td class="p-4 text-right">
            <span class="text-warning">{{ data.totalViews.toLocaleString() }}</span>
          </td>
          <td class="p-4 text-right">
            <span class="text-muted-foreground">{{ fmtRelative(data.lastSeen) }}</span>
          </td>
          <td class="p-4 text-center">
            <span class="text-muted-foreground">{{ data.albums.length }}</span>
          </td>
          <td class="p-4 text-right">
            <div class="flex gap-2 justify-end">
              <Button variant="ghost" size="sm" :disabled="pending === data.domain" @click.stop="confirmAction = { domain: data.domain, action: data.banned ? 'unban' : 'ban' }">
                {{ data.banned ? 'Unban' : 'Ban' }}
              </Button>
              <Button variant="danger" size="sm" :disabled="pending === data.domain" @click.stop="confirmAction = { domain: data.domain, action: 'forget' }">
                Forget
              </Button>
            </div>
          </td>
        </tr>
      </tbody>
    </Table>

    <!-- pagination -->
    <div v-if="filtered.length > PAGE_SIZE" class="flex items-center justify-center gap-1">
      <Button variant="ghost" size="sm" :disabled="page === 0" @click="page = 0">First</Button>
      <Button variant="ghost" size="sm" :disabled="page === 0" @click="page--">Prev</Button>
      <span class="text-xs text-muted-foreground px-2">{{ page + 1 }} / {{ Math.ceil(filtered.length / PAGE_SIZE) }}</span>
      <Button variant="ghost" size="sm" :disabled="(page + 1) * PAGE_SIZE >= filtered.length" @click="page++">Next</Button>
      <Button variant="ghost" size="sm" :disabled="(page + 1) * PAGE_SIZE >= filtered.length" @click="page = Math.ceil(filtered.length / PAGE_SIZE) - 1">Last</Button>
    </div>

    <!-- domain detail dialog -->
    <Dialog :open="!!detailDomain" @update:open="(v) => { if (!v) detailDomain = null; }" class="max-w-lg">
      <div class="space-y-4">
        <h3 class="font-medium">{{ detailDomain?.domain }}</h3>
        <p class="text-muted-foreground text-xs">
          {{ detailDomain?.totalViews.toLocaleString() }} views · {{ detailDomain?.albums.length }} albums
        </p>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 max-h-60 overflow-y-auto">
          <router-link
            v-for="a in detailDomain?.albums"
            :key="a.album"
            :to="`/admin/assets/${encodeURIComponent(a.album)}`"
            class="block bg-background border border-border p-3 hover:border-primary/50 transition-colors rounded-lg"
          >
            <div class="text-primary text-xs mb-1">{{ a.album }}</div>
            <div class="text-xs text-muted-foreground">
              {{ a.images.reduce((s, i) => s + i.views, 0) }} views · {{ a.images.length }} images
            </div>
          </router-link>
        </div>
      </div>
    </Dialog>

    <!-- confirmation dialog -->
    <Dialog :open="!!confirmAction" @update:open="(v) => { if (!v) confirmAction = null; }" class="max-w-sm">
      <div class="space-y-4">
        <h3 class="font-medium">{{ confirmAction ? confirmAction.action.charAt(0).toUpperCase() + confirmAction.action.slice(1) : '' }} Domain</h3>
        <p class="text-muted-foreground text-xs">
          Are you sure you want to {{ confirmAction?.action }} "{{ confirmAction?.domain }}"?
        </p>
        <div class="flex justify-end gap-2">
          <Button variant="ghost" size="sm" @click="confirmAction = null">Cancel</Button>
          <Button
            :variant="confirmAction?.action === 'unban' ? 'default' : 'danger'"
            size="sm"
            :loading="!!pending"
            @click="confirmAction && handleAction(confirmAction.domain, confirmAction.action)"
          >
            {{ pending ? 'Working...' : confirmAction ? confirmAction.action.charAt(0).toUpperCase() + confirmAction.action.slice(1) : '' }}
          </Button>
        </div>
      </div>
    </Dialog>
  </div>
</template>
