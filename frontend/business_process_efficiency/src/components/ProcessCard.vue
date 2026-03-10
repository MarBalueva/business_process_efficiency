<template>
  <div class="employee-page-container">
    <div class="employee-card">
      <button class="back-btn" @click="$router.push('/processes')">← Назад</button>
      <button class="delete-top-btn" @click="openDeleteDialog">Удалить</button>

      <h2 class="employee-title">{{ process.Name }}</h2>

      <div class="tabs">
        <button :class="{ active: activeTab === 'data' }" @click="activeTab = 'data'">
          Данные
        </button>

        <button
          v-for="(version, index) in sortedVersions"
          :key="version.ID"
          :class="{ active: activeTab === 'version' + index }"
          @click="activeTab = 'version' + index"
        >
          Версия {{ version.Version }}
        </button>

        <button class="add-version-btn" @click="createVersion">
          + Версия
        </button>

        <button :class="{ active: activeTab === 'analytics' }" @click="activeTab = 'analytics'">
          Аналитика
        </button>
      </div>

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

            <div class="form-group">
              <label>Единица времени</label>
              <select v-model="process.RegularityUnit">
                <option value="">Не выбрано</option>
                <option value="day">День</option>
                <option value="week">Неделя</option>
                <option value="month">Месяц</option>
                <option value="quarter">Квартал</option>
                <option value="halfyear">Полугодие</option>
                <option value="year">Год</option>
              </select>
            </div>

            <div class="form-group">
              <label>Количество процессов в единицу времени</label>
              <input v-model.number="process.RegularityCount" type="number" min="0" />
            </div>
          </div>

          <div class="column">
            <div class="form-group">
              <label>Дата создания</label>
              <input type="text" :value="formatDateTime(process.CreatedAt)" disabled />
            </div>

            <div class="form-group checkbox-group">
              <label>
                <input type="checkbox" v-model="process.IsActive" />
                Актуальный
              </label>
            </div>

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

      <div v-else-if="activeTab === 'analytics'" class="analytics-content">
        <div class="analytics-kpis">
          <div class="kpi-card">
            <div class="kpi-label">Общее время процесса</div>
            <div class="kpi-value">{{ formatMinutes(totalProcessMinutes) }} мин</div>
          </div>
          <div class="kpi-card">
            <div class="kpi-label">Общая стоимость процесса</div>
            <div class="kpi-value">{{ formatCurrency(totalProcessCost) }}</div>
          </div>
          <div class="kpi-card">
            <div class="kpi-label">Версия для аналитики</div>
            <div class="kpi-value">v{{ analyticsVersion?.Version ?? "—" }}</div>
          </div>
        </div>

        <div class="analytics-grid">
          <div class="chart-card">
            <h4>Длительность этапов</h4>
            <div class="chart-wrap">
              <Bar :data="stepDurationChartData" :options="stepDurationChartOptions" />
            </div>
          </div>

          <div class="chart-card">
            <h4>Стоимость этапов</h4>
            <div class="chart-wrap">
              <Bar :data="stepCostChartData" :options="stepCostChartOptions" />
            </div>
          </div>

          <div class="chart-card">
            <h4>Время ожидания между этапами</h4>
            <div class="chart-wrap">
              <Line :data="waitingChartData" :options="waitingChartOptions" />
            </div>
          </div>

          <div class="chart-card">
            <h4>Плановое и фактическое время этапов</h4>
            <div class="chart-wrap">
              <Bar :data="planFactChartData" :options="planFactChartOptions" />
            </div>
          </div>

          <div class="chart-card">
            <h4>Распределение сотрудников по этапам</h4>
            <div class="chart-wrap">
              <Doughnut :data="employeeDistributionChartData" :options="employeeDistributionChartOptions" />
            </div>
          </div>
        </div>
      </div>

      <ProcessVersionView
        v-if="activeVersion"
        :key="activeVersion.ID"
        :version="activeVersion"
        :processId="process.ID"
        @refresh="fetchProcess"
      />
    </div>

    <div v-if="toast.show" class="toast">{{ toast.message }}</div>

    <div v-if="showDeleteDialog" class="modal-overlay">
      <div class="modal">
        <h3>Удаление процесса</h3>

        <p>
          Вы действительно хотите удалить процесс
          <b>{{ process.Name }}</b>?
        </p>

        <div class="modal-actions">
          <button class="cancel-btn" @click="showDeleteDialog = false">
            Отмена
          </button>

          <button class="save-btn" @click="deleteProcess">
            Удалить
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue"
import { useRoute } from "vue-router"
import { Bar, Doughnut, Line } from "vue-chartjs"
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  ArcElement,
  PointElement,
  LineElement,
  CategoryScale,
  LinearScale
} from "chart.js"
import api from "../api/axios"
import ProcessVersionView from "../components/ProcessVersionView.vue"

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  BarElement,
  ArcElement,
  PointElement,
  LineElement,
  CategoryScale,
  LinearScale
)

