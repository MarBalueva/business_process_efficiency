<template>
  <div class="folder-block">

    <div class="folder-header">

      <div class="folder-title" @click="toggleFolder">
        <ChevronRight
          v-if="hasChildren"
          :size="16"
          class="chevron"
          :class="{ open: isOpen }"
        />

        <span>{{ folder.name }}</span>
      </div>

      <div class="dict-actions">
        <button class="icon-btn edit" @click="$emit('edit-folder', folder.id)">
          <Pencil :size="16"/>
        </button>

        <button class="icon-btn delete" @click="$emit('delete-folder', folder.id)">
          <Trash :size="16"/>
        </button>
      </div>

    </div>

    <!-- содержимое папки -->
    <div v-if="isOpen">

      <div class="processes-list">
        <div v-for="process in folder.processes" :key="process.id" class="process-row">
          <router-link :to="`/processes/${process.id}`" class="process-name">
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
      />

    </div>

  </div>
</template>

<script setup>
import { ref, computed } from "vue"
import { Pencil, Trash, ChevronRight } from "lucide-vue-next"

const props = defineProps({
  folder: {
    type: Object,
    required: true
  }
})

defineEmits(["delete-folder", "edit-folder"])

const isOpen = ref(true)

const hasChildren = computed(() => {
  return props.folder.children?.length || props.folder.processes?.length
})

const toggleFolder = () => {
  if (hasChildren.value) {
    isOpen.value = !isOpen.value
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
}

.process-name { 
  color:#4f46e5; 
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
  background:#f3f4f6;
  border-radius:6px;
  padding:6px;
  cursor:pointer;
}

.icon-btn.edit:hover{
  background:#e0e7ff;
}

.icon-btn.delete:hover{
  background:#fee2e2;
}

</style>