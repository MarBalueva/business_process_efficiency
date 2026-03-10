<template>
  <div
    class="folder-block"
    @dragover.prevent
    @drop.prevent="onDropToFolder"
  >

    <div class="folder-header">

      <div class="folder-title" @click="toggleFolder" draggable="true" @dragstart="onFolderDragStart">
        <ChevronRight
          v-if="hasChildren"
          :size="16"
          class="chevron"
          :class="{ open: isOpen }"
        />

        <span>{{ folder.name }}</span>
      </div>

      <div class="dict-actions">
        <button class="icon-btn edit" @click="emit('edit-folder', folder)">
          <Pencil :size="16"/>
        </button>

        <button class="icon-btn delete" @click="emit('delete-folder', folder.id)">
          <Trash :size="16"/>
        </button>
      </div>

    </div>

    <!-- содержимое папки -->
    <div v-if="isOpen">

      <div class="processes-list" @dragover.prevent @drop.prevent="onDropToFolder">
        <div
          v-for="process in folder.processes"
          :key="process.id"
          class="process-row"
        >
          <span
            class="process-drag-handle"
            title="Перетащить процесс"
            draggable="true"
            @dragstart.stop="onProcessDragStart(process.id)"
          >
            <GripVertical :size="14" />
          </span>
          <router-link
            :to="`/processes/${process.id}`"
            class="process-name"
          >
            {{ process.name }}
          </router-link>
        </div>
      </div>

      <ProcessFolder
        v-for="child in folder.children"
        :key="child.id"
        :folder="child"
        @delete-folder="$emit('delete-folder', $event)"
        @edit-folder="$emit('edit-folder', $event)"
        @move-process="$emit('move-process', $event)"
        @move-folder="$emit('move-folder', $event)"
      />

    </div>

  </div>
</template>

<script setup>
import { ref, computed } from "vue"
import { Pencil, Trash, ChevronRight, GripVertical } from "lucide-vue-next"

const props = defineProps({
  folder: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(["delete-folder", "edit-folder", "move-process", "move-folder"])

const isOpen = ref(true)

const hasChildren = computed(() => {
  return props.folder.children?.length || props.folder.processes?.length
})

const toggleFolder = () => {
  if (hasChildren.value) {
    isOpen.value = !isOpen.value
  }
}

const onProcessDragStart = (processId) => (event) => {
  const payload = JSON.stringify({ type: "process", id: Number(processId) })
  window.__bpeDragPayload = payload
  event.dataTransfer.effectAllowed = "move"
  event.dataTransfer.setData("application/json", payload)
  event.dataTransfer.setData("text/plain", payload)
}

const onFolderDragStart = (event) => {
  const payload = JSON.stringify({ type: "folder", id: Number(props.folder.id) })
  window.__bpeDragPayload = payload
  event.dataTransfer.effectAllowed = "move"
  event.dataTransfer.setData("application/json", payload)
  event.dataTransfer.setData("text/plain", payload)
}

const onDropToFolder = (event) => {
  const raw =
    event.dataTransfer.getData("application/json") ||
    event.dataTransfer.getData("text/plain") ||
    window.__bpeDragPayload
  if (!raw) return

  let payload
  try {
    payload = JSON.parse(raw)
  } catch {
    return
  }

  if (payload?.type === "process") {
    const processId = Number(payload.id)
    const targetFolderId = Number(props.folder.id)
    if (Number.isFinite(processId) && Number.isFinite(targetFolderId)) {
      emit("move-process", { processId, targetFolderId })
      window.__bpeDragPayload = null
    }
    return
  }

  if (payload?.type === "folder") {
    const folderId = Number(payload.id)
    const targetParentId = Number(props.folder.id)
    if (!Number.isFinite(folderId) || !Number.isFinite(targetParentId) || folderId === targetParentId) {
      return
    }
    emit("move-folder", { folderId, targetParentId })
    window.__bpeDragPayload = null
  }
}
</script>

<style scoped>

.folder-block { 
  margin-left:20px; 
  border-left:1px solid #eee; 
  padding-left:10px; 
  margin-bottom:10px; 
}

.folder-header { 
  display:flex; 
  justify-content:space-between; 
  align-items:center; 
  margin-bottom:4px; 
}

.folder-title{
  display:flex;
  align-items:center;
  gap:6px;
  cursor:pointer;
  font-weight:600;
}

.chevron{
  transition: transform 0.2s;
}

.chevron.open{
  transform: rotate(90deg);
}

.processes-list {
  margin-left: 10px;
}

.process-row { 
  padding:4px 0;
  display:flex;
  align-items:center;
  gap:6px;
}

.process-drag-handle {
  display:inline-flex;
  align-items:center;
  justify-content:center;
  width:18px;
  height:18px;
  color:#6b7280;
  cursor:grab;
  border-radius:4px;
}

.process-drag-handle:hover {
  background:#eef2ff;
}

.process-drag-handle:active {
  cursor:grabbing;
}

.process-name { 
  color:var(--color-primary); 
  text-decoration:none; 
}

.process-name:hover { 
  text-decoration:underline; 
}

.dict-actions{
  display:flex;
  gap:10px;
}

.icon-btn{
  border:none;
  background:var(--color-soft-bg);
  border-radius:6px;
  padding:6px;
  cursor:pointer;
}

.icon-btn.edit:hover{
  background:var(--color-edit-hover);
}

.icon-btn.delete:hover{
  background:var(--color-delete-hover);
}

</style>