const route = useRoute()
const process = ref({
  Name: "",
  Description: "",
  Regulations: "",
  OwnerID: null,
  IsActive: true,
  RegularityCount: 0,
  RegularityUnit: "",
  Versions: [],
  CreatedAt: ""
})

const employees = ref([])
const selectedOwner = ref("")

const activeTab = ref("data")
const showDeleteDialog = ref(false)

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
    process.value.RegularityCount = Number(process.value.RegularityCount || 0)
    process.value.RegularityUnit = process.value.RegularityUnit || ""

    selectedOwner.value = process.value.OwnerID ? process.value.OwnerID.toString() : ""
  } catch (err) {
    console.error(err)
  }
}

const activeVersionIndex = computed(() => {
  if (!activeTab.value.startsWith("version")) return null
  return Number(activeTab.value.replace("version", ""))
})

const activeVersion = computed(() => {
  if (activeVersionIndex.value === null) return null
  return sortedVersions.value[activeVersionIndex.value]
})

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

const sortedVersions = computed(() => {
  return [...process.value.Versions].sort((a, b) => Number(a.Version) - Number(b.Version))
})

const analyticsVersion = computed(() => {
  if (!sortedVersions.value.length) return null
  return sortedVersions.value[sortedVersions.value.length - 1]
})

const analyticsSteps = computed(() => {
  const steps = Array.isArray(analyticsVersion.value?.Steps) ? [...analyticsVersion.value.Steps] : []
  const trackable = steps.filter((step) => ["SUBPROCESS", "OPERATION"].includes(step?.Type))
  return trackable.sort((a, b) => {
    const orderA = Number(a?.StepOrder)
    const orderB = Number(b?.StepOrder)
    const idA = Number(a?.ID || 0)
    const idB = Number(b?.ID || 0)
    if (Number.isFinite(orderA) && Number.isFinite(orderB) && orderA !== orderB) return orderA - orderB
    return idA - idB
  })
})

const analyticsAllSteps = computed(() => {
  const steps = Array.isArray(analyticsVersion.value?.Steps) ? [...analyticsVersion.value.Steps] : []
  return steps.sort((a, b) => {
    const orderA = Number(a?.StepOrder)
    const orderB = Number(b?.StepOrder)
    const idA = Number(a?.ID || 0)
    const idB = Number(b?.ID || 0)
    if (Number.isFinite(orderA) && Number.isFinite(orderB) && orderA !== orderB) return orderA - orderB
    return idA - idB
  })
})

const getStepName = (step) => step?.Name || `Этап #${step?.ID ?? ""}`
const getStepExecutors = (step) => (Array.isArray(step?.StepExecutors) ? step.StepExecutors : [])
const getStepMeasurements = (step) => (Array.isArray(step?.Measurements) ? step.Measurements : [])
const getStepActualMinutes = (step) => {
  const n = Number(step?.FinalDurationMin)
  return Number.isFinite(n) ? n : 0
}
const getStepPlannedMinutes = (step) => {
  const n = Number(step?.Metrics?.PlannedTimeMin)
  return Number.isFinite(n) ? n : 0
}
const getStepParallelIds = (step) => {
  return (Array.isArray(step?.ParallelSteps) ? step.ParallelSteps : [])
    .map((p) => Number(p?.ParallelStepID))
    .filter((id) => Number.isFinite(id) && id > 0)
}

const employeeRateById = computed(() => {
  const map = new Map()
  for (const e of employees.value || []) {
    const id = Number(e?.id ?? e?.ID)
    const salary = Number(e?.salary ?? e?.Salary ?? 0)
    if (Number.isFinite(id)) map.set(id, Number.isFinite(salary) ? salary : 0)
  }
  return map
})

const getStepCost = (step) => {
  const totalMin = getStepActualMinutes(step)
  if (totalMin <= 0) return 0

  const loads = getStepExecutors(step)
  const fallbackExecutors = Array.isArray(step?.Executors) ? step.Executors : []

  if (loads.length === 0 && fallbackExecutors.length === 0) return 0

  let cost = 0
  if (loads.length > 0) {
    for (const load of loads) {
      const employeeId = Number(load?.EmployeeID ?? load?.employeeId ?? 0)
      const ratePerHour = Number(employeeRateById.value.get(employeeId) || 0)
      const percent = Number(load?.WorkloadPercent ?? load?.workloadPercent ?? 0)
      if (!Number.isFinite(ratePerHour) || !Number.isFinite(percent)) continue
      cost += (totalMin / 60) * ratePerHour * (percent / 100)
    }
  } else {
    const share = 100 / fallbackExecutors.length
    for (const ex of fallbackExecutors) {
      const employeeId = Number(ex?.ID ?? ex?.id ?? 0)
      const ratePerHour = Number(employeeRateById.value.get(employeeId) || 0)
      if (!Number.isFinite(ratePerHour)) continue
      cost += (totalMin / 60) * ratePerHour * (share / 100)
    }
  }
  return Number.isFinite(cost) ? cost : 0
}

const stepAnalyticsRows = computed(() => {
  return analyticsSteps.value.map((step) => ({
    id: Number(step?.ID || 0),
    name: getStepName(step),
    actualMin: getStepActualMinutes(step),
    plannedMin: getStepPlannedMinutes(step),
    cost: getStepCost(step)
  }))
})

const stepProbabilityById = computed(() => {
  const prob = new Map(stepAnalyticsRows.value.map((row) => [row.id, 1]))
  for (const step of analyticsAllSteps.value) {
    if (step?.Type !== "CONDITION") continue
    const branches = Array.isArray(step?.ConditionBranches) ? step.ConditionBranches : []
    for (const branch of branches) {
      const nextId = Number(branch?.NextStepID)
      const p = Number(branch?.ProbabilityPercent)
      if (!Number.isFinite(nextId) || !Number.isFinite(p) || p < 0) continue
      if (!prob.has(nextId)) continue
      const current = Number(prob.get(nextId) || 1)
      prob.set(nextId, current * (p / 100))
    }
  }
  return prob
})

const groupedExpectedTotals = computed(() => {
  const rows = stepAnalyticsRows.value
  if (rows.length === 0) return { minutes: 0, cost: 0 }

  const rowById = new Map(rows.map((r) => [r.id, r]))
  const parent = new Map(rows.map((r) => [r.id, r.id]))
  const find = (x) => {
    let p = parent.get(x)
    while (p !== parent.get(p)) p = parent.get(p)
    let cur = x
    while (parent.get(cur) !== p) {
      const next = parent.get(cur)
      parent.set(cur, p)
      cur = next
    }
    return p
  }
  const union = (a, b) => {
    const ra = find(a)
    const rb = find(b)
    if (ra !== rb) parent.set(rb, ra)
  }

  for (const step of analyticsSteps.value) {
    const id = Number(step?.ID)
    if (!Number.isFinite(id) || !rowById.has(id)) continue
    for (const pid of getStepParallelIds(step)) {
      if (rowById.has(pid)) union(id, pid)
    }
  }

  const groups = new Map()
  for (const row of rows) {
    const root = find(row.id)
    if (!groups.has(root)) groups.set(root, [])
    groups.get(root).push(row)
  }

  let minutes = 0
  let cost = 0
  for (const groupRows of groups.values()) {
    let groupMaxDuration = 0
    let groupCost = 0
    for (const row of groupRows) {
      const weight = Number(stepProbabilityById.value.get(row.id) ?? 1)
      const adjustedDuration = Number(row.actualMin || 0) * weight
      const adjustedCost = Number(row.cost || 0) * weight
      groupMaxDuration = Math.max(groupMaxDuration, adjustedDuration)
      groupCost += adjustedCost
    }
    minutes += groupMaxDuration
    cost += groupCost
  }

  return { minutes, cost }
})

