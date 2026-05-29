<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import Card from '@/components/ui/Card.vue';
import Button from '@/components/ui/Button.vue';
import Input from '@/components/ui/Input.vue';
import Textarea from '@/components/ui/Textarea.vue';
import Dialog from '@/components/ui/Dialog.vue';
import { useToast } from '@/composables/useToast';
import SortableDraggable from '@/components/assets/SortableDraggable.vue';
import { Trash2 } from 'lucide-vue-next';
import { assetUrl } from '@/lib/contentType';

interface AlbumInfo {
  album: string;
  count: number;
  cover?: string;
  coverContentType?: string;
  position?: number;
}

const router = useRouter();
const toast = useToast();

const albums = ref<AlbumInfo[]>([]);
const importAlbum = ref('');
const importUrls = ref('');
const importing = ref(false);
const deleteTarget = ref<string | null>(null);
const deleting = ref(false);

async function loadAlbums() {
  const res = await fetch('/api/albums');
  albums.value = res.ok ? await res.json() : [];
}

onMounted(loadAlbums);

async function handleImport() {
  if (!importAlbum.value || !importUrls.value) return;
  importing.value = true;
  const urls = importUrls.value.split('\n').map(u => u.trim()).filter(Boolean);
  try {
    const res = await fetch('/api/albums', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ album: importAlbum.value, urls }),
    });
    if (!res.ok) throw new Error();
    const albumName = importAlbum.value;
    importUrls.value = '';
    importAlbum.value = '';
    toast.add({ severity: 'success', summary: `Imported into "${albumName}"`, life: 3000 });
    await loadAlbums();
  } catch {
    toast.add({ severity: 'error', summary: 'Import failed', life: 3000 });
  } finally {
    importing.value = false;
  }
}

async function handleDelete(album: string) {
  deleting.value = true;
  try {
    const res = await fetch(`/api/albums/${encodeURIComponent(album)}`, { method: 'DELETE' });
    if (!res.ok) throw new Error();
    albums.value = albums.value.filter(a => a.album !== album);
    toast.add({ severity: 'success', summary: `Deleted "${album}"`, life: 3000 });
  } catch {
    toast.add({ severity: 'error', summary: 'Delete failed', life: 3000 });
  } finally {
    deleting.value = false;
    deleteTarget.value = null;
  }
}

async function onReorder(newList: AlbumInfo[]) {
  albums.value = newList;
  await fetch('/api/albums/reorder', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ ids: newList.map(a => a.album) }),
  });
}
</script>

<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-2xl font-semibold">Assets</h1>
      <p class="text-sm text-muted-foreground mt-1">Image album management.</p>
    </div>

    <!-- album grid -->
    <SortableDraggable
      v-model="albums"
      item-key="album"
      tag="div"
      class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4"
      @end="onReorder(albums)"
    >
      <div
        v-for="a in albums"
        :key="a.album"
        class="group relative border border-border hover:border-primary/50 rounded-xl transition-all cursor-grab active:cursor-grabbing"
      >
        <Card class="border-0 bg-transparent overflow-hidden rounded-xl">
          <router-link :to="`/admin/assets/${encodeURIComponent(a.album)}`" class="block">
            <div class="aspect-square bg-background relative">
              <img v-if="a.cover" :src="assetUrl(a.cover, a.coverContentType)" class="w-full h-full object-cover" alt="" draggable="false" />
              <div v-else class="w-full h-full flex items-center justify-center text-muted-foreground text-xs">
                No images
              </div>
            </div>
          </router-link>
          <div class="flex items-center justify-between px-4 py-3">
            <router-link :to="`/admin/assets/${encodeURIComponent(a.album)}`" class="min-w-0">
              <div class="text-sm text-foreground font-medium truncate">{{ a.album }}</div>
              <div class="text-xs text-muted-foreground mt-0.5">{{ a.count }} images</div>
            </router-link>
            <button
              @click.stop="deleteTarget = a.album"
              class="text-muted-foreground hover:text-destructive p-1 rounded transition-colors"
              title="Delete album"
            >
              <Trash2 :size="14" />
            </button>
          </div>
        </Card>
      </div>
    </SortableDraggable>

    <!-- delete dialog -->
    <Dialog :open="!!deleteTarget" @update:open="(v) => { if (!v) deleteTarget = null; }" class="max-w-sm">
      <div class="space-y-4">
        <h3 class="text-foreground text-sm font-medium">Delete Album</h3>
        <p class="text-muted-foreground text-xs">
          Delete "{{ deleteTarget }}" and all its images? This cannot be undone.
        </p>
        <div class="flex justify-end gap-2">
          <Button variant="ghost" size="sm" @click="deleteTarget = null">Cancel</Button>
          <Button variant="danger" size="sm" :loading="deleting" @click="deleteTarget && handleDelete(deleteTarget)">Delete</Button>
        </div>
      </div>
    </Dialog>

    <!-- bulk import -->
    <Card class="p-6">
      <h3 class="text-sm font-medium mb-4">Bulk Import</h3>
      <form @submit.prevent="handleImport" class="space-y-4">
        <div class="space-y-1">
          <label class="text-xs text-muted-foreground">Album name</label>
          <Input v-model="importAlbum" placeholder="e.g. concept_art" required class="w-full" />
        </div>
        <div class="space-y-1">
          <label class="text-xs text-muted-foreground">Source URLs (one per line)</label>
          <Textarea v-model="importUrls" placeholder="https://imgchest.com/p/..." class="h-48 w-full" required />
        </div>
        <Button type="submit" class="w-full" :loading="importing">Import Images</Button>
      </form>
    </Card>
  </div>
</template>
