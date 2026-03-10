<template>
  <div class="employees-page-container">
    <div class="employees-card">

      <div class="employees-header">
        <h2>Сотрудники</h2>

        <div class="header-actions">

          <div class="search-box">
            <Search :size="16" />
            <input v-model="search" placeholder="Поиск по ФИО..." />
          </div>

          <button class="add-btn" @click="openCreateModal">
            Добавить
          </button>

        </div>
      </div>

      <div class="employees-table">

        <div class="employees-row-header">
            <span class="employee-name-header">ФИО</span>
            <span class="employee-dept">Отдел</span>
            <span class="employee-pos">Должность</span>
            <span class="employee-remote">Удаленно</span>
            <span class="employee-hire">Дата приема</span>
            <span class="employee-fire">Дата увольнения</span>
        </div>

        <div
          v-for="employee in paginatedEmployees"
          :key="employee.ID"
          class="employee-row"
        >
          <router-link :to="`/employees/${employee.id}`" class="employee-name">
            {{ employee.last_name }} {{ employee.first_name }} {{ employee.middle_name }}
          </router-link>
          <span class="employee-dept">{{ employee.department }}</span>
          <span class="employee-pos">{{ employee.position }}</span>
          <span class="employee-remote">{{ employee.is_remote ? "Да" : "Нет" }}</span>
          <span class="employee-hire">{{ formatDate(employee.hire_date) }}</span>
          <span class="employee-fire">{{ employee.fire_date ? formatDate(employee.fire_date) : "" }}</span>
        </div>

        <div class="pagination" v-if="totalPages > 1">
          <button :disabled="currentPage === 1" @click="prevPage">◀</button>
          <span>{{ currentPage }} из {{ totalPages }}</span>
          <button :disabled="currentPage === totalPages" @click="nextPage">▶</button>
        </div>

      </div>
    </div>

    <!-- Модалка добавления -->
    <div v-if="modalOpen" class="modal-overlay">
    <div class="modal modal-wide">
        <h3>Добавить сотрудника</h3>

        <div class="modal-columns">

        <div class="modal-column">
            <label>Фамилия</label>
            <input v-model="form.last_name" type="text" />

            <label>Имя</label>
            <input v-model="form.first_name" type="text" />

            <label>Отчество</label>
            <input v-model="form.middle_name" type="text" />

            <label>Отдел</label>
            <select v-model="form.department_id">
            <option value="" disabled>Выберите отдел</option>
            <option v-for="d in departments" :key="d.ID" :value="d.ID">
                {{ d.Name }}
            </option>
            </select>

            <label>Должность</label>
            <select v-model="form.position_id">
            <option value="" disabled>Выберите должность</option>
            <option v-for="p in positions" :key="p.ID" :value="p.ID">
                {{ p.Name }}
            </option>
            </select>
        </div>

        <div class="modal-column">
            <label>Дата рождения</label>
            <input v-model="form.birth_date" type="date" />

            <label>Дата приема</label>
            <input v-model="form.hire_date" type="date" />

            <label>Ставка в час</label>
            <input v-model.number="form.salary" type="number" step="0.01" />

            <label class="checkbox-label">
            <input type="checkbox" v-model="form.is_remote" /> Удаленно
            </label>
        </div>

        </div>

        <div class="modal-actions">
        <button @click="saveEmployee" class="save-btn">Сохранить</button>
        <button @click="modalOpen=false" class="cancel-btn">Отмена</button>
        </div>
    </div>
    </div>

    <!-- Toast -->
    <div v-if="toast.show" class="toast">{{ toast.message }}</div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from "vue"
import { Search } from "lucide-vue-next"
import api from "../api/axios"

const employees = ref([])
const search = ref("")
const modalOpen = ref(false)

const openCreateModal = () => {
  form.value = {
    last_name: "",
    first_name: "",
    middle_name: "",
    department_id: "",
    position_id: "",
    birth_date: "",
    hire_date: "",
    is_remote: false,
    salary: 0,
  }
  modalOpen.value = true
}
const form = ref({
  last_name: "",
  first_name: "",
  middle_name: "",
  department_id: "",
  position_id: "",
  birth_date: "",
  hire_date: "",
  is_remote: false,
  salary: 0,
})

const departments = ref([])
const positions = ref([])

const fetchDictionaries = async () => {
  const token = localStorage.getItem("jwt")
  try {
    const res = await api.get("/dict", { headers: { Authorization: `Bearer ${token}` } })
    departments.value = res.data.departments || []
    positions.value = res.data.positions || []
  } catch (err) {
    console.error("Ошибка загрузки справочников:", err)
  }
}
const toast = ref({ show: false, message: "" })

const showToast = (msg) => {
  toast.value.message = msg
  toast.value.show = true
  setTimeout(() => (toast.value.show = false), 2500)
}

const currentPage = ref(1)
const itemsPerPage = 10

const fetchEmployees = async () => {
  try {
    const token = localStorage.getItem("jwt")
    const res = await api.get("/employees", { headers: { Authorization: `Bearer ${token}` } })
    employees.value = res.data
  } catch (err) {
    console.error("Ошибка загрузки сотрудников:", err)
  }
}

const filteredEmployees = computed(() => {
  if (!search.value) return employees.value
  const q = search.value.toLowerCase()
  return employees.value.filter(e =>
    `${e.last_name} ${e.first_name} ${e.middle_name}`.toLowerCase().includes(q)
  )
})

const totalPages = computed(() => Math.ceil(filteredEmployees.value.length / itemsPerPage))

const paginatedEmployees = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage
  return filteredEmployees.value.slice(start, start + itemsPerPage)
})

const prevPage = () => { if (currentPage.value > 1) currentPage.value-- }
const nextPage = () => { if (currentPage.value < totalPages.value) currentPage.value++ }

watch(search, () => { currentPage.value = 1 })

const saveEmployee = async () => {
  const token = localStorage.getItem("jwt")

  const requiredFields = ["last_name","first_name","middle_name","department_id","position_id","birth_date","hire_date","salary"]
  for(const field of requiredFields){
    if(!form.value[field] && form.value[field] !== 0){
      showToast("Заполните все поля")
      return
    }
  }

  try {
    await api.post("/employees", form.value, { headers: { Authorization: `Bearer ${token}` } })
    showToast("Сотрудник добавлен")
    modalOpen.value = false
    fetchEmployees()
  } catch (err) {
    showToast("Ошибка добавления")
    console.error(err)
  }
}

onMounted(fetchDictionaries)

const formatDate = (date) => {
  if (!date) return ""
  return new Date(date).toLocaleDateString()
}

onMounted(fetchEmployees)
</script>

<style scoped>
.employees-page-container{
  margin-left:220px;
  padding:20px 40px;
  width:calc(95% - 220px);
}

.employees-card{
  background:white;
  border-radius:12px;
  padding:25px;
  box-shadow:0 4px 20px rgba(0,0,0,0.08);
  min-height:520px;
}

.employees-header{
  display:flex;
  justify-content:space-between;
  align-items:center;
  margin-bottom:20px;
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
  width:180px;
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

.employees-table{
  display:flex;
  flex-direction:column;
  gap:6px;
}

.employees-row-header,
.employee-row{
  display:flex;
  align-items:center;
  gap:10px;
  padding:8px 0;
  border-bottom:1px solid #eee;
  text-decoration:none;
}

.employees-row-header{
  font-weight:600;
  border-bottom:2px solid #eee;
}

.employee-name{
  flex-basis: 350px;
  text-decoration:none;
  color:var(--color-primary);
  font-weight:500;
}

.employee-name-header{
  flex-basis: 350px;
  text-align:left;
}

.employee-name:hover{
  text-decoration:underline;
}

.employee-dept{
  flex-basis: 250px;
  text-align:left;
}

.employee-pos{
  flex-basis: 250px;
  text-align:left;
}

.employee-remote{
  flex-basis: 80px;
  text-align:center;
}

.employee-hire,
.employee-fire{
  flex-basis: 140px;
  text-align:center;
}

.employee-name,
.employee-dept,
.employee-pos{
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.checkbox-label{
  display:flex;
  align-items:center;
  gap:6px;
  font-weight:500;
  margin-top:8px;
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