const totalProcessMinutes = computed(() => groupedExpectedTotals.value.minutes)
const totalProcessCost = computed(() => groupedExpectedTotals.value.cost)

const waitingRows = computed(() => {
  const rows = []
  for (let i = 0; i < analyticsSteps.value.length - 1; i += 1) {
    const current = analyticsSteps.value[i]
    const next = analyticsSteps.value[i + 1]

    const currentFinished = getStepMeasurements(current)
      .filter((m) => !!m?.FinishedAt)
      .map((m) => new Date(m.FinishedAt).getTime())
      .filter((v) => Number.isFinite(v))
    const nextStarted = getStepMeasurements(next)
      .filter((m) => !!m?.StartedAt)
      .map((m) => new Date(m.StartedAt).getTime())
      .filter((v) => Number.isFinite(v))

    const maxFinish = currentFinished.length ? Math.max(...currentFinished) : null
    const minStart = nextStarted.length ? Math.min(...nextStarted) : null

    let waitSec = 0
    if (maxFinish !== null && minStart !== null) {
      waitSec = Math.max(0, Math.round((minStart - maxFinish) / 1000))
    }

    rows.push({
      label: `${getStepName(current)} -> ${getStepName(next)}`,
      waitMin: waitSec / 60
    })
  }
  return rows
})

const employeeDistributionRows = computed(() => {
  const byEmployee = new Map()
  for (const step of analyticsSteps.value) {
    const seenOnStep = new Set()
    for (const load of getStepExecutors(step)) {
      const id = Number(load?.EmployeeID ?? load?.employeeId ?? 0)
      if (!Number.isFinite(id) || id <= 0 || seenOnStep.has(id)) continue
      seenOnStep.add(id)
      byEmployee.set(id, (byEmployee.get(id) || 0) + 1)
    }
  }

  const formatName = (e) => {
    const ln = e?.last_name ?? e?.LastName ?? ""
    const fn = e?.first_name ?? e?.FirstName ?? ""
    const mn = e?.middle_name ?? e?.MiddleName ?? ""
    return `${ln} ${fn} ${mn}`.trim() || `ID ${e?.id ?? e?.ID}`
  }

  const employeeById = new Map((employees.value || []).map((e) => [Number(e?.id ?? e?.ID), e]))
  return [...byEmployee.entries()]
    .map(([employeeId, stepsCount]) => ({
      employeeId,
      label: employeeById.has(employeeId) ? formatName(employeeById.get(employeeId)) : `ID ${employeeId}`,
      stepsCount
    }))
    .sort((a, b) => b.stepsCount - a.stepsCount)
})

const palette = {
  primary: "#2563eb",
  secondary: "#0ea5e9",
  accent: "#14b8a6",
  warning: "#f59e0b",
  info: "#8b5cf6",
  neutral: "#64748b",
  success: "#10b981",
  rose: "#f43f5e"
}
const chartColors = [
  palette.primary,
  palette.secondary,
  palette.accent,
  palette.warning,
  palette.info,
  palette.success,
  palette.rose,
  "#22c55e"
]
const rubFormat = new Intl.NumberFormat("ru-RU", { style: "currency", currency: "RUB", maximumFractionDigits: 2 })

const formatCurrency = (value) => rubFormat.format(Number(value || 0))
const formatMinutes = (value) => Number(value || 0).toFixed(2)

const commonChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: "bottom" }
  }
}

const stepDurationRowsDesc = computed(() =>
  [...stepAnalyticsRows.value]
    .sort((a, b) => Number(b.actualMin || 0) - Number(a.actualMin || 0))
)

const stepCostRowsDesc = computed(() =>
  [...stepAnalyticsRows.value]
    .sort((a, b) => Number(b.cost || 0) - Number(a.cost || 0))
)

const stepDurationChartData = computed(() => ({
  labels: stepDurationRowsDesc.value.map((s) => s.name),
  datasets: [
    {
      label: "Время (мин)",
      data: stepDurationRowsDesc.value.map((s) => Number(s.actualMin.toFixed(2))),
      backgroundColor: palette.secondary,
      borderRadius: 6
    }
  ]
}))

const stepDurationChartOptions = {
  ...commonChartOptions,
  indexAxis: "y",
  scales: {
    x: { beginAtZero: true, title: { display: true, text: "Минуты" } },
    y: { ticks: { autoSkip: false } }
  }
}

const stepCostChartData = computed(() => ({
  labels: stepCostRowsDesc.value.map((s) => s.name),
  datasets: [
    {
      label: "Стоимость (₽)",
      data: stepCostRowsDesc.value.map((s) => Number(s.cost.toFixed(2))),
      backgroundColor: palette.success,
      borderRadius: 6
    }
  ]
}))

const stepCostChartOptions = {
  ...commonChartOptions,
  indexAxis: "y",
  scales: {
    x: { beginAtZero: true, title: { display: true, text: "₽" } },
    y: { ticks: { autoSkip: false } }
  }
}

const waitingChartData = computed(() => ({
  labels: waitingRows.value.map((r) => r.label),
  datasets: [
    {
      label: "Ожидание (мин)",
      data: waitingRows.value.map((r) => Number(r.waitMin.toFixed(2))),
      borderColor: palette.warning,
      backgroundColor: "rgba(245,158,11,0.20)",
      fill: true,
      tension: 0.3
    }
  ]
}))

const waitingChartOptions = {
  ...commonChartOptions,
  scales: {
    x: { ticks: { maxRotation: 40, minRotation: 20 } },
    y: { beginAtZero: true, title: { display: true, text: "Минуты" } }
  }
}

const planFactChartData = computed(() => ({
  labels: stepAnalyticsRows.value.map((s) => s.name),
  datasets: [
    {
      label: "План (мин)",
      data: stepAnalyticsRows.value.map((s) => Number(s.plannedMin.toFixed(2))),
      backgroundColor: "rgba(14,165,233,0.35)"
    },
    {
      label: "Факт (мин)",
      data: stepAnalyticsRows.value.map((s) => Number(s.actualMin.toFixed(2))),
      backgroundColor: palette.primary
    }
  ]
}))

const planFactChartOptions = {
  ...commonChartOptions,
  indexAxis: "y",
  scales: {
    x: { beginAtZero: true, title: { display: true, text: "Минуты" } },
    y: { ticks: { autoSkip: false } }
  }
}

const employeeDistributionChartData = computed(() => ({
  labels: employeeDistributionRows.value.map((r) => r.label),
  datasets: [
    {
      label: "Количество этапов",
      data: employeeDistributionRows.value.map((r) => r.stepsCount),
      backgroundColor: employeeDistributionRows.value.map((_, i) => chartColors[i % chartColors.length]),
      borderWidth: 1
    }
  ]
}))

const employeeDistributionChartOptions = {
  ...commonChartOptions,
  plugins: {
    ...commonChartOptions.plugins,
    tooltip: {
      callbacks: {
        label: (ctx) => `${ctx.label}: ${ctx.raw} этап(ов)`
      }
    }
  }
}

const saveProcess = async () => {
  const token = localStorage.getItem("jwt")

  try {
    await api.put(`/processes/${route.params.id}`, {
      name: process.value.Name,
      description: process.value.Description,
      regulations: process.value.Regulations,
      owner_id: selectedOwner.value ? Number(selectedOwner.value) : 0,
      is_active: process.value.IsActive,
      regularity_count: Number(process.value.RegularityCount || 0),
      regularity_unit: Number(process.value.RegularityCount || 0) > 0 ? process.value.RegularityUnit : ""
    }, {
      headers: { Authorization: `Bearer ${token}` }
    })

    showToast("Процесс сохранен")
  } catch (err) {
    console.error("Ошибка сохранения процесса:", err)
    showToast("Ошибка при сохранении процесса")
  }
}

const openDeleteDialog = () => {
  showDeleteDialog.value = true
}

const deleteProcess = async () => {
  const token = localStorage.getItem("jwt")

  try {
    await api.delete(`/processes/${route.params.id}`, {
      headers: { Authorization: `Bearer ${token}` }
    })

    showToast("Процесс удален")

    window.location.href = "/processes"
  } catch (err) {
    console.error("Ошибка удаления:", err)
    showToast("Ошибка удаления")
  }
}

const createVersion = async () => {
  const token = localStorage.getItem("jwt")

  try {
    await api.post(`/processes/versions`, {
      processId: Number(route.params.id)
    }, {
      headers: { Authorization: `Bearer ${token}` }
    })

    showToast("Версия создана")

    await fetchProcess()
  } catch (err) {
    console.error(err)
    showToast("Ошибка создания версии")
  }
}

const formatDateTime = (iso) => {
  if (!iso) return ""
  const d = new Date(iso)
  const day = d.getDate().toString().padStart(2, "0")
  const month = (d.getMonth() + 1).toString().padStart(2, "0")
  const year = d.getFullYear()
  const hours = d.getHours().toString().padStart(2, "0")
  const minutes = d.getMinutes().toString().padStart(2, "0")
  return `${day}.${month}.${year} ${hours}:${minutes}`
}

onMounted(async () => {
  await fetchEmployees()
  await fetchProcess()
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  justify-content: center;
  align-items: center;
}

.modal {
  background: white;
  padding: 30px;
  border-radius: 10px;
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.modal.modal-wide {
  width: 600px;
  padding: 30px;
}

.modal label {
  font-weight: 500;
  margin-bottom: 4px;
}

.modal-columns {
  display: flex;
  gap: 20px;
}

.modal-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 10px;
}

.save-btn {
  background: var(--color-primary);
  color: white;
  border: none;
  padding: 8px 14px;
  border-radius: 6px;
  cursor: pointer;
}

.cancel-btn {
  background: var(--color-muted-bg);
  border: none;
  padding: 8px 14px;
  border-radius: 6px;
  cursor: pointer;
}

.toast {
  position: fixed;
  bottom: 30px;
  right: 30px;
  background: var(--color-primary);
  color: white;
  padding: 12px 18px;
  border-radius: 8px;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
}

.employee-page-container {
  margin-left: 240px;
  padding: 24px;
  width: calc(100% - 240px);
  box-sizing: border-box;
  display: flex;
  justify-content: center;
}

.employee-card {
  background-color: #fff;
  border-radius: 12px;
  padding: 30px;
  width: 100%;
  max-width: 1700px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
  gap: 20px;
  position: relative;
}

@media (max-width: 1500px) {
  .employee-page-container {
    padding: 16px;
  }

  .employee-card {
    max-width: 100%;
  }
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

.back-btn:hover,
.add-version-btn:hover,
.delete-top-btn:hover {
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

.add-version-btn {
  margin-left: 10px;
  background: var(--color-soft-bg);
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

.analytics-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.analytics-kpis {
  display: grid;
  grid-template-columns: repeat(3, minmax(220px, 1fr));
  gap: 12px;
}

.kpi-card {
  background: linear-gradient(160deg, #f8fbff 0%, #eef6ff 100%);
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 14px;
}

.kpi-label {
  font-size: 13px;
  color: #64748b;
  margin-bottom: 8px;
}

.kpi-value {
  font-size: 24px;
  font-weight: 700;
  color: #1e293b;
}

.analytics-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(320px, 1fr));
  gap: 14px;
}

.chart-card {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 12px 14px;
}

.chart-card h4 {
  margin: 0 0 8px;
  font-size: 15px;
  color: #334155;
}

.chart-wrap {
  height: 280px;
}

.analytics-grid .chart-card:last-child {
  grid-column: 1 / -1;
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
.form-group input[type="number"],
.form-group select,
.form-group textarea {
  padding: 10px;
  border-radius: 6px;
  border: 1px solid #ccc;
  transition: border-color 0.2s;
  font-size: 14px;
  font-family: "Inter", sans-serif;
}

.form-group input[type="text"]:focus,
.form-group input[type="number"]:focus,
.form-group select:focus,
.form-group textarea:focus {
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

@media (max-width: 1200px) {
  .analytics-kpis {
    grid-template-columns: 1fr;
  }

  .analytics-grid {
    grid-template-columns: 1fr;
  }

  .analytics-grid .chart-card:last-child {
    grid-column: auto;
  }
}

.toast {
  position: fixed;
  bottom: 30px;
  right: 30px;
  background: var(--color-primary);
  color: white;
  padding: 12px 18px;
  border-radius: 8px;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
}
</style>
