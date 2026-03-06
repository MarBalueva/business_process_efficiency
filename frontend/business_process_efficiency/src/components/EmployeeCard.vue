<template>
  <div class="employee-page-container">
    <div class="employee-card">
      <h2 class="employee-title">{{ fullName }}</h2>

      <div class="tabs">
        <button 
          :class="{ active: activeTab === 'data' }" 
          @click="activeTab = 'data'">Данные</button>
        <button 
          :class="{ active: activeTab === 'user' }" 
          @click="activeTab = 'user'">Пользователь</button>
      </div>

      <div v-if="activeTab === 'data'" class="tab-content">
        <div class="columns">
          <div class="column">
            <div class="form-group">
              <label>Фамилия</label>
              <input v-model="employee.last_name" type="text" />
            </div>
            <div class="form-group">
              <label>Имя</label>
              <input v-model="employee.first_name" type="text" />
            </div>
            <div class="form-group">
              <label>Отчество</label>
              <input v-model="employee.middle_name" type="text" />
            </div>
            <div class="form-group">
              <label>Отдел</label>
              <input v-model="employee.department" type="text" />
            </div>
            <div class="form-group">
              <label>Должность</label>
              <input v-model="employee.position" type="text" />
            </div>
          </div>

          <div class="column">
            <div class="form-group">
              <label>Удаленно</label>
              <input type="checkbox" v-model="employee.is_remote" />
            </div>
            <div class="form-group">
              <label>Дата рождения</label>
              <input type="date" v-model="employee.birth_date" />
            </div>
            <div class="form-group">
              <label>Дата приема</label>
              <input type="date" v-model="employee.hire_date" />
            </div>
            <div class="form-group">
              <label>Дата увольнения</label>
              <input type="date" v-model="employee.fire_date" />
            </div>
            <div class="form-group">
              <label>Зарплата</label>
              <input type="number" v-model="employee.salary" />
            </div>
          </div>
        </div>
        <button class="save-btn" @click="saveEmployee">Сохранить</button>
      </div>

      <div v-if="activeTab === 'user'" class="tab-content">
        <p>Здесь будут данные пользователя</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue"
import { useRoute } from "vue-router"
import api from "../api/axios"

const route = useRoute()
const employee = ref({})
const activeTab = ref("data")

const fetchEmployee = async () => {
  try {
    const token = localStorage.getItem("jwt")
    const res = await api.get(`/employees/${route.params.id}`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    employee.value = {
    ...res.data,
    birth_date: formatForInput(res.data.birth_date),
    hire_date: formatForInput(res.data.hire_date),
    fire_date: res.data.fire_date ? formatForInput(res.data.fire_date) : '', 
    created_at: formatForInput(res.data.created_at)
    }
  } catch (err) {
    console.error("Ошибка загрузки сотрудника:", err)
  }
}

const formatForInput = (isoString) => {
  if (!isoString) return ''
  const d = new Date(isoString)
  const month = (d.getMonth() + 1).toString().padStart(2, '0')
  const day = d.getDate().toString().padStart(2, '0')
  return `${d.getFullYear()}-${month}-${day}`
}

const saveEmployee = async () => {
  try {
    const token = localStorage.getItem("jwt")

    const payload = {
    ...employee.value,
    birth_date: employee.value.birth_date || null,
    hire_date: employee.value.hire_date || null,
    fire_date: employee.value.fire_date || null,
    }

    await api.put(`/employees/${route.params.id}`, payload, {
      headers: { Authorization: `Bearer ${token}` }
    })
    alert("Сотрудник сохранен")
  } catch (err) {
    console.error("Ошибка сохранения:", err)
    alert("Ошибка при сохранении")
  }
}

onMounted(fetchEmployee)

const fullName = computed(() => {
  return `${employee.value.last_name || ''} ${employee.value.first_name || ''} ${employee.value.middle_name || ''}`
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
}

.employee-title {
  font-size: 28px;
  margin-bottom: 20px;
  color: #333;
}

.tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
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
.form-group input[type="date"],
.form-group input[type="number"] {
  padding: 8px 12px;
  border-radius: 6px;
  border: 1px solid #ccc;
  transition: border-color 0.2s;
}

.form-group input[type="text"]:focus,
.form-group input[type="date"]:focus,
.form-group input[type="number"]:focus {
  border-color: #4f46e5;
  outline: none;
}

.form-group input[type="checkbox"] {
  width: 20px;
  height: 20px;
  margin-top: 4px;
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
</style>