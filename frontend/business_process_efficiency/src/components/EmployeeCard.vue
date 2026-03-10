<template>
  <div class="employee-page-container">
    <div class="employee-card">

      <button class="back-btn" @click="$router.push('/employees')">← Назад</button>
      <button class="delete-top-btn" @click="openDeleteEmployeeDialog">Удалить</button>

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
              <select v-model="employee.department_id">
                <option value="" disabled>Выберите отдел</option>
                <option v-for="d in departments" :key="d.ID" :value="d.ID">
                  {{ d.Name }}
                </option>
              </select>
            </div>
            <div class="form-group">
              <label>Должность</label>
              <select v-model="employee.position_id">
                <option value="" disabled>Выберите должность</option>
                <option v-for="p in positions" :key="p.ID" :value="p.ID">
                  {{ p.Name }}
                </option>
              </select>
            </div>
          </div>

          <div class="column">
            <div class="form-group checkbox-group">
              <label>
                <input type="checkbox" v-model="employee.is_remote" />
                Удаленно
              </label>
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
              <label>Ставка в час</label>
              <input type="number" v-model="employee.salary" />
            </div>
          </div>
        </div>

        <button class="save-btn" @click="saveEmployee">Сохранить</button>
      </div>

    <div v-if="activeTab === 'user'" class="tab-content user-tab">

    <div v-if="!user">

        <p class="no-user">У сотрудника нет пользователя</p>

        <button class="save-btn" @click="showCreateUserModal = true">
        Добавить пользователя
        </button>

    </div>

    <div v-else class="user-info">

        <div class="user-header">
        <div class="user-login">
            <b>Логин:</b> {{ user.Login }}
        </div>

        <div class="user-actions">
          <button class="icon-btn edit" @click="showEditUserModal = true" title="Редактировать пользователя">
            <Pencil :size="16"/>
          </button>
          <button class="icon-btn delete" @click="openDeleteUserDialog" title="Удалить пользователя">
            <Trash :size="16"/>
          </button>
        </div>
        </div>

        <div class="access-groups">

        <div class="groups-header">
            <b>Группы доступа</b>

            <button class="icon-btn edit" @click="showAddGroupModal = true" title="Редактировать группы доступа">
            <Pencil :size="16"/>
            </button>
        </div>

        <div v-if="userGroups.length === 0" class="no-groups">
            Нет групп доступа
        </div>

        <div 
            v-for="g in userGroups"
            :key="g.AccessGroupID"
            class="group-row"
        >
            <span>{{ g.Name }}</span>

            <button class="icon-btn delete" @click="openDeleteGroupDialog(g.AccessGroupID)">
              <Trash :size="16"/>
            </button>
        </div>

        </div>

  </div>

</div>

    </div>
    <div v-if="showCreateUserModal" class="modal-overlay">

  <div class="modal">

    <h3>Создать пользователя</h3>

    <div class="form-group">
      <label>Логин</label>
      <input v-model="newUser.login" type="text"> 
    </div>

    <div class="form-group">
      <label>Пароль</label>
      <input v-model="newUser.password" type="password">
    </div>

    <div class="modal-actions">
      <button class="save-btn" @click="createUser">Сохранить</button>
      <button class="cancel-btn" @click="showCreateUserModal=false">Отмена</button>
    </div>

  </div>

</div>

<div v-if="showEditUserModal" class="modal-overlay">

  <div class="modal">

    <h3>Изменить пользователя</h3>

    <div class="form-group">
      <label>Логин</label>
      <input v-model="editUser.login" type="text">
    </div>

    <div class="form-group">
      <label>Пароль</label>
      <input v-model="editUser.password" type="password">
    </div>

    <div class="modal-actions">
      <button class="save-btn" @click="updateUser">Сохранить</button>
      <button class="cancel-btn" @click="showEditUserModal=false">Отмена</button>
    </div>

  </div>

</div>

<div v-if="showAddGroupModal" class="modal-overlay">

  <div class="modal">

    <h3>Добавить группу доступа</h3>

    <div class="form-group">
      <label>Группа</label>

      <select v-model="selectedGroup">
        <option value="">Выберите группу</option>

        <option
          v-for="g in accessGroups"
          :key="g.ID"
          :value="g.ID"
        >
          {{ g.Name }}
        </option>

      </select>

    </div>

    <div class="modal-actions">
      <button class="save-btn" @click="addGroup">Добавить</button>
      <button class="cancel-btn" @click="showAddGroupModal=false">Отмена</button>
    </div>

  </div>

 </div>

<div v-if="showDeleteDialog" class="modal-overlay">
  <div class="modal">
    <h3>{{ deleteDialog.title }}</h3>
    <p>{{ deleteDialog.message }}</p>
    <div class="modal-actions">
      <button class="cancel-btn" @click="cancelDeleteDialog">Отмена</button>
      <button class="save-btn" @click="confirmDeleteDialog">Удалить</button>
    </div>
  </div>
</div>
    <div v-if="toast.show" class="toast">{{ toast.message }}</div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue"
import { useRoute } from "vue-router"
import { Pencil, Trash } from "lucide-vue-next"
import api from "../api/axios"

const route = useRoute()
const employee = ref({})
const activeTab = ref("data")
const departments = ref([])
const positions = ref([])

const user = ref(null)
const userGroups = ref([])
const accessGroups = ref([])

const showCreateUserModal = ref(false)
const showEditUserModal = ref(false)
const showAddGroupModal = ref(false)
const showDeleteDialog = ref(false)
const deleteDialog = ref({
  title: "",
  message: "",
  actionType: "",
  groupId: null
})

const selectedGroup = ref("")

const newUser = ref({
  login: "",
  password: ""
})

const editUser = ref({
  login: "",
  password: ""
})

const toast = ref({ show: false, message: "" })

const showToast = (msg) => {
  toast.value.message = msg
  toast.value.show = true
  setTimeout(() => (toast.value.show = false), 2500)
}

const fetchDictionaries = async () => {
  const token = localStorage.getItem("jwt")
  try {
    const res = await api.get("/dict", { headers: { Authorization: `Bearer ${token}` } })
    departments.value = res.data.departments || []
    positions.value = res.data.positions || []
    accessGroups.value = res.data.access_groups || []
  } catch (err) {
    console.error("Ошибка загрузки справочников:", err)
  }
}

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

const fetchUser = async () => {

  const token = localStorage.getItem("jwt")

  try {

    const res = await api.get(
      `/users/by-employee/${route.params.id}`,
      { headers: { Authorization: `Bearer ${token}` } }
    )

    user.value = res.data

    editUser.value.login = res.data.login

    fetchUserGroups()

  } catch {
    user.value = null
  }

}

const fetchUserGroups = async () => {

  const token = localStorage.getItem("jwt")

  const res = await api.get(
    `/users/${user.value.ID}/access-groups`,
    { headers: { Authorization: `Bearer ${token}` } }
  )

  const relations = res.data || []

  userGroups.value = relations.map(rel => {

    const group = accessGroups.value.find(
      g => g.ID === rel.AccessGroupID
    )

    return {
      AccessGroupID: rel.AccessGroupID,
      Name: group ? group.Name : "Неизвестная группа"
    }

  })

}

const createUser = async () => {

  const token = localStorage.getItem("jwt")

  await api.post(
    "/users",
    {
      login: newUser.value.login,
      password: newUser.value.password,
      employee_id: Number(route.params.id)
    },
    { headers: { Authorization: `Bearer ${token}` } }
  )

  showCreateUserModal.value = false

  showToast("Пользователь создан")

  fetchUser()

}

const updateUser = async () => {

  const token = localStorage.getItem("jwt")

  await api.put(
    `/users/${user.value.ID}`,
    {
      login: editUser.value.login,
      password: editUser.value.password
    },
    { headers: { Authorization: `Bearer ${token}` } }
  )

  showEditUserModal.value = false

  showToast("Пользователь обновлен")

  fetchUser()

}

const addGroup = async () => {

  const token = localStorage.getItem("jwt")

  await api.post(
    `/users/${user.value.ID}/access-groups`,
    { access_group_id: selectedGroup.value },
    { headers: { Authorization: `Bearer ${token}` } }
  )

  showAddGroupModal.value = false

  fetchUserGroups()

}

const removeGroup = async (groupId) => {

  const token = localStorage.getItem("jwt")

  await api.delete(
    `/users/${user.value.ID}/access-groups/${groupId}`,
    { headers: { Authorization: `Bearer ${token}` } }
  )

  fetchUserGroups()

}

const openDeleteEmployeeDialog = () => {
  deleteDialog.value = {
    title: "Удалить сотрудника?",
    message: "Вы действительно хотите удалить этого сотрудника?",
    actionType: "employee",
    groupId: null
  }
  showDeleteDialog.value = true
}

const openDeleteUserDialog = () => {
  if (!user.value?.ID) return
  deleteDialog.value = {
    title: "Удалить пользователя?",
    message: "Вы действительно хотите удалить пользователя этого сотрудника?",
    actionType: "user",
    groupId: null
  }
  showDeleteDialog.value = true
}

const openDeleteGroupDialog = (groupId) => {
  deleteDialog.value = {
    title: "Удалить группу доступа?",
    message: "Вы действительно хотите удалить группу доступа у пользователя?",
    actionType: "group",
    groupId
  }
  showDeleteDialog.value = true
}

const cancelDeleteDialog = () => {
  showDeleteDialog.value = false
  deleteDialog.value = { title: "", message: "", actionType: "", groupId: null }
}

const confirmDeleteDialog = async () => {
  const token = localStorage.getItem("jwt")

  try {
    if (deleteDialog.value.actionType === "employee") {
      await api.delete(`/employees/${route.params.id}`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      cancelDeleteDialog()
      showToast("Сотрудник удален")
      window.location.href = "/employees"
      return
    }

    if (deleteDialog.value.actionType === "user" && user.value?.ID) {
      await api.delete(`/users/${user.value.ID}`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      user.value = null
      userGroups.value = []
      cancelDeleteDialog()
      showToast("Пользователь удален")
      return
    }

    if (deleteDialog.value.actionType === "group" && deleteDialog.value.groupId) {
      await removeGroup(deleteDialog.value.groupId)
      cancelDeleteDialog()
      showToast("Группа доступа удалена")
      return
    }
  } catch (err) {
    console.error(err)
    showToast("Ошибка удаления")
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
      last_name: employee.value.last_name,
      first_name: employee.value.first_name,
      middle_name: employee.value.middle_name,

      birth_date: employee.value.birth_date || null,
      hire_date: employee.value.hire_date || null,
      fire_date: employee.value.fire_date || null,

      department_id: employee.value.department_id,
      position_id: employee.value.position_id,

      is_remote: employee.value.is_remote,
      salary: Number(employee.value.salary)
    }

    await api.put(`/employees/${route.params.id}`, payload, {
      headers: { Authorization: `Bearer ${token}` }
    })

    showToast("Сотрудник сохранен")

  } catch (err) {
    console.error("Ошибка сохранения:", err)
    showToast("Ошибка при сохранении")
  }
}

const mapDictionaryIds = () => {
  const dep = departments.value.find(d => d.Name === employee.value.department)
  const pos = positions.value.find(p => p.Name === employee.value.position)

  if (dep) employee.value.department_id = dep.ID
  if (pos) employee.value.position_id = pos.ID
}

onMounted(async () => {
  await fetchDictionaries()
  await fetchEmployee()
  await fetchUser()
  mapDictionaryIds()
})

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

.delete-top-btn {
  position: absolute;
  right: 20px;
  top: 20px;
  background: var(--color-soft-bg);
  border: none;
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
}

.delete-top-btn:hover {
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
  background: var(--color-primary);
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
.form-group input[type="number"],
.form-group input[type="password"],
.form-group select {
  padding: 10px;
  border-radius: 6px;
  border: 1px solid #ccc;
  transition: border-color 0.2s;
  box-sizing:border-box;
  font-size:14px;
}

.form-group input[type="text"]:focus,
.form-group input[type="date"]:focus,
.form-group input[type="number"]:focus,
.form-group input[type="password"]:focus,
.form-group select:focus {
  border-color: var(--color-primary);
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
  background-color: var(--color-primary);
  color: white;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  align-self: flex-start;
  font-weight: 500;
  transition: 0.2s;
}

.save-btn:hover {
  background-color: var(--color-primary-hover);
}

.user-header{
  display:flex;
  justify-content:space-between;
  align-items:center;
}

.user-actions{
  display:flex;
  gap:8px;
}

.groups-header{
  display:flex;
  justify-content:space-between;
  align-items:center;
  margin-top:10px;
}

.group-row{
  display:flex;
  justify-content:space-between;
  padding:6px 0;
  border-bottom:1px solid #eee;
}

.icon-btn{
  border:none;
  background:var(--color-soft-bg);
  border-radius:6px;
  padding:6px;
  cursor:pointer;
  display:inline-flex;
  align-items:center;
  justify-content:center;
}

.icon-btn.edit:hover{
  background:var(--color-edit-hover);
}

.icon-btn.delete:hover{
  background:var(--color-delete-hover);
}

.icon-btn:hover{
  background:var(--color-muted-bg);
}

.delete-icon{
  cursor:pointer;
  color:#ef4444;
}

.modal-overlay{
  position:fixed;
  inset:0;
  background:rgba(0,0,0,0.3);
  display:flex;
  align-items:center;
  justify-content:center;
}

.modal{
  background:white;
  padding:24px;
  border-radius:10px;
  width:360px;
  display:flex;
  flex-direction:column;
  gap:14px;
}

.modal-actions{
  display:flex;
  justify-content:flex-end;
  gap:10px;
}

.cancel-btn{
  padding:10px 14px;
  background:var(--color-muted-bg);
  border:none;
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
</style>

