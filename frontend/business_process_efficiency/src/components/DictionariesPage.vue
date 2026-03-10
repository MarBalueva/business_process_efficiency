<template>
  <div class="dict-page-container">
    <div class="dict-card">

      <div class="dict-layout">

        <!-- Список справочников -->
        <div class="dict-sidebar">
          <h3>Справочники</h3>

          <div
            v-for="dict in dictionaries"
            :key="dict.key"
            class="dict-item"
            :class="{active: activeDict === dict.key}"
            @click="selectDictionary(dict)"
          >
            {{ dict.name }}
          </div>
        </div>

        <!-- Содержимое -->
        <div class="dict-content">

          <div class="dict-header">

            <h3>{{ activeDictName }}</h3>

            <div class="header-actions">

                <div class="search-box">
                <Search :size="16" />
                <input
                    v-model="search"
                    placeholder="Поиск..."
                />
                </div>

                <button class="add-btn" @click="openCreateModal">
                Добавить
                </button>

            </div>

            </div>

          <div class="dict-table">

            <div
            v-for="item in paginatedItems"
            :key="item.ID"
            class="dict-row"
            >
              <span class="dict-name">{{ item.Name }}</span>

              <div class="dict-actions">
                <button class="icon-btn edit" @click="openEditModal(item)">
                  <Pencil :size="16"/>
                </button>

                <button class="icon-btn delete" @click="openDeleteDialog(item.ID)">
                  <Trash :size="16"/>
                </button>
              </div>
            </div>

          </div>
          <div class="pagination" v-if="currentItems.length > itemsPerPage">
            <button 
                :disabled="currentPage === 1"
                @click="currentPage--"
            >
                ◀
            </button>

            <span>{{ currentPage }} из {{ totalPages }}</span>

            <button 
                :disabled="currentPage === totalPages"
                @click="currentPage++"
            >
                ▶
            </button>
            </div>

        </div>

      </div>
    </div>

    <!-- МОДАЛКА -->
    <div v-if="modalOpen" class="modal-overlay">
      <div class="modal">

        <h3>{{ editMode ? "Редактировать" : "Добавить" }}</h3>

        <label>Наименование</label>
        <input
          v-model="form.name"
          placeholder="Название"
        />

        <label>Код</label>
        <input
          v-model="form.code"
          placeholder="Код"
        />

        <div class="modal-actions">
          <button @click="saveItem" class="save-btn">Сохранить</button>
          <button @click="modalOpen=false" class="cancel-btn">Отмена</button>
        </div>

      </div>
    </div>

    <div v-if="showDeleteDialog" class="modal-overlay">
      <div class="modal">
        <h3>Удалить запись?</h3>
        <p>Вы действительно хотите удалить выбранную запись?</p>
        <div class="modal-actions">
          <button class="cancel-btn" @click="cancelDelete">Отмена</button>
          <button class="save-btn" @click="confirmDelete">Удалить</button>
        </div>
      </div>
    </div>

    <!-- TOAST -->
    <div v-if="toast.show" class="toast">
      {{ toast.message }}
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch  } from "vue"
import api from "../api/axios"
import { Pencil, Trash, Search } from "lucide-vue-next"


const dictionaries = [
  { key: "departments", name: "Отделы" },
  { key: "positions", name: "Должности" },
  { key: "access_groups", name: "Группы доступа" }
]

const activeDict = ref(null)
const items = ref({})

const modalOpen = ref(false)
const editMode = ref(false)
const search = ref("")
const showDeleteDialog = ref(false)
const deletingId = ref(null)

const currentPage = ref(1) 
const itemsPerPage = 10

const paginatedItems = computed(() => {
  const list = items.value[activeDict.value] || []

  const filtered = search.value
    ? list.filter(i =>
        i.Name?.toLowerCase().includes(search.value.toLowerCase()) ||
        i.Code?.toLowerCase().includes(search.value.toLowerCase())
      )
    : list

  const start = (currentPage.value - 1) * itemsPerPage
  const end = start + itemsPerPage

  return filtered.slice(start, end)
})

const totalPages = computed(() => {
  const list = items.value[activeDict.value] || []
  const filtered = search.value
    ? list.filter(i =>
        i.Name?.toLowerCase().includes(search.value.toLowerCase()) ||
        i.Code?.toLowerCase().includes(search.value.toLowerCase())
      )
    : list
  return Math.ceil(filtered.length / itemsPerPage)
})

const form = ref({
  id:null,
  name:"",
  code:""
})

const toast = ref({
  show:false,
  message:""
})

const showToast = (msg)=>{
  toast.value.message = msg
  toast.value.show = true

  setTimeout(()=>{
    toast.value.show = false
  },2500)
}

const selectDictionary = (dict) => {
  activeDict.value = dict.key
  search.value = ""
  currentPage.value = 1
}

watch(search, () => {
  currentPage.value = 1
})

const activeDictName = computed(()=>{
  const d = dictionaries.find(d=>d.key === activeDict.value)
  return d ? d.name : ""
})

const currentItems = computed(()=>{

  const list = items.value[activeDict.value] || []

  if(!search.value) return list

  const q = search.value.toLowerCase()

  return list.filter(i =>
    i.Name?.toLowerCase().includes(q) ||
    i.Code?.toLowerCase().includes(q)
  )
})

const fetchDictionaries = async ()=>{
  const token = localStorage.getItem("jwt")

  const res = await api.get("/dict",{
    headers:{ Authorization:`Bearer ${token}` }
  })

  items.value = res.data
}

const openCreateModal = ()=>{
  editMode.value = false
  form.value = {id:null,name:"",code:""}
  modalOpen.value = true
}

const openEditModal = (item)=>{
  editMode.value = true

  form.value = {
    id:item.ID,
    name:item.Name,
    code:item.Code
  }

  modalOpen.value = true
}

const saveItem = async ()=>{
  const token = localStorage.getItem("jwt")

  try{

    if(editMode.value){

      await api.put(`/dict/${activeDict.value}/${form.value.id}`,{
        name:form.value.name,
        code:form.value.code
      },{
        headers:{Authorization:`Bearer ${token}`}
      })

      showToast("Запись обновлена")

    }else{

      await api.post(`/dict/${activeDict.value}`,{
        name:form.value.name,
        code:form.value.code
      },{
        headers:{Authorization:`Bearer ${token}`}
      })

      showToast("Запись создана")
    }

    modalOpen.value=false
    fetchDictionaries()

  }catch(e){
    showToast("Ошибка сохранения")
  }
}

const openDeleteDialog = (id) => {
  deletingId.value = id
  showDeleteDialog.value = true
}

const cancelDelete = () => {
  deletingId.value = null
  showDeleteDialog.value = false
}

const confirmDelete = async () => {
  if (!deletingId.value) return

  const token = localStorage.getItem("jwt")

  try {
    await api.delete(`/dict/${activeDict.value}/${deletingId.value}`, {
      headers: { Authorization: `Bearer ${token}` }
    })

    showToast("������ �������")
    cancelDelete()
    fetchDictionaries()
  } catch (e) {
    showToast("������ ��������")
  }
}

onMounted(()=>{
  activeDict.value = dictionaries[0].key
  fetchDictionaries()
})
</script>

<style scoped>

.dict-page-container{
  margin-left:220px;
  padding:20px 40px;
  width:calc(95% - 220px);
}

.dict-card{
  background:white;
  border-radius:12px;
  padding:25px;
  box-shadow:0 4px 20px rgba(0,0,0,0.08);
  min-height:520px;
}

.dict-layout{
  display:flex;
  gap:30px;
}

.dict-sidebar{
  width:200px;
  border-right:1px solid #eee;
}

.dict-sidebar h3{
  margin-bottom:15px;
}

.dict-item{
  padding:10px;
  border-radius:6px;
  cursor:pointer;
  transition:.2s;
}

.dict-item:hover{
  background:#f5f5f5;
}

.dict-item.active{
  background:var(--color-primary);
  color:white;
}

.dict-content{
  flex:1;
}

.dict-header{
  display:flex;
  justify-content:space-between;
  align-items:center;
  margin-bottom:20px;
}

.add-btn{
  padding:6px 14px;
  background:var(--color-primary);
  color:white;
  border:none;
  border-radius:6px;
  cursor:pointer;
  font-size:14px;
}

.add-btn:hover{
  background:var(--color-primary-hover);
}

.dict-row{
  display:flex;
  justify-content:space-between;
  align-items:center;
  padding:10px 0;
  border-bottom:1px solid #eee;
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

.modal input{
  padding:8px;
  border-radius:6px;
  border:1px solid #ccc;
}

.modal-actions{
  display:flex;
  justify-content:flex-end;
  gap:10px;
  margin-top:10px;
}

.save-btn{
  background:var(--color-primary);
  color:white;
  border:none;
  padding:8px 14px;
  border-radius:6px;
  cursor:pointer;
}

.cancel-btn{
  background:var(--color-muted-bg);
  border:none;
  padding:8px 14px;
  border-radius:6px;
  cursor:pointer;
}

.toast{
  position:fixed;
  bottom:30px;
  right:30px;
  background:var(--color-primary);
  color:white;
  padding:12px 18px;
  border-radius:8px;
  box-shadow:0 4px 10px rgba(0,0,0,0.2);
}

.header-actions{
  display:flex;
  align-items:center;
  gap:12px;
}

.search-box{
  display:flex;
  align-items:center;
  gap:6px;
  background:var(--color-soft-bg);
  padding:6px 10px;
  border-radius:6px;
}

.search-box input{
  border:none;
  background:transparent;
  outline:none;
  font-size:14px;
  width:150px;
}

.pagination {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 12px;
  justify-content: flex-end;
}

.pagination button {
  padding: 6px 12px;
  border-radius: 6px;
  border: 1px solid #ccc;
  background: #f9f9f9;
  cursor: pointer;
  transition: 0.2s;
}

.pagination button:disabled {
  opacity: 0.5;
  cursor: default;
}

.pagination button:hover:not(:disabled) {
  background: #e0e0e0;
}
</style>


