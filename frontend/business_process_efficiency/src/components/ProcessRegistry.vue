<template>
  <div class="employees-page-container">
    <div class="employees-card">

      <div class="employees-header">
        <h2>Реестр процессов</h2>

        <div class="header-actions">

          <div class="search-box">
            <Search :size="16" />
            <input v-model="search" placeholder="Поиск по папкам или процессам..." />
          </div>

          <button class="add-btn" @click="openCreateFolderModal">
            Создать папку
          </button>

          <button class="add-btn" @click="openCreateProcessModal">
            Добавить процесс
          </button>

        </div>
      </div>

      <div class="employees-table">

        <ProcessFolder
          v-for="folder in paginatedFolders"
          :key="folder.id"
          :folder="folder"
          @delete-folder="deleteFolder"
          @edit-folder="openEditFolderModal"
        />

      </div>

      <div class="pagination" v-if="totalPages > 1">
        <button :disabled="currentPage === 1" @click="prevPage">◀</button>
        <span>{{ currentPage }} из {{ totalPages }}</span>
        <button :disabled="currentPage === totalPages" @click="nextPage">▶</button>
      </div>

    </div>

    <div v-if="folderModalOpen" class="modal-overlay">
      <div class="modal modal-wide">

        <h3>{{ editingFolder ? "Редактировать папку" : "Создать папку" }}</h3>

        <label>Название</label>
        <input v-model="folderForm.name" type="text" />

        <label>Родительская папка</label>

        <select v-model="folderForm.parentId">

          <option :value="null">Без родителя</option>

          <option
            v-for="f in treeFolders"
            :key="f.id"
            :value="f.id"
            :disabled="f.id === folderForm.id"
          >
            {{ f.label }}
          </option>

        </select>

        <div class="modal-actions">
          <button class="save-btn" @click="saveFolder">
            {{ editingFolder ? "Сохранить" : "Создать" }}
          </button>

          <button class="cancel-btn" @click="folderModalOpen=false">
            Отмена
          </button>
        </div>

      </div>
    </div>


    <div v-if="processModalOpen" class="modal-overlay">

      <div class="modal modal-wide">

        <h3>Создать процесс</h3>

        <label>Название</label>
        <input v-model="processForm.name" type="text" />

        <label>Папка</label>

        <select v-model="processForm.folderId">

          <option
            v-for="f in treeFolders"
            :key="f.id"
            :value="f.id"
          >
            {{ f.label }}
          </option>

        </select>

        <div class="modal-actions">
          <button class="save-btn" @click="saveProcess">
            Сохранить
          </button>

          <button class="cancel-btn" @click="processModalOpen=false">
            Отмена
          </button>
        </div>

      </div>

    </div>

    <div v-if="toast.show" class="toast">
      {{ toast.message }}
    </div>

  </div>
</template>

<script setup>

import { ref, onMounted, computed, watch } from "vue"
import { Search } from "lucide-vue-next"
import api from "../api/axios"
import ProcessFolder from "./ProcessFolder.vue"

const folders = ref([])
const treeFolders = ref([])

const folderModalOpen = ref(false)
const processModalOpen = ref(false)

const editingFolder = ref(false)

const folderForm = ref({
  id: null,
  name: "",
  parentId: null
})

const processForm = ref({
  name: "",
  folderId: null
})

const search = ref("")

const toast = ref({
  show: false,
  message: ""
})

const showToast = (msg) => {
  toast.value.message = msg
  toast.value.show = true
  setTimeout(() => toast.value.show = false, 2500)
}

const fetchFolders = async () => {

  try {

    const token = localStorage.getItem("jwt")

    const res = await api.get("/processes/registry", {
      headers: { Authorization: `Bearer ${token}` }
    })

    folders.value = res.data

    treeFolders.value = buildTreeSelect(res.data)

  } catch (err) {
    console.error(err)
  }

}

const buildTreeSelect = (folders, level = 0) => {

  let result = []

  folders.forEach(folder => {

    result.push({
      id: folder.id,
      label: `${"— ".repeat(level)}${folder.name}`
    })

    if (folder.children?.length) {

      result = result.concat(
        buildTreeSelect(folder.children, level + 1)
      )

    }

  })

  return result

}

const openCreateFolderModal = () => {

  editingFolder.value = false

  folderForm.value = {
    id: null,
    name: "",
    parentId: null
  }

  folderModalOpen.value = true

}

const openEditFolderModal = (folder) => {

  editingFolder.value = true

  folderForm.value = {
    id: folder.id,
    name: folder.name,
    parentId: folder.parentId || null
  }

  folderModalOpen.value = true

}

const openCreateProcessModal = () => {

  processForm.value = {
    name: "",
    folderId: treeFolders.value[0]?.id || null
  }

  processModalOpen.value = true

}

const saveFolder = async () => {

  if (!folderForm.value.name) {
    showToast("Введите название папки")
    return
  }

  try {

    const token = localStorage.getItem("jwt")

    if (editingFolder.value) {

      await api.put(
        `/process-folders/${folderForm.value.id}`,
        folderForm.value,
        { headers: { Authorization: `Bearer ${token}` } }
      )

      showToast("Папка обновлена")

    } else {

        const payload = {
        name: folderForm.value.name,
        parentId: folderForm.value.parentId
    }

      await api.post(
        "/process-folders",
        payload,
        { headers: { Authorization: `Bearer ${token}` } }
      )

      showToast("Папка создана")

    }

    folderModalOpen.value = false

    fetchFolders()

  } catch (err) {

    console.error(err)

    showToast("Ошибка сохранения папки")

  }

}

const saveProcess = async () => {

  if (!processForm.value.name || !processForm.value.folderId) {
    showToast("Введите название и выберите папку")
    return
  }

  try {

    const token = localStorage.getItem("jwt")

    const res = await api.post(
      "/processes",
      processForm.value,
      { headers: { Authorization: `Bearer ${token}` } }
    )

    showToast("Процесс создан")

    processModalOpen.value = false

    window.location.href = `/processes/${res.data.ID}`

  } catch (err) {

    console.error(err)

    showToast("Ошибка создания процесса")

  }

}

const deleteFolder = async (id) => {

  if (!confirm("Удалить папку?")) return

  try {

    const token = localStorage.getItem("jwt")

    await api.delete(`/process-folders/${id}`, {
      headers: { Authorization: `Bearer ${token}` }
    })

    showToast("Папка удалена")

    fetchFolders()

  } catch (err) {

    console.error(err)

    showToast("Ошибка удаления папки")

  }

}

const filteredFolders = computed(() => {

  if (!search.value) return folders.value

  const query = search.value.toLowerCase()

  const filterFolder = (folder) => {

    const nameMatch = folder.name.toLowerCase().includes(query)

    const processMatch = folder.processes?.some(p =>
      p.name.toLowerCase().includes(query)
    )

    const childrenMatch =
      folder.children?.map(filterFolder).filter(Boolean) || []

    if (nameMatch || processMatch || childrenMatch.length) {

      return {
        ...folder,
        children: childrenMatch
      }

    }

    return null

  }

  return folders.value.map(filterFolder).filter(Boolean)

})

const currentPage = ref(1)

const itemsPerPage = 10

const totalPages = computed(() =>
  Math.ceil(filteredFolders.value.length / itemsPerPage)
)

const paginatedFolders = computed(() => {

  const start = (currentPage.value - 1) * itemsPerPage

  return filteredFolders.value.slice(start, start + itemsPerPage)

})

const prevPage = () => {
  if (currentPage.value > 1) currentPage.value--
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) currentPage.value++
}

watch(search, () => {
  currentPage.value = 1
})

onMounted(fetchFolders)

</script>

<style scoped>  
.employees-page-container{margin-left:220px;padding:20px 40px;width:calc(95% - 220px);}
.employees-card{background:white;border-radius:12px;padding:25px;box-shadow:0 4px 20px rgba(0,0,0,0.08);min-height:520px;}
.employees-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:20px;}
.header-actions{display:flex;align-items:center;gap:12px;}
.search-box{display:flex;align-items:center;gap:6px;background:#f3f4f6;padding:6px 10px;border-radius:6px;}
.search-box input{border:none;background:transparent;outline:none;font-size:14px;width:200px;}
.add-btn{padding:6px 14px;background:#4f46e5;color:white;border:none;border-radius:6px;cursor:pointer;font-size:14px;}
.add-btn:hover{background:#3730a3;}
.employees-table{display:flex;flex-direction:column;gap:6px;}
.folder-block{margin-left:20px;border-left:1px solid #eee;padding-left:10px;margin-bottom:10px;}
.folder-header{display:flex;justify-content:space-between;align-items:center;font-weight:600;margin-bottom:4px;}
.process-row{padding:4px 0;}
.process-name{color:#4f46e5;text-decoration:none;}
.process-name:hover{text-decoration:underline;}
.delete-btn{padding:2px 8px;background:#e5e7eb;border:none;border-radius:4px;cursor:pointer;font-size:12px;}
.delete-btn:hover{background:#d1d5db;}
.pagination{display:flex;align-items:center;gap:10px;margin-top:12px;justify-content:flex-end;}
.pagination button{padding:6px 12px;border-radius:6px;border:1px solid #ccc;background:#f9f9f9;cursor:pointer;transition:0.2s;}
.pagination button:disabled{opacity:0.5;cursor:default;}
.pagination button:hover:not(:disabled){background:#e0e0e0;}

.modal-overlay{
  position:fixed;
  top:0;
  left:0;
  right:0;
  bottom:0;
  background:rgba(0,0,0,0.4);
  display:flex;
  justify-content:center;
  align-items:center;
}

.modal{
  background:white;
  padding:30px;
  border-radius:10px;
  width:320px;
  display:flex;
  flex-direction:column;
  gap:12px;
}

.modal.modal-wide{
  width: 600px;
  padding: 30px;
}

.modal label{
  font-weight:500;
  margin-bottom:4px;
}

.modal input[type="text"],
.modal input[type="date"],
.modal input[type="number"],
.modal select{
  padding:10px;
  border-radius:6px;
  border:1px solid #ccc;
  width:100%;
  box-sizing:border-box;
  font-size:14px;
}

.modal-columns{
  display:flex;
  gap:20px;
}

.modal-column{
  flex:1;
  display:flex;
  flex-direction:column;
  gap:12px;
}

.modal-actions{
  display:flex;
  justify-content:flex-end;
  gap:10px;
  margin-top:10px;
}

.save-btn{
  background:#4f46e5;
  color:white;
  border:none;
  padding:8px 14px;
  border-radius:6px;
  cursor:pointer;
}

.cancel-btn{
  background:#e5e7eb;
  border:none;
  padding:8px 14px;
  border-radius:6px;
  cursor:pointer;
}

.toast{
  position:fixed;
  bottom:30px;
  right:30px;
  background:#4f46e5;
  color:white;
  padding:12px 18px;
  border-radius:8px;
  box-shadow:0 4px 10px rgba(0,0,0,0.2);
}
</style>