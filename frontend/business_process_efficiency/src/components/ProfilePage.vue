<template>
  <div class="employee-page-container">
    <div class="employee-card">

      <h2 class="employee-title">{{ fullName }}</h2>

      <div class="tab-content profile-tab">
        <div class="form-group">
          <label>Логин</label>
          <input v-model="profile.login" type="text" disabled>
        </div>

        <div class="form-group">
          <label>Фамилия</label>
          <input v-model="profile.last_name" type="text">
        </div>

        <div class="form-group">
          <label>Имя</label>
          <input v-model="profile.first_name" type="text">
        </div>

        <div class="form-group">
          <label>Отчество</label>
          <input v-model="profile.middle_name" type="text">
        </div>

        <div class="form-group">
          <label>Новый пароль</label>
          <input v-model="profile.password" type="password" placeholder="Оставьте пустым, если не менять">
        </div>

        <button class="save-btn" @click="updateProfile">Сохранить</button>
      </div>

      <div v-if="toast.show" class="toast">{{ toast.message }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue"
import api from "../api/axios"

const profile = ref({
  login: "",
  last_name: "",
  first_name: "",
  middle_name: "",
  password: ""
})

const toast = ref({ show: false, message: "" })
const showToast = (msg) => {
  toast.value.message = msg
  toast.value.show = true
  setTimeout(() => (toast.value.show = false), 2500)
}

const fetchProfile = async () => {
  try {
    const token = localStorage.getItem("jwt")
    const res = await api.get("/profile/me", {
      headers: { Authorization: `Bearer ${token}` }
    })

    profile.value.login = res.data.login
    profile.value.last_name = res.data.employee.last_name
    profile.value.first_name = res.data.employee.first_name
    profile.value.middle_name = res.data.employee.middle_name
    profile.value.password = ""
  } catch (err) {
    console.error("Ошибка загрузки профиля:", err)
    showToast("Не удалось загрузить профиль")
  }
}

// Обновление профиля
const updateProfile = async () => {
  try {
    const token = localStorage.getItem("jwt")
    const payload = {
      last_name: profile.value.last_name,
      first_name: profile.value.first_name,
      middle_name: profile.value.middle_name,
      password: profile.value.password || undefined
    }

    await api.put("/profile/me", payload, {
      headers: { Authorization: `Bearer ${token}` }
    })

    showToast("Профиль обновлен")
    profile.value.password = ""
  } catch (err) {
    console.error("Ошибка обновления профиля:", err)
    showToast("Не удалось обновить профиль")
  }
}

const fullName = computed(() => {
  return `${profile.value.last_name || ''} ${profile.value.first_name || ''} ${profile.value.middle_name || ''}`
})

onMounted(() => {
  fetchProfile()
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
  max-width: 800px;
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
  background: var(--color-soft-bg);
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

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
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

.form-group input {
  padding: 10px;
  border-radius: 6px;
  max-width: 400px;
  border: 1px solid #ccc;
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-group input:focus {
  border-color: var(--color-primary);
  outline: none;
}

.save-btn {
  padding: 12px 20px;
  background-color: var(--color-primary);
  color: white;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-weight: 500;
  transition: 0.2s;
  align-self: flex-start;
}

.save-btn:hover {
  background-color: var(--color-primary-hover);
}

.toast {
  position: fixed;
  bottom: 30px;
  right: 30px;
  background: var(--color-primary);
  color: white;
  padding: 12px 18px;
  border-radius: 8px;
  box-shadow: 0 4px 10px rgba(0,0,0,0.2);
}
</style>
