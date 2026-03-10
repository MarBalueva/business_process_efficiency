<template>
  <div class="dashboards-page-container">
    <div class="dashboards-card">
      <div class="dashboards-header">
        <h2>Дашборды процессов</h2>
        <button class="refresh-btn" @click="loadDashboardData" :disabled="loading">
          {{ loading ? "Обновление..." : "Обновить" }}
        </button>
      </div>

      <div v-if="error" class="error-box">{{ error }}</div>

      <div class="filters">
        <label>
          Период для загрузки сотрудников
          <select v-model="selectedLoadPeriod">
            <option value="day">День</option>
            <option value="week">Неделя</option>
            <option value="month">Месяц</option>
            <option value="quarter">Квартал</option>
            <option value="halfyear">Полугодие</option>
            <option value="year">Год</option>
          </select>
        </label>

        <label>
          Подразделение
          <select v-model="selectedDepartment">
            <option value="">Все подразделения</option>
            <option v-for="d in departmentOptions" :key="d" :value="d">{{ d }}</option>
          </select>
        </label>

        <label>
          Тип этапа
          <select v-model="selectedStepType">
            <option value="">Все типы</option>
            <option v-for="t in stepTypeOptions" :key="t.value" :value="t.value">{{ t.label }}</option>
          </select>
        </label>
      </div>

      <div class="charts-grid">
        <div class="chart-card">
          <h3>Линейный график длительности этапов</h3>
          <p class="chart-subtitle">Срез по подразделениям и типам этапов</p>
          <div class="chart-wrap">
            <Line :data="durationByDepartmentAndTypeData" :options="durationByDepartmentAndTypeOptions" />
          </div>
        </div>

        <div class="chart-card">
          <h3>План и факт по отклонениям</h3>
          <p class="chart-subtitle">Максимальное, минимальное и среднее отклонение</p>
          <div class="chart-wrap">
            <Bar :data="planFactDeviationData" :options="planFactDeviationOptions" />
          </div>
        </div>

        <div class="chart-card chart-card-wide">
          <h3>Загрузка сотрудников на процессах</h3>
          <p class="chart-subtitle">Оценка занятости по выбранному периоду</p>
          <div class="chart-wrap">
            <Bar :data="employeeLoadData" :options="employeeLoadOptions" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue"
import { Bar, Line } from "vue-chartjs"
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  PointElement,
  LineElement,
  CategoryScale,
  LinearScale
} from "chart.js"
import api from "../api/axios"

ChartJS.register(Title, Tooltip, Legend, BarElement, PointElement, LineElement, CategoryScale, LinearScale)

const loading = ref(false)
const error = ref("")
const selectedLoadPeriod = ref("month")
const selectedDepartment = ref("")
const selectedStepType = ref("")
const processes = ref([])
const employees = ref([])

const stepTypeLabels = {
  START: "Старт",
  END: "Финиш",
  INTERMEDIATE: "Промежуточный",
  SUBPROCESS: "Подпроцесс",
  OPERATION: "Операция",
  CONDITION: "Условие"
}

const palette = {
  blue: "#2563eb",
  sky: "#0ea5e9",
  teal: "#14b8a6",
  amber: "#f59e0b",
  violet: "#8b5cf6",
  green: "#10b981",
  rose: "#f43f5e",
  slate: "#64748b"
}

const chartColors = [palette.blue, palette.sky, palette.teal, palette.amber, palette.violet, palette.green, palette.rose]

const getAuthHeaders = () => ({
  headers: { Authorization: `Bearer ${localStorage.getItem("jwt")}` }
})

const normalizeEmployeeId = (v) => Number(v?.id ?? v?.ID ?? 0)
const normalizeProcessRegularityCount = (p) => Number(p?.RegularityCount ?? p?.regularity_count ?? 0)
const normalizeProcessRegularityUnit = (p) => String(p?.RegularityUnit ?? p?.regularity_unit ?? "").toLowerCase()
const normalizeStepType = (s) => String(s?.Type ?? "")
const normalizeStepActualMin = (s) => Number(s?.FinalDurationMin ?? 0)
const normalizeStepPlannedMin = (s) => Number(s?.Metrics?.PlannedTimeMin ?? 0)

const getStepExecutorsWithPercent = (step) => {
  const loads = Array.isArray(step?.StepExecutors) ? step.StepExecutors : []
  if (loads.length > 0) {
    return loads
      .map((x) => ({
        employeeId: Number(x?.EmployeeID ?? x?.employeeId ?? 0),
        percent: Number(x?.WorkloadPercent ?? x?.workloadPercent ?? 0)
      }))
      .filter((x) => Number.isFinite(x.employeeId) && x.employeeId > 0)
  }

  const executors = Array.isArray(step?.Executors) ? step.Executors : []
  if (executors.length === 0) return []
  const equalPercent = 100 / executors.length
  return executors
    .map((e) => ({
      employeeId: Number(e?.ID ?? e?.id ?? 0),
      percent: equalPercent
    }))
    .filter((x) => Number.isFinite(x.employeeId) && x.employeeId > 0)
}

const getProcessLatestVersion = (process) => {
  const versions = Array.isArray(process?.Versions) ? [...process.Versions] : []
  if (!versions.length) return null
  versions.sort((a, b) => Number(a?.Version ?? 0) - Number(b?.Version ?? 0))
  return versions[versions.length - 1]
}

const getVersionSteps = (version) => {
  const steps = Array.isArray(version?.Steps) ? [...version.Steps] : []
  return steps.sort((a, b) => {
    const oa = Number(a?.StepOrder ?? 0)
    const ob = Number(b?.StepOrder ?? 0)
    if (oa !== ob) return oa - ob
    return Number(a?.ID ?? 0) - Number(b?.ID ?? 0)
  })
}

const extractProcessIdsFromRegistry = (folders) => {
  const ids = []
  const walk = (nodes) => {
    for (const node of nodes || []) {
      for (const p of node?.processes || node?.Processes || []) {
        const id = Number(p?.id ?? p?.ID ?? 0)
        if (Number.isFinite(id) && id > 0) ids.push(id)
      }
      walk(node?.children || node?.Children || [])
    }
  }
  walk(folders || [])
  return [...new Set(ids)]
}

const loadDashboardData = async () => {
  loading.value = true
  error.value = ""
  try {
    const [employeeRes, registryRes] = await Promise.all([
      api.get("/employees", getAuthHeaders()),
      api.get("/processes/registry", getAuthHeaders())
    ])

    employees.value = employeeRes?.data || []
    const processIds = extractProcessIdsFromRegistry(registryRes?.data || [])

    const processResponses = await Promise.all(
      processIds.map((id) => api.get(`/processes/${id}`, getAuthHeaders()).catch(() => null))
    )

    processes.value = processResponses
      .map((r) => r?.data)
      .filter(Boolean)
  } catch (e) {
    console.error(e)
    error.value = "Не удалось загрузить данные для дашбордов"
  } finally {
    loading.value = false
  }
}

onMounted(loadDashboardData)

const employeeById = computed(() => {
  const map = new Map()
  for (const e of employees.value) {
    map.set(normalizeEmployeeId(e), e)
  }
  return map
})

const departmentOptions = computed(() => {
  const set = new Set()
  for (const e of employees.value) {
    const dep = String(e?.department ?? e?.Department ?? "").trim()
    if (dep) set.add(dep)
  }
  return [...set].sort((a, b) => a.localeCompare(b, "ru"))
})

const stepTypeOptions = computed(() =>
  Object.entries(stepTypeLabels).map(([value, label]) => ({ value, label }))
)

const stepMatchesTypeFilter = (step) => {
  if (!selectedStepType.value) return true
  return normalizeStepType(step) === selectedStepType.value
}

const employeeMatchesDepartmentFilter = (employee) => {
  if (!selectedDepartment.value) return true
  const dep = String(employee?.department ?? employee?.Department ?? "")
  return dep === selectedDepartment.value
}

const stepsGlobal = computed(() => {
  const rows = []
  for (const process of processes.value) {
    const version = getProcessLatestVersion(process)
    const steps = getVersionSteps(version)
    for (const step of steps) {
      rows.push({ process, step })
    }
  }
  return rows
})

const durationByDepartmentAndTypeData = computed(() => {
  const typeOrder = ["START", "INTERMEDIATE", "OPERATION", "CONDITION", "SUBPROCESS", "END"]
  const typeLabels = typeOrder.map((t) => stepTypeLabels[t] || t)

  const deptTypeMinutes = new Map()

  for (const row of stepsGlobal.value) {
    const step = row.step
    if (!stepMatchesTypeFilter(step)) continue

    const stepType = normalizeStepType(step)
    const typeKey = typeOrder.includes(stepType) ? stepType : "INTERMEDIATE"
    const actualMin = normalizeStepActualMin(step)
    if (!Number.isFinite(actualMin) || actualMin <= 0) continue

    const loads = getStepExecutorsWithPercent(step)
    if (loads.length === 0) continue

    for (const load of loads) {
      const employee = employeeById.value.get(load.employeeId)
      if (!employeeMatchesDepartmentFilter(employee)) continue
      const dept = String(employee?.department ?? employee?.Department ?? "Без подразделения")
      const key = `${dept}__${typeKey}`
      const minutes = actualMin * (Number(load.percent || 0) / 100)
      deptTypeMinutes.set(key, (deptTypeMinutes.get(key) || 0) + minutes)
    }
  }

  const departments = [...new Set([...deptTypeMinutes.keys()].map((k) => k.split("__")[0]))]
  departments.sort((a, b) => a.localeCompare(b, "ru"))

  const datasets = departments.slice(0, 8).map((dept, idx) => ({
    label: dept,
    data: typeOrder.map((type) => Number((deptTypeMinutes.get(`${dept}__${type}`) || 0).toFixed(2))),
    borderColor: chartColors[idx % chartColors.length],
    backgroundColor: chartColors[idx % chartColors.length],
    tension: 0.35
  }))

  return { labels: typeLabels, datasets }
})

const durationByDepartmentAndTypeOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { position: "bottom" } },
  scales: {
    y: { beginAtZero: true, title: { display: true, text: "Минуты" } }
  }
}

const planFactDeviationData = computed(() => {
  const pairs = stepsGlobal.value
    .map((row) => {
      if (!stepMatchesTypeFilter(row.step)) return null

      const loads = getStepExecutorsWithPercent(row.step)
      if (selectedDepartment.value) {
        const hasDepartmentExecutor = loads.some((l) => {
          const employee = employeeById.value.get(l.employeeId)
          return employeeMatchesDepartmentFilter(employee)
        })
        if (!hasDepartmentExecutor) return null
      }

      const planned = normalizeStepPlannedMin(row.step)
      const actual = normalizeStepActualMin(row.step)
      return { planned, actual, absDev: Math.abs(actual - planned) }
    })
    .filter((x) => x && Number.isFinite(x.planned) && Number.isFinite(x.actual) && x.planned > 0 && x.actual > 0)

  if (pairs.length === 0) {
    return {
      labels: ["Нет данных"],
      datasets: [
        { label: "План (мин)", data: [0], backgroundColor: "rgba(14,165,233,0.35)" },
        { label: "Факт (мин)", data: [0], backgroundColor: palette.blue }
      ]
    }
  }

  const sorted = [...pairs].sort((a, b) => a.absDev - b.absDev)
  const minPair = sorted[0]
  const maxPair = sorted[sorted.length - 1]
  const avgPlanned = pairs.reduce((a, x) => a + x.planned, 0) / pairs.length
  const avgActual = pairs.reduce((a, x) => a + x.actual, 0) / pairs.length

  return {
    labels: ["Наибольшее отклонение", "Наименьшее отклонение", "Среднее отклонение"],
    datasets: [
      {
        label: "План (мин)",
        data: [maxPair.planned, minPair.planned, avgPlanned].map((v) => Number(v.toFixed(2))),
        backgroundColor: "rgba(14,165,233,0.35)"
      },
      {
        label: "Факт (мин)",
        data: [maxPair.actual, minPair.actual, avgActual].map((v) => Number(v.toFixed(2))),
        backgroundColor: palette.blue
      }
    ]
  }
})

const planFactDeviationOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { position: "bottom" } },
  scales: {
    y: { beginAtZero: true, title: { display: true, text: "Минуты" } }
  }
}

const periodToMonths = {
  day: 1 / 30,
  week: 7 / 30,
  month: 1,
  quarter: 3,
  halfyear: 6,
  year: 12
}

const regularityToMonthlyCount = (count, unit) => {
  const c = Number(count || 0)
  if (!Number.isFinite(c) || c <= 0) return 0
  switch (String(unit || "").toLowerCase()) {
    case "day":
      return c * 30
    case "week":
      return c * 4.2857
    case "month":
      return c
    case "quarter":
      return c / 3
    case "halfyear":
      return c / 6
    case "year":
      return c / 12
    default:
      return 0
  }
}

const employeeLoadData = computed(() => {
  const selectedFactor = periodToMonths[selectedLoadPeriod.value] || 1
  const loadHoursByEmployee = new Map()

  for (const process of processes.value) {
    const monthlyExec = regularityToMonthlyCount(
      normalizeProcessRegularityCount(process),
      normalizeProcessRegularityUnit(process)
    )
    if (monthlyExec <= 0) continue

    const execInPeriod = monthlyExec * selectedFactor
    const version = getProcessLatestVersion(process)
    const steps = getVersionSteps(version)

    for (const step of steps) {
      if (!stepMatchesTypeFilter(step)) continue
      const actualMin = normalizeStepActualMin(step)
      if (!Number.isFinite(actualMin) || actualMin <= 0) continue
      const loads = getStepExecutorsWithPercent(step)
      for (const load of loads) {
        const employee = employeeById.value.get(load.employeeId)
        if (!employeeMatchesDepartmentFilter(employee)) continue
        const hours = (actualMin / 60) * (Number(load.percent || 0) / 100) * execInPeriod
        loadHoursByEmployee.set(load.employeeId, (loadHoursByEmployee.get(load.employeeId) || 0) + hours)
      }
    }
  }

  const rows = [...loadHoursByEmployee.entries()]
    .map(([employeeId, hours]) => {
      const e = employeeById.value.get(employeeId)
      const label = `${e?.last_name ?? e?.LastName ?? ""} ${e?.first_name ?? e?.FirstName ?? ""}`.trim() || `ID ${employeeId}`
      return { label, hours: Number(hours.toFixed(2)) }
    })
    .sort((a, b) => b.hours - a.hours)
    .slice(0, 15)

  return {
    labels: rows.map((r) => r.label),
    datasets: [
      {
        label: "Занятость (часы)",
        data: rows.map((r) => r.hours),
        backgroundColor: palette.teal,
        borderRadius: 6
      }
    ]
  }
})

const employeeLoadOptions = {
  responsive: true,
  maintainAspectRatio: false,
  indexAxis: "y",
  plugins: { legend: { position: "bottom" } },
  scales: {
    x: { beginAtZero: true, title: { display: true, text: "Часы" } }
  }
}
</script>

<style scoped>
.dashboards-page-container {
  margin-left: 240px;
  padding: 24px;
  width: calc(100% - 240px);
  box-sizing: border-box;
}

.dashboards-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  padding: 24px;
}

.dashboards-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.dashboards-header h2 {
  margin: 0;
}

.refresh-btn {
  border: none;
  background: var(--color-primary);
  color: #fff;
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-box {
  background: #fef2f2;
  color: #991b1b;
  border: 1px solid #fecaca;
  border-radius: 8px;
  padding: 10px 12px;
  margin-bottom: 12px;
}

.filters {
  margin-bottom: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.filters label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-width: 320px;
  font-weight: 500;
}

.filters select {
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid #cbd5e1;
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(340px, 1fr));
  gap: 14px;
}

.chart-card {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 12px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
}

.chart-card-wide {
  grid-column: 1 / -1;
}

.chart-card h3 {
  margin: 0;
  font-size: 16px;
}

.chart-subtitle {
  margin: 6px 0 10px;
  color: #64748b;
  font-size: 13px;
}

.chart-wrap {
  height: 330px;
}

@media (max-width: 1200px) {
  .dashboards-page-container {
    margin-left: 0;
    width: 100%;
    padding: 14px;
  }

  .charts-grid {
    grid-template-columns: 1fr;
  }

  .chart-card-wide {
    grid-column: auto;
  }
}
</style>
