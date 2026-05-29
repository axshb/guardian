<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import Table from '@/components/ui/Table.vue';
import ScrollArea from '@/components/ui/ScrollArea.vue';

interface AssetView {
  domain: string;
  count: number;
  first_seen: number;
  last_seen: number;
}

interface AssetStatsData {
  total: number;
  referrers: AssetView[];
}

const props = defineProps<{ assetId: string }>();

const stats = ref<AssetStatsData | null>(null);
const loading = ref(true);

async function loadStats() {
  loading.value = true;
  try {
    const res = await fetch(`/api/assets/${props.assetId}/stats`);
    stats.value = await res.json();
  } finally {
    loading.value = false;
  }
}

onMounted(loadStats);
watch(() => props.assetId, loadStats);

function fmtDate(ts: number) {
  if (!ts) return '—';
  return new Date(ts * 1000).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}
</script>

<template>
  <ScrollArea class="border border-border max-h-48 rounded-lg">
    <div v-if="loading" class="p-3 text-xs text-muted-foreground">Loading...</div>
    <div v-else-if="!stats || stats.total === 0" class="p-3 text-xs text-muted-foreground">No views yet.</div>
    <div v-else-if="stats.referrers.length === 0" class="p-3 text-xs text-muted-foreground">{{ stats.total.toLocaleString() }} views · referrer data not available</div>
    <Table v-else class="text-xs">
      <thead>
        <tr class="border-b border-border">
          <th class="h-8 px-3 text-left font-medium text-muted-foreground uppercase text-[10px] tracking-wider">Domain</th>
          <th class="h-8 px-3 text-left font-medium text-muted-foreground uppercase text-[10px] tracking-wider">Count</th>
          <th class="h-8 px-3 text-left font-medium text-muted-foreground uppercase text-[10px] tracking-wider">First</th>
          <th class="h-8 px-3 text-left font-medium text-muted-foreground uppercase text-[10px] tracking-wider">Last</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in stats.referrers" :key="r.domain" class="border-b border-border">
          <td class="p-3 text-foreground">{{ r.domain }}</td>
          <td class="p-3 text-warning">{{ r.count }}</td>
          <td class="p-3 text-muted-foreground">{{ fmtDate(r.first_seen) }}</td>
          <td class="p-3 text-muted-foreground">{{ fmtDate(r.last_seen) }}</td>
        </tr>
      </tbody>
    </Table>
  </ScrollArea>
</template>
