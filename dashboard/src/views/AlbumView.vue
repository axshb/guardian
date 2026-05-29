<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import Card from '@/components/ui/Card.vue';
import Button from '@/components/ui/Button.vue';
import Input from '@/components/ui/Input.vue';
import Dialog from '@/components/ui/Dialog.vue';
import { useToast } from '@/composables/useToast';
import SortableDraggable from '@/components/assets/SortableDraggable.vue';
import { ChevronLeft, Copy, Pencil, Trash2 } from 'lucide-vue-next';
import AssetStats from '@/components/assets/AssetStats.vue';
import { assetUrl } from '@/lib/contentType';

interface Image {
  id: string;
  views: number;
  position?: number;
  contentType?: string;
}

const route = useRoute();
const router = useRouter();
const toast = useToast();

const album = computed(() => decodeURIComponent(route.params.album as string));

const images = ref<Image[]>([]);
const modalImage = ref<string | null>(null);
const editingId = ref<string | null>(null);
const editUrl = ref('');
const deleteTarget = ref<string | null>(null);
const deleting = ref(false);

async function loadImages() {
  const res = await fetch(`/api/albums/${encodeURIComponent(album.value)}/images`);
  images.value = res.ok ? await res.json() : [];
}

onMounted(loadImages);

async function handleDelete(id: string) {
  deleting.value = true;
  try {
    const res = await fetch(`/api/albums/${encodeURIComponent(album.value)}/images/${encodeURIComponent(id)}`, { method: 'DELETE' });
    if (!res.ok) throw new Error();
    images.value = images.value.filter(i => i.id !== id);
    toast.add({ severity: 'success', summary: 'Image deleted', life: 3000 });
  } catch {
    toast.add({ severity: 'error', summary: 'Delete failed', life: 3000 });
  } finally {
    deleting.value = false;
    deleteTarget.value = null;
  }
}

async function handleUpdateSource(id: string) {
  try {
    const res = await fetch(`/api/albums/${encodeURIComponent(album.value)}/images/${encodeURIComponent(id)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url: editUrl.value }),
    });
    if (!res.ok) throw new Error();
    toast.add({ severity: 'success', summary: 'Source updated', life: 3000 });
  } catch {
    toast.add({ severity: 'error', summary: 'Update failed', life: 3000 });
  } finally {
    editingId.value = null;
    editUrl.value = '';
  }
}

async function onReorder(newList: Image[]) {
  images.value = newList;
  await fetch(`/api/albums/${encodeURIComponent(album.value)}/reorder`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ ids: newList.map(i => i.id) }),
  });
}

async function copyUrl(img: Image) {
  await navigator.clipboard.writeText(`${window.location.origin}${assetUrl(img.id, img.contentType)}`);
  toast.add({ severity: 'success', summary: 'URL copied', life: 3000 });
}
</script>

<template>
  <div class="space-y-8">
    <div>
      <router-link to="/assets" class="inline-flex items-center gap-1 text-sm text-primary hover:text-foreground transition-colors">
        <ChevronLeft :size="14" />
        Back to Assets
      </router-link>
      <h1 class="text-2xl font-semibold mt-4">{{ album }}</h1>
      <p class="text-sm text-muted-foreground mt-1">{{ images.length }} images</p>
    </div>

    <SortableDraggable
      v-model="images"
      item-key="id"
      tag="div"
      class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
      @end="onReorder(images)"
    >
      <div
        v-for="img in images"
        :key="img.id"
        class="transition-all"
      >
        <Card class="border border-border overflow-hidden rounded-xl">
          <button @click="modalImage = img.id" class="aspect-[4/3] bg-background border-b border-border overflow-hidden cursor-zoom-in block w-full">
            <img :src="assetUrl(img.id, img.contentType)" class="w-full h-full object-cover" alt="" draggable="false" />
          </button>
          <div class="px-4 py-3">
            <div class="text-xs text-muted-foreground truncate">{{ img.id }}</div>
            <div class="text-xs text-warning mt-1">{{ img.views ?? 0 }} views</div>

            <div v-if="editingId === img.id" class="flex gap-2 mt-3">
              <Input v-model="editUrl" placeholder="New source URL" class="text-xs flex-1 min-w-0" />
              <Button size="sm" @click="handleUpdateSource(img.id)">Save</Button>
              <Button variant="ghost" size="sm" @click="editingId = null; editUrl = ''">Cancel</Button>
            </div>

            <div v-else class="flex gap-2 mt-3">
              <Button variant="ghost" size="sm" class="flex-1" @click="copyUrl(img)">
                <Copy :size="12" class="mr-1" /> Copy
              </Button>
              <Button variant="ghost" size="sm" class="flex-1" @click="editingId = img.id; editUrl = ''">
                <Pencil :size="12" class="mr-1" /> Edit
              </Button>
              <Button variant="danger" size="sm" class="flex-1" @click="deleteTarget = img.id">
                <Trash2 :size="12" class="mr-1" /> Delete
              </Button>
            </div>
          </div>
        </Card>
      </div>
    </SortableDraggable>

    <!-- image modal -->
    <Dialog :open="!!modalImage" @update:open="(v) => { if (!v) modalImage = null; }" class="max-w-3xl">
      <div v-if="modalImage" class="flex flex-col max-h-[85vh]">
        <div class="flex items-center justify-center bg-black/40 min-h-[200px] rounded-t-lg">
          <img :src="assetUrl(modalImage, images.find(i => i.id === modalImage)?.contentType)" alt="" class="max-w-full max-h-[50vh] object-contain" />
        </div>
        <div class="p-4 overflow-y-auto">
          <AssetStats :asset-id="modalImage" />
        </div>
      </div>
    </Dialog>

    <!-- delete dialog -->
    <Dialog :open="!!deleteTarget" @update:open="(v) => { if (!v) deleteTarget = null; }" class="max-w-sm">
      <div class="space-y-4">
        <h3 class="text-foreground text-sm font-medium">Delete Image</h3>
        <p class="text-muted-foreground text-xs">
          Permanently delete this image? This cannot be undone.
        </p>
        <div class="flex justify-end gap-2">
          <Button variant="ghost" size="sm" @click="deleteTarget = null">Cancel</Button>
          <Button variant="danger" size="sm" :loading="deleting" @click="deleteTarget && handleDelete(deleteTarget)">Delete</Button>
        </div>
      </div>
    </Dialog>
  </div>
</template>
