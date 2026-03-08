<template>
  <div class="employee-page-container">
    <div class="employee-card">

      <button class="back-btn" @click="$router.push('/processes')">← Назад</button>

      <h2 class="employee-title">{{ process.Name }}</h2>

      <div class="tabs">
        <button 
          :class="{ active: activeTab === 'data' }" 
          @click="activeTab = 'data'">Данные</button>
        <button 
          v-for="(version, index) in process.Versions"
          :key="version.ID"
          :class="{ active: activeTab === 'version' + index }"
          @click="activeTab = 'version' + index"
        >
          Версия {{ version.Version }}
        </button>
      </div>

      <!-- Данные процесса -->
      <div v-if="activeTab === 'data'" class="tab-content">
        <div class="columns">
          <div class="column">
            <div class="form-group">
              <label>Название процесса</label>
              <input v-model="process.Name" type="text" />
            </div>

            <div class="form-group">
            <label>Владелец процесса</label>
            <select v-model="selectedOwner">
                <option value="" disabled>Выберите сотрудника</option>
                <option v-for="e in employees" :key="e.id" :value="e.id.toString()">
                {{ formatFullName(e) }}
                </option>
            </select>
            </div>

            <div class="form-group checkbox-group">
              <label>
                <input type="checkbox" v-model="process.IsActive" />
                Актуальный
              </label>
            </div>

            <div class="form-group">
              <label>Дата создания</label>
              <input type="text" :value="formatDateTime(process.CreatedAt)" disabled />
            </div>
          </div>

          <div class="column">
            <div class="form-group">
              <label>Описание</label>
              <textarea v-model="process.Description" rows="5"></textarea>
            </div>

            <div class="form-group">
              <label>Регламенты</label>
              <textarea v-model="process.Regulations" rows="5"></textarea>
            </div>
          </div>
        </div>

        <button class="save-btn" @click="saveProcess">Сохранить</button>
      </div>

      <!-- Вкладки с версиями -->
      <div 
        v-for="(version, index) in process.Versions" 
        :key="version.ID"
        v-if="activeTab === 'version' + index"
        class="tab-content"
      >
        <h3>Версия {{ version.Version }}</h3>
        <p><b>Опубликована:</b> {{ version.IsPublished ? 'Да' : 'Нет' }}</p>
        <p><b>Создана:</b> {{ formatDateTime(version.CreatedAt) }}</p>

        <div v-if="version.Steps.length === 0">Шагов нет</div>
        <ul v-else>
          <li v-for="step in version.Steps" :key="step.ID">{{ step.Name }}</li>
        </ul>
      </div>

    </div>

    <div v-if="toast.show" class="toast">{{ toast.message }}</div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue"
import { useRoute } from "vue-router"
import api from "../api/axios"

const route = useRoute()
const process = ref({
  Name: '',
  Description: '',
  Regulations: '',
  OwnerID: null,
  Versions: [],
  CreatedAt: ''
})

const employees = ref([]) 
const selectedOwner = ref(null)

const activeTab = ref('data')

const toast = ref({ show: false, message: "" })
const showToast = (msg) => {
  toast.value.message = msg
  toast.value.show = true
  setTimeout(() => (toast.value.show = false), 2500)
}

const fetchProcess = async () => {
  const token = localStorage.getItem("jwt")
  try {
    const res = await api.get(`/processes/${route.params.id}`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    process.value = res.data

    selectedOwner.value = process.value.OwnerID ? process.value.OwnerID.toString() : ''
  } catch (err) {
    console.error(err)
  }
}

const fetchEmployees = async () => {
  const token = localStorage.getItem("jwt")
  try {
    const res = await api.get("/employees", {
      headers: { Authorization: `Bearer ${token}` }
    })
    employees.value = res.data
  } catch (err) {
    console.error("Ошибка загрузки сотрудников:", err)
  }
}

const formatFullName = (e) => {
  return `${e.last_name} ${e.first_name} ${e.middle_name}`
}

const saveProcess = async () => {
  const token = localStorage.getItem("jwt")
  try {
    await api.put(`/processes/${process.value.ID}`, {
      ...process.value,
      OwnerID: selectedOwner.value ? Number(selectedOwner.value) : null,
      IsActive: process.value.IsActive
    }, { headers: { Authorization: `Bearer ${token}` } })
    showToast("Процесс сохранен")
  } catch (err) {
    console.error("Ошибка сохранения процесса:", err)
    showToast("Ошибка при сохранении процесса")
  }
}

const formatDateTime = (iso) => {
  if (!iso) return ''
  const d = new Date(iso)
  const day = d.getDate().toString().padStart(2, '0')
  const month = (d.getMonth() + 1).toString().padStart(2, '0')
  const year = d.getFullYear()
  const hours = d.getHours().toString().padStart(2, '0')
  const minutes = d.getMinutes().toString().padStart(2, '0')
  return `${day}.${month}.${year} ${hours}:${minutes}`
}

onMounted(async () => {
  await fetchEmployees()
  await fetchProcess()
  selectedOwner.value = process.OwnerID
})
</script>

<style scoped>
.employee-page-container {
  margin-left: 240px;
  padding: 40px;
  width: calc(100% - 240px);
  display: flex;
  justify-content: center;
}

.employee-card {
  background-color: #fff;
  border-radius: 12px;
  padding: 30px;
  width: 100%;
  max-width: 1000px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.08);
  display: flex;
  flex-direction: column;
  gap: 20px;
  position: relative;
}

.back-btn {
  position: absolute;
  top: 20px;
  left: 20px;
  padding: 6px 12px;
  background: #f3f4f6;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: 0.2s;
}

.back-btn:hover {
  background: #e0e0e0;
}

.employee-title {
  font-size: 28px;
  margin-bottom: 20px;
  color: #333;
  text-align: center;
}

.tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
  justify-content: center;
}

.tabs button {
  padding: 10px 18px;
  border: none;
  background: #f0f0f0;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: 0.2s;
}

.tabs button.active {
  background: #4f46e5;
  color: white;
}

.tabs button:hover:not(.active) {
  background: #e0e0e0;
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.columns {
  display: flex;
  gap: 40px;
}

.column {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-group label {
  font-weight: 500;
  margin-bottom: 4px;
  color: #555;
}

.form-group input[type="text"],
.form-group select,
.form-group textarea {
  padding: 10px;
  border-radius: 6px;
  border: 1px solid #ccc;
  transition: border-color 0.2s;
  font-size: 14px;
  font-family: 'Inter', sans-serif;
}

.form-group input[type="text"]:focus,
.form-group select:focus,
.form-group textarea:focus {
  border-color: #4f46e5;
  outline: none;
}

.checkbox-group input[type="checkbox"] {
  width: 20px;
  height: 20px;
  margin-right: 8px;
  vertical-align: middle;
}

.save-btn {
  padding: 12px 20px;
  background-color: #4f46e5;
  color: white;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  align-self: flex-start;
  font-weight: 500;
  transition: 0.2s;
}

.save-btn:hover {
  background-color: #3730a3;
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