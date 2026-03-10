<template>
  <div class="version-container">
    <div class="version-header">
      <h3>Версия {{ version.Version || "?" }}</h3>
      <button class="delete-version-btn" @click="showDeleteDialog = true">Удалить версию</button>
    </div>

    <div class="version-info">
      <p><b>Создана:</b> {{ formatDateTime(version.CreatedAt) }}</p>
      <p><b>Опубликована:</b> {{ version.IsPublished ? 'Да' : 'Нет' }}</p>
    </div>

    <div class="steps-header">
      <h4>Этапы процесса</h4>
      <div class="steps-header-actions">
        <div class="steps-view-tabs">
          <button
            :class="['steps-view-btn', { active: stepsViewMode === 'table' }]"
            @click="stepsViewMode = 'table'"
            type="button"
          >
            Таблица
          </button>
          <button
            :class="['steps-view-btn', { active: stepsViewMode === 'graph' }]"
            @click="stepsViewMode = 'graph'"
            type="button"
          >
            Граф
          </button>
        </div>
        <button class="add-step-btn" @click="openCreateStepModal">+ Добавить этап</button>
      </div>
    </div>

    <div v-if="stepsViewMode === 'table'" class="steps-table">
      <table class="steps-table-grid">
        <thead>
          <tr>
            <th>#</th>
            <th>Название этапа</th>
            <th>Тип этапа</th>
            <th>Исполнитель</th>
            <th>Занятость (%)</th>
            <th>Время на этапе (мин)</th>
            <th>Итоговая длительность (мин)</th>
            <th>Плановое время (мин)</th>
            <th>Действия</th>
          </tr>
        </thead>
        <tbody v-if="sortedSteps.length > 0">
          <template v-for="(step, index) in sortedSteps" :key="step.ID">
            <tr v-for="(executorRow, execIndex) in getStepExecutorRows(step)" :key="`${step.ID}-${execIndex}`">
              <td v-if="execIndex === 0" :rowspan="getStepExecutorRows(step).length" class="steps-cell-center">
                {{ getStepDisplayOrder(step, index) }}
              </td>
              <td v-if="execIndex === 0" :rowspan="getStepExecutorRows(step).length">
                {{ step.Name }}
              </td>
              <td v-if="execIndex === 0" :rowspan="getStepExecutorRows(step).length">
                {{ getStepTypeLabel(step.Type) }}
              </td>
              <td>{{ executorRow.fio }}</td>
              <td class="steps-cell-center">{{ formatPercent(executorRow.percent) }}</td>
              <td class="steps-cell-center">{{ getExecutorTimeOnStep(step, executorRow.percent) }}</td>
              <td v-if="execIndex === 0" :rowspan="getStepExecutorRows(step).length" class="steps-cell-center">
                {{ isTimeTrackableType(step.Type) ? formatMinutes(step.FinalDurationMin) : "—" }}
              </td>
              <td v-if="execIndex === 0" :rowspan="getStepExecutorRows(step).length" class="steps-cell-center">
                {{ isTimeTrackableType(step.Type) ? formatPlannedMinutes(step.Metrics?.PlannedTimeMin) : "—" }}
              </td>
              <td v-if="execIndex === 0" :rowspan="getStepExecutorRows(step).length">
                <div class="step-actions-wrap">
                  <button class="icon-btn edit" @click="openEditStepModal(step)">
                    <Pencil :size="16" />
                  </button>
                  <button class="icon-btn delete" @click="openDeleteStepDialog(step.ID)">
                    <Trash :size="16" />
                  </button>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>

      <div v-if="sortedSteps.length === 0" class="empty steps-empty">Этапов нет</div>
    </div>

    <div v-else class="steps-graph-wrap">
      <div v-if="sortedSteps.length === 0" class="empty steps-empty">Этапов нет</div>
      <div v-else class="steps-graph-layout">
        <div class="graph-canvas-wrap">
          <svg
            class="steps-graph-svg"
            :viewBox="`0 0 ${graphSvgWidth} ${graphSvgHeight}`"
            preserveAspectRatio="xMinYMin meet"
          >
            <defs>
              <marker
                id="stepArrow"
                markerWidth="10"
                markerHeight="8"
                refX="9"
                refY="4"
                orient="auto"
                markerUnits="strokeWidth"
              >
                <path d="M0,0 L10,4 L0,8 Z" fill="#94a3b8" />
              </marker>
            </defs>

            <polyline
              v-for="edge in graphEdges"
              :key="edge.id"
              :points="edge.points"
              fill="none"
              stroke="#cbd5e1"
              stroke-width="2"
              marker-end="url(#stepArrow)"
            />

            <g
              v-for="node in graphNodes"
              :key="node.id"
              class="graph-node"
              :class="{ active: Number(selectedGraphStep?.ID) === Number(node.id) }"
              @click="selectGraphStep(node.id)"
            >
              <rect
                :x="node.x"
                :y="node.y"
                :width="GRAPH_NODE_WIDTH"
                :height="GRAPH_NODE_HEIGHT"
                rx="10"
                ry="10"
              />
              <text :x="node.x + 10" :y="node.y + 20" class="graph-node-order">{{ node.order }}.</text>
              <text :x="node.x + 28" :y="node.y + 20" class="graph-node-title">{{ shortStepName(node.name) }}</text>
              <text :x="node.x + 10" :y="node.y + 38" class="graph-node-type">{{ getStepTypeLabel(node.type) }}</text>
            </g>
          </svg>
        </div>

        <div class="graph-info-panel" v-if="selectedGraphStep">
          <h4>{{ selectedGraphStep.Name }}</h4>
          <p><b>№:</b> {{ selectedGraphOrder }}</p>
          <p><b>Тип:</b> {{ getStepTypeLabel(selectedGraphStep.Type) }}</p>
          <p><b>Описание:</b> {{ selectedGraphStep.Description || "—" }}</p>
          <template v-if="isExecutorAllowedType(selectedGraphStep.Type)">
            <p><b>Исполнители:</b> {{ selectedGraphExecutorsDisplay }}</p>
            <p><b>Итоговая длительность:</b> {{ formatMinutes(selectedGraphStep.FinalDurationMin) }} мин</p>
            <p><b>Плановое время:</b> {{ formatPlannedMinutes(selectedGraphStep.Metrics?.PlannedTimeMin) }} мин</p>
          </template>
          <button class="save-btn graph-edit-btn" @click="openEditStepModal(selectedGraphStep)">Редактировать этап</button>
        </div>
      </div>
    </div>

    <div v-if="modalOpen" class="modal-overlay step-edit-overlay">
      <div class="modal modal-wide step-edit-modal">
        <div class="modal-title-row">
          <h3>{{ editMode ? "Редактировать этап" : "Новый этап" }}</h3>
          <button class="modal-close-btn" @click="closeStepModal" type="button" aria-label="Закрыть">
            <X :size="18" />
          </button>
        </div>

        <div class="edit-main-block">
          <div class="modal-columns">
            <div class="modal-column">
              <label>Название *</label>
              <input v-model="form.name" type="text" />

              <label>Тип *</label>
              <select v-model="form.type" @change="onStepTypeChange">
                <option disabled value="">Выберите тип</option>
                <option v-for="(label, code) in stepTypes" :key="code" :value="code">{{ label }}</option>
              </select>

              <label>Описание</label>
              <textarea v-model="form.description" rows="3"></textarea>
            </div>

            <div class="modal-column">
              <template v-if="isTimeTrackableType(form.type)">
                <label>Плановое время</label>
                <input v-model.number="form.plannedTimeMin" type="number" min="0" placeholder="минуты" />
                <small class="field-hint">Значение указывается в минутах</small>
              </template>

              <template v-if="editMode && isTimeTrackableType(form.type)">
                <label>Итоговая длительность (минут)</label>
                <input type="text" :value="finalDurationDisplay" readonly />
              </template>

              <template v-if="isExecutorAllowedType(form.type)">
                <label>Исполнители</label>
                <button class="picker-toggle" @click="showExecutorPicker = !showExecutorPicker" type="button">
                  {{ selectedExecutorsLabel }}
                </button>

                <div v-if="showExecutorPicker" class="executors-picker">
                  <input v-model="executorSearch" placeholder="Поиск по ФИО..." class="search-executors" />
                  <div
                    v-if="editMode && form.executorIds.length > 0"
                    class="executor-sum-hint"
                    :class="{ valid: executorWorkloadIsValid, invalid: !executorWorkloadIsValid }"
                  >
                    Сумма занятости: {{ executorWorkloadSumDisplay }}%
                    <span class="executor-sum-status">
                      {{ executorWorkloadIsValid ? "корректно" : "должно быть 100%" }}
                    </span>
                  </div>
                  <div class="executors-table-header" :class="{ 'executors-table-compact': !editMode }">
                    <span>Выбрать</span>
                    <span>Сотрудник</span>
                    <span v-if="editMode">Занятость (%)</span>
                  </div>
                  <div
                    class="executors-table-row"
                    :class="{ 'executors-table-compact': !editMode }"
                    v-for="e in filteredExecutors"
                    :key="getExecutorId(e)"
                  >
                    <span>
                      <input
                        type="checkbox"
                        :value="getExecutorId(e)"
                        v-model="form.executorIds"
                        @change="onExecutorCheckboxChange(getExecutorId(e), $event.target.checked)"
                      />
                    </span>
                    <span>{{ formatExecutorFullName(e) }}</span>
                    <span v-if="editMode">
                      <input
                        type="number"
                        min="0"
                        max="100"
                        step="0.01"
                        :disabled="!isExecutorSelected(getExecutorId(e))"
                        :value="getExecutorPercent(getExecutorId(e))"
                        @input="setExecutorPercent(getExecutorId(e), $event.target.value)"
                        class="workload-input"
                      />
                    </span>
                  </div>
                  <div v-if="filteredExecutors.length === 0" class="empty">Ничего не найдено</div>
                </div>
              </template>
              <div v-else class="empty">
                Исполнители доступны только для типов этапа "Операция" и "Подпроцесс".
              </div>
            </div>
          </div>
        </div>

        <div v-if="editMode && isTimeTrackableType(form.type)" class="indicators-block">
          <div class="indicators-header">
            <h4>Показатели</h4>
            <label class="stats-checkbox">
              <input
                type="checkbox"
                :checked="form.useStatistics"
                @change="onStatisticsToggle($event.target.checked)"
              />
              Статистические данные
            </label>
          </div>

          <div v-if="form.useStatistics" class="stats-pairs">
            <div class="stats-pair">
              <label>мин. (минут)<input type="number" min="0" v-model.number="form.minTimeMin" /></label>
              <label>мин. (%)<input type="number" step="0.01" v-model.number="form.minPercent" /></label>
            </div>
            <div class="stats-pair">
              <label>ср. (минут)<input type="number" min="0" v-model.number="form.avgTimeMin" /></label>
              <label>ср. (%)<input type="number" step="0.01" v-model.number="form.avgPercent" /></label>
            </div>
            <div class="stats-pair">
              <label>макс. (минут)<input type="number" min="0" v-model.number="form.maxTimeMin" /></label>
              <label>макс. (%)<input type="number" step="0.01" v-model.number="form.maxPercent" /></label>
            </div>
            <div class="stats-pair single">
              <label>
                взвешенное среднее (минут)
                <input type="text" :value="calculatedWeightedAvgDisplay" readonly />
              </label>
            </div>
          </div>

          <div v-else>
            <p v-if="activeMeasurement" class="active-hint">
              Активен замер №{{ activeMeasurement.MeasurementNumber || activeMeasurement.ID }}
              <span v-if="isActiveMeasurementPaused">(на паузе)</span>
            </p>
            <p v-if="!activeMeasurement && editedMeasurements.length >= 3" class="active-hint warning-hint">
              Достигнут лимит: не более 3 замеров на этап.
            </p>

            <div class="measure-controls">
              <button class="save-btn" :disabled="!!activeMeasurement || editedMeasurements.length >= 3" @click="startMeasurement">Старт</button>
              <button class="cancel-btn" :disabled="!activeMeasurement" @click="pauseOrResumeMeasurement">
                {{ isActiveMeasurementPaused ? "Продолжить" : "Пауза" }}
              </button>
              <button class="delete-version-btn" :disabled="!activeMeasurement" @click="stopMeasurement">Стоп</button>
            </div>

            <div class="measurements-table">
              <div class="measurements-header">
                <span>#</span>
                <span>Старт</span>
                <span>Финиш</span>
                <span>Пауза (сек)</span>
                <span>Длительность (сек)</span>
                <span>Действия</span>
              </div>
              <div v-for="m in editedMeasurements" :key="m.ID" class="measurements-row">
                <span>{{ m.MeasurementNumber || m.ID }}</span>
                <span>{{ formatDateTime(m.StartedAt) || '-' }}</span>
                <span>{{ formatDateTime(m.FinishedAt) || '-' }}</span>
                <span>{{ getMeasurementPausedSeconds(m) }}</span>
                <span>{{ m.DurationSec ?? '-' }}</span>
                <span>
                  <button class="icon-btn delete" @click="openDeleteMeasurementDialog(m.ID)">
                    <Trash :size="14" />
                  </button>
                </span>
              </div>
              <div v-if="editedMeasurements.length === 0" class="empty">Замеров нет</div>
            </div>
          </div>
        </div>

        <div v-else-if="editMode" class="indicators-block">
          <div class="empty">Для этого типа этапа настройки замеров времени и статистики недоступны.</div>
        </div>

        <div class="modal-actions">
          <button class="save-btn" @click="saveStep">Сохранить</button>
        </div>
      </div>
    </div>

    <div v-if="showDeleteDialog" class="modal-overlay">
      <div class="modal">
        <h3>Удалить версию?</h3>
        <p>Вы действительно хотите удалить версию <b>{{ version.Version }}</b> ?</p>
        <div class="modal-actions">
          <button class="cancel-btn" @click="showDeleteDialog = false">Отмена</button>
          <button class="save-btn" @click="deleteVersion">Удалить</button>
        </div>
      </div>
    </div>

    <div v-if="showDeleteMeasurementDialog" class="modal-overlay">
      <div class="modal">
        <h3>Удалить замер?</h3>
        <p>Вы действительно хотите удалить выбранный замер времени?</p>
        <div class="modal-actions">
          <button class="cancel-btn" @click="cancelDeleteMeasurement">Отмена</button>
          <button class="save-btn" @click="confirmDeleteMeasurement">Удалить</button>
        </div>
      </div>
    </div>

    <div v-if="showDeleteStepDialog" class="modal-overlay">
      <div class="modal">
        <h3>Удалить этап?</h3>
        <p>Вы действительно хотите удалить выбранный этап процесса?</p>
        <div class="modal-actions">
          <button class="cancel-btn" @click="cancelDeleteStep">Отмена</button>
          <button class="save-btn" @click="confirmDeleteStep">Удалить</button>
        </div>
      </div>
    </div>

    <div v-if="toast.show" class="toast">{{ toast.message }}</div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue"
import { Pencil, Trash, X } from "lucide-vue-next"
import api from "../api/axios"

const props = defineProps({ version: Object, processId: Number })
const emit = defineEmits(["refresh"])

const showDeleteDialog = ref(false)
const showDeleteMeasurementDialog = ref(false)
const showDeleteStepDialog = ref(false)
const modalOpen = ref(false)
const showExecutorPicker = ref(false)
const stepsViewMode = ref("table")
const editMode = ref(false)
const editingStepId = ref(null)
const editedMeasurements = ref([])
const deletingMeasurementId = ref(null)
const deletingStepId = ref(null)
const selectedGraphStepId = ref(null)

const form = ref({
  name: "",
  type: "",
  description: "",
  executorIds: [],
  executorPercents: {},
  plannedTimeMin: 0,
  finalDurationMin: 0,
  useStatistics: false,
  minTimeMin: 0,
  minPercent: 0,
  avgTimeMin: 0,
  avgPercent: 0,
  maxTimeMin: 0,
  maxPercent: 0
})

const executorSearch = ref("")
const executors = ref([])
const toast = ref({ show: false, message: "" })

const stepTypes = {
  START: "Стартовое событие",
  END: "Конечное событие",
  INTERMEDIATE: "Промежуточное событие",
  SUBPROCESS: "Подпроцесс",
  OPERATION: "Операция",
  CONDITION: "Условие"
}
const TIME_TRACKABLE_STEP_TYPES = ["SUBPROCESS", "OPERATION"]
const isTimeTrackableType = (type) => TIME_TRACKABLE_STEP_TYPES.includes(type)
const isExecutorAllowedType = (type) => TIME_TRACKABLE_STEP_TYPES.includes(type)
const GRAPH_NODE_WIDTH = 220
const GRAPH_NODE_HEIGHT = 60
const GRAPH_COLS = 1
const GRAPH_GAP_X = 90
const GRAPH_GAP_Y = 80
const GRAPH_PADDING_X = 24
const GRAPH_PADDING_Y = 24

const getStepId = (step) => Number(step?.ID)
const getExecutorId = (executor) => Number(executor?.id ?? executor?.ID)
const getExecutorLastName = (executor) => executor?.last_name ?? executor?.LastName ?? ""
const getExecutorFirstName = (executor) => executor?.first_name ?? executor?.FirstName ?? ""
const getExecutorMiddleName = (executor) => executor?.middle_name ?? executor?.MiddleName ?? ""

const formatExecutorFullName = (executor) =>
  `${getExecutorLastName(executor)} ${getExecutorFirstName(executor)} ${getExecutorMiddleName(executor)}`.trim()

const showToast = (message) => {
  toast.value.message = message
  toast.value.show = true
  setTimeout(() => {
    toast.value.show = false
  }, 2500)
}

const getStatisticsPercentSum = () => {
  return Number(form.value.minPercent || 0) +
    Number(form.value.avgPercent || 0) +
    Number(form.value.maxPercent || 0)
}

const resetForm = () => {
  form.value = {
    name: "",
    type: "",
    description: "",
    executorIds: [],
    executorPercents: {},
    plannedTimeMin: 0,
    finalDurationMin: 0,
    useStatistics: false,
    minTimeMin: 0,
    minPercent: 0,
    avgTimeMin: 0,
    avgPercent: 0,
    maxTimeMin: 0,
    maxPercent: 0
  }
  editedMeasurements.value = []
  executorSearch.value = ""
  showExecutorPicker.value = false
  editMode.value = false
  editingStepId.value = null
  deletingMeasurementId.value = null
  showDeleteMeasurementDialog.value = false
  deletingStepId.value = null
  showDeleteStepDialog.value = false
}

const closeStepModal = () => {
  modalOpen.value = false
  resetForm()
}

const selectedExecutorsLabel = computed(() => {
  const selectedCount = form.value.executorIds.length
  if (selectedCount === 0) return "Выбрать исполнителей"
  return `Выбрано: ${selectedCount}`
})

const sortedSteps = computed(() => {
  const steps = Array.isArray(props.version?.Steps) ? [...props.version.Steps] : []
  return steps.sort((a, b) => {
    const orderA = Number(a?.StepOrder)
    const orderB = Number(b?.StepOrder)
    const hasOrderA = Number.isFinite(orderA) && orderA > 0
    const hasOrderB = Number.isFinite(orderB) && orderB > 0

    if (hasOrderA && hasOrderB && orderA !== orderB) return orderA - orderB
    if (hasOrderA !== hasOrderB) return hasOrderA ? -1 : 1

    const idA = Number(a?.ID) || 0
    const idB = Number(b?.ID) || 0
    return idA - idB
  })
})

const graphNodes = computed(() => {
  return sortedSteps.value.map((step, index) => {
    const row = Math.floor(index / GRAPH_COLS)
    const col = index % GRAPH_COLS
    return {
      id: Number(step.ID),
      name: step.Name || "Без названия",
      type: step.Type || "",
      order: getStepDisplayOrder(step, index),
      x: GRAPH_PADDING_X + col * (GRAPH_NODE_WIDTH + GRAPH_GAP_X),
      y: GRAPH_PADDING_Y + row * (GRAPH_NODE_HEIGHT + GRAPH_GAP_Y)
    }
  })
})

const graphEdges = computed(() => {
  const nodes = graphNodes.value
  if (nodes.length < 2) return []
  return nodes.slice(0, -1).map((fromNode, index) => {
    const toNode = nodes[index + 1]
    const fromX = fromNode.x + GRAPH_NODE_WIDTH / 2
    const fromY = fromNode.y + GRAPH_NODE_HEIGHT
    const toX = toNode.x + GRAPH_NODE_WIDTH / 2
    const toY = toNode.y
    const midY = Math.round((fromY + toY) / 2)
    return {
      id: `${fromNode.id}-${toNode.id}`,
      points: `${fromX},${fromY} ${fromX},${midY} ${toX},${midY} ${toX},${toY}`
    }
  })
})

const graphSvgWidth = computed(() => {
  const cols = Math.min(GRAPH_COLS, Math.max(1, sortedSteps.value.length))
  return GRAPH_PADDING_X * 2 + cols * GRAPH_NODE_WIDTH + Math.max(0, cols - 1) * GRAPH_GAP_X
})

const graphSvgHeight = computed(() => {
  const rows = Math.max(1, Math.ceil(sortedSteps.value.length / GRAPH_COLS))
  return GRAPH_PADDING_Y * 2 + rows * GRAPH_NODE_HEIGHT + Math.max(0, rows - 1) * GRAPH_GAP_Y
})

const selectedGraphStep = computed(() => {
  const selected = sortedSteps.value.find((s) => Number(s.ID) === Number(selectedGraphStepId.value))
  return selected || sortedSteps.value[0] || null
})

const selectedGraphOrder = computed(() => {
  if (!selectedGraphStep.value) return "—"
  const idx = sortedSteps.value.findIndex((s) => Number(s.ID) === Number(selectedGraphStep.value.ID))
  return idx >= 0 ? getStepDisplayOrder(selectedGraphStep.value, idx) : "—"
})

const selectedGraphExecutorsDisplay = computed(() => {
  const step = selectedGraphStep.value
  if (!step) return "—"
  if (!isExecutorAllowedType(step.Type)) return "Не применимо для данного типа"
  const rows = getStepExecutorRows(step).filter((r) => r.fio !== "—")
  if (rows.length === 0) return "—"
  return rows.map((r) => `${r.fio} (${formatPercent(r.percent)}%)`).join(", ")
})

const selectGraphStep = (stepId) => {
  selectedGraphStepId.value = Number(stepId)
}

const shortStepName = (name) => {
  const str = String(name || "").trim()
  if (str.length <= 20) return str
  return `${str.slice(0, 17)}...`
}

const executorsById = computed(() => {
  const m = new Map()
  for (const e of executors.value) {
    const id = getExecutorId(e)
    if (Number.isFinite(id)) m.set(id, e)
  }
  return m
})

const syncEditedStepFromBackend = async () => {
  if (!editingStepId.value) return

  const token = localStorage.getItem("jwt")
  const res = await api.get(`/processes/measurements?stepId=${editingStepId.value}`, {
    headers: { Authorization: `Bearer ${token}` }
  })

  editedMeasurements.value = [...(res?.data || [])].sort((a, b) => Number(a.ID) - Number(b.ID))
}

const activeMeasurement = computed(() => {
  const active = editedMeasurements.value.filter((m) => !m.FinishedAt)
  if (active.length === 0) return null
  return active[active.length - 1]
})

const isActiveMeasurementPaused = computed(() => {
  const m = activeMeasurement.value
  if (!m) return false
  return (m.Pauses || []).some((p) => !p.PauseEnd)
})

const getMeasurementPausedSeconds = (measurement) => {
  if (!measurement) return 0

  let seconds = Number(measurement.PausedSeconds || 0)
  const isCurrentActive = Number(activeMeasurement.value?.ID) === Number(measurement.ID)
  if (!isCurrentActive || !isActiveMeasurementPaused.value) return seconds

  const openPause = (measurement.Pauses || []).find((p) => !p.PauseEnd)
  if (!openPause) return seconds

  const started = new Date(openPause.PauseStart).getTime()
  if (Number.isNaN(started)) return seconds
  const now = Date.now()
  const extra = Math.max(0, Math.round((now - started) / 1000))
  return seconds + extra
}

const openCreateStepModal = () => {
  resetForm()
  modalOpen.value = true
}

const openEditStepModal = (step) => {
  const metrics = step.Metrics || {}
  const stat = metrics.TimeStatistics || {}
  const timeTrackable = isTimeTrackableType(step.Type || "")

  editMode.value = true
  editingStepId.value = getStepId(step)
  form.value = {
    name: step.Name || "",
    type: step.Type || "",
    description: step.Description || "",
    executorIds: isExecutorAllowedType(step.Type)
      ? (step.StepExecutors || []).map((se) => Number(se.EmployeeID)).filter(Number.isFinite)
      : [],
    executorPercents: isExecutorAllowedType(step.Type)
      ? Object.fromEntries((step.StepExecutors || []).map((se) => [Number(se.EmployeeID), Number(se.WorkloadPercent || 0)]))
      : {},
    plannedTimeMin: timeTrackable ? (metrics.PlannedTimeMin ?? 0) : 0,
    finalDurationMin: Number(step.FinalDurationMin || 0),
    useStatistics: timeTrackable ? !!step.Metrics : false,
    minTimeMin: timeTrackable ? (stat.MinTime ?? 0) : 0,
    minPercent: timeTrackable ? (stat.MinPercent ?? 0) : 0,
    avgTimeMin: timeTrackable ? (stat.AvgTime ?? 0) : 0,
    avgPercent: timeTrackable ? (stat.AvgPercent ?? 0) : 0,
    maxTimeMin: timeTrackable ? (stat.MaxTime ?? 0) : 0,
    maxPercent: timeTrackable ? (stat.MaxPercent ?? 0) : 0
  }
  editedMeasurements.value = [...(step.Measurements || [])].sort((a, b) => Number(a.ID) - Number(b.ID))

  executorSearch.value = ""
  showExecutorPicker.value = false
  modalOpen.value = true
  syncEditedStepFromBackend()
}

const getStepTypeLabel = (type) => stepTypes[type] || type || "—"

const getStepDisplayOrder = (step, fallbackIndex) => {
  const order = Number(step?.StepOrder)
  if (Number.isFinite(order) && order > 0) return order
  return fallbackIndex + 1
}

const getStepExecutorRows = (step) => {
  if (!isExecutorAllowedType(step?.Type)) {
    return [{ fio: "—", percent: null }]
  }

  const loads = step?.StepExecutors || []
  if (!Array.isArray(loads) || loads.length === 0) {
    return [{ fio: "—", percent: null }]
  }

  return loads.map((item) => {
    const executor = item.Employee || executorsById.value.get(Number(item.EmployeeID))
    const fio = executor ? formatExecutorFullName(executor) : `ID ${item.EmployeeID}`
    return {
      fio,
      percent: Number(item.WorkloadPercent || 0)
    }
  })
}

const formatPercent = (value) => {
  if (value === null || value === undefined || value === "") return "—"
  const n = Number(value)
  if (!Number.isFinite(n)) return "—"
  return n.toFixed(2)
}

const formatMinutes = (value) => {
  const n = Number(value)
  if (!Number.isFinite(n)) return "—"
  return n.toFixed(2)
}

const formatPlannedMinutes = (value) => {
  const n = Number(value)
  if (!Number.isFinite(n)) return "—"
  return String(Math.round(n))
}

const getExecutorTimeOnStep = (step, workloadPercent) => {
  if (workloadPercent === null || workloadPercent === undefined || workloadPercent === "") return "—"
  const total = Number(step?.FinalDurationMin)
  const percent = Number(workloadPercent)
  if (!Number.isFinite(total) || !Number.isFinite(percent)) return "—"
  return ((total * percent) / 100).toFixed(2)
}

const isExecutorSelected = (executorId) => form.value.executorIds.includes(Number(executorId))

const getExecutorPercent = (executorId) => {
  const value = form.value.executorPercents?.[executorId]
  if (value === undefined || value === null || value === "") return 0
  return value
}

const setExecutorPercent = (executorId, value) => {
  if (!form.value.executorPercents) form.value.executorPercents = {}
  form.value.executorPercents[executorId] = value
}

const onExecutorCheckboxChange = (executorId, checked) => {
  if (!form.value.executorPercents) form.value.executorPercents = {}
  if (checked && form.value.executorPercents[executorId] === undefined) {
    form.value.executorPercents[executorId] = 0
  }
  if (!checked) {
    delete form.value.executorPercents[executorId]
  }
}

const onStepTypeChange = () => {
  if (!isExecutorAllowedType(form.value.type)) {
    form.value.executorIds = []
    form.value.executorPercents = {}
    showExecutorPicker.value = false
  }
}

const fetchExecutors = async () => {
  const token = localStorage.getItem("jwt")
  try {
    const res = await api.get("/employees", { headers: { Authorization: `Bearer ${token}` } })
    executors.value = res.data
  } catch (err) {
    console.error(err)
  }
}

onMounted(fetchExecutors)

const filteredExecutors = computed(() => {
  const q = executorSearch.value.toLowerCase()
  const matched = executors.value.filter(
    (e) => Number.isFinite(getExecutorId(e)) && formatExecutorFullName(e).toLowerCase().includes(q)
  )

  return matched.sort((a, b) => {
    const aSelected = isExecutorSelected(getExecutorId(a)) ? 1 : 0
    const bSelected = isExecutorSelected(getExecutorId(b)) ? 1 : 0

    if (aSelected !== bSelected) return bSelected - aSelected
    return formatExecutorFullName(a).localeCompare(formatExecutorFullName(b), "ru")
  })
})

const calculatedWeightedAvg = computed(() => {
  const minTime = Number(form.value.minTimeMin)
  const minPercent = Number(form.value.minPercent)
  const avgTime = Number(form.value.avgTimeMin)
  const avgPercent = Number(form.value.avgPercent)
  const maxTime = Number(form.value.maxTimeMin)
  const maxPercent = Number(form.value.maxPercent)

  if (![minTime, minPercent, avgTime, avgPercent, maxTime, maxPercent].every(Number.isFinite)) {
    return null
  }

  // Percent values come as 20, 30, ... (not 0.2), so divide by 100.
  return (minTime * minPercent + avgTime * avgPercent + maxTime * maxPercent) / 100
})

const calculatedWeightedAvgDisplay = computed(() => {
  if (calculatedWeightedAvg.value === null) return "—"
  return calculatedWeightedAvg.value.toFixed(2)
})

const onStatisticsToggle = (enabled) => {
  if (!isTimeTrackableType(form.value.type)) {
    form.value.useStatistics = false
    return
  }
  if (enabled && editedMeasurements.value.length > 0) {
    showToast("Нельзя включить статистику: у этапа уже есть замеры")
    return
  }
  form.value.useStatistics = enabled
}

const measuredAverageDurationMin = computed(() => {
  const finished = editedMeasurements.value.filter(
    (m) => m.FinishedAt && Number.isFinite(Number(m.DurationSec))
  )
  if (finished.length === 0) return null
  const sumSec = finished.reduce((acc, m) => acc + Number(m.DurationSec || 0), 0)
  return (sumSec / finished.length) / 60
})

const executorWorkloadSum = computed(() => {
  return form.value.executorIds
    .map((id) => Number(form.value.executorPercents?.[id]))
    .reduce((acc, v) => acc + (Number.isFinite(v) ? v : 0), 0)
})

const executorWorkloadSumDisplay = computed(() => executorWorkloadSum.value.toFixed(2))
const executorWorkloadIsValid = computed(() => Math.abs(executorWorkloadSum.value - 100) <= 0.0001)

const finalDurationMin = computed(() => {
  if (measuredAverageDurationMin.value !== null) return measuredAverageDurationMin.value
  if (form.value.useStatistics && calculatedWeightedAvg.value !== null) return calculatedWeightedAvg.value
  if (Number.isFinite(Number(form.value.finalDurationMin))) return Number(form.value.finalDurationMin)
  return null
})

const finalDurationDisplay = computed(() => {
  if (finalDurationMin.value === null) return "—"
  return finalDurationMin.value.toFixed(2)
})

const saveStep = async () => {
  if (!form.value.name || !form.value.type) {
    alert("Заполните обязательные поля")
    return
  }

  const canUseTimeSettings = isTimeTrackableType(form.value.type)
  const canUseExecutors = isExecutorAllowedType(form.value.type)
  if (!canUseTimeSettings) {
    form.value.useStatistics = false
  }
  if (!canUseExecutors) {
    form.value.executorIds = []
    form.value.executorPercents = {}
  }

  if (canUseTimeSettings && form.value.useStatistics) {
    if (editedMeasurements.value.length > 0) {
      showToast("Нельзя сохранить статистику: у этапа уже есть замеры")
      return
    }

    const requiredStatValues = [
      form.value.minTimeMin,
      form.value.minPercent,
      form.value.avgTimeMin,
      form.value.avgPercent,
      form.value.maxTimeMin,
      form.value.maxPercent
    ]
    const hasEmpty = requiredStatValues.some((v) => v === null || v === undefined || v === "")
    if (hasEmpty) {
      showToast("Заполните все поля статистических данных")
      return
    }

    const percentSum = getStatisticsPercentSum()
    if (Math.abs(percentSum - 100) > 0.0001) {
      showToast("Сумма процентов статистики должна быть ровно 100%")
      return
    }

    if (calculatedWeightedAvg.value === null) {
      showToast("Некорректные значения для расчета средневзвешенного")
      return
    }
  }

  const executorIds = canUseExecutors
    ? form.value.executorIds.map(Number).filter(Number.isFinite)
    : []
  if (editMode.value && executorIds.length > 0) {
    const invalid = executorIds.find((id) => {
      const p = Number(form.value.executorPercents?.[id])
      return !Number.isFinite(p) || p < 0 || p > 100
    })
    if (invalid) {
      showToast("Укажите корректную занятость 0..100 для каждого выбранного исполнителя")
      return
    }

    if (Math.abs(executorWorkloadSum.value - 100) > 0.0001) {
      showToast("Сумма занятости исполнителей должна быть ровно 100%")
      return
    }
  }

  try {
    const token = localStorage.getItem("jwt")

    if (editMode.value && editingStepId.value) {
      const executorLoads = executorIds.map((id) => ({
        employeeId: id,
        workloadPercent: Number(form.value.executorPercents?.[id] ?? 0)
      }))

      await api.put(
        `/processes/steps/${editingStepId.value}`,
        {
          Name: form.value.name,
          Type: form.value.type,
          Description: form.value.description,
          ExecutorLoads: canUseExecutors
            ? executorLoads.map((x) => ({
                employeeId: x.employeeId,
                workloadPercent: x.workloadPercent
              }))
            : [],
          Metrics: canUseTimeSettings
            ? {
                PlannedTimeMin: Number(form.value.plannedTimeMin) || 0,
                TimeStatistics: form.value.useStatistics
                  ? {
                      MinTime: Number(form.value.minTimeMin) || 0,
                      MinPercent: Number(form.value.minPercent) || 0,
                      AvgTime: Number(form.value.avgTimeMin) || 0,
                      AvgPercent: Number(form.value.avgPercent) || 0,
                      MaxTime: Number(form.value.maxTimeMin) || 0,
                      MaxPercent: Number(form.value.maxPercent) || 0,
                      WeightedAvg: Number(calculatedWeightedAvg.value) || 0
                    }
                  : null
              }
            : null
        },
        { headers: { Authorization: `Bearer ${token}` } }
      )
      showToast("Этап обновлен")
    } else {
      await api.post(
        "/processes/steps",
        {
          processVersionId: props.version.ID,
          name: form.value.name,
          type: form.value.type,
          description: form.value.description,
          executorIds: canUseExecutors ? executorIds : []
        },
        { headers: { Authorization: `Bearer ${token}` } }
      )
      showToast("Этап добавлен")
      closeStepModal()
    }

    emit("refresh")
  } catch (err) {
    console.error(err)
    showToast("Ошибка сохранения этапа")
  }
}

const openDeleteStepDialog = (stepId) => {
  deletingStepId.value = stepId
  showDeleteStepDialog.value = true
}

const cancelDeleteStep = () => {
  deletingStepId.value = null
  showDeleteStepDialog.value = false
}

const confirmDeleteStep = async () => {
  if (!deletingStepId.value) return

  const token = localStorage.getItem("jwt")
  try {
    await api.delete(`/processes/steps/${deletingStepId.value}`, { headers: { Authorization: `Bearer ${token}` } })
    showToast("Этап удален")
    cancelDeleteStep()
    emit("refresh")
  } catch (err) {
    console.error(err)
    showToast("Ошибка удаления этапа")
  }
}

const startMeasurement = async () => {
  if (!editingStepId.value) return
  if (!isTimeTrackableType(form.value.type)) return

  const token = localStorage.getItem("jwt")
  try {
    const res = await api.post(`/processes/measurements/start?stepId=${editingStepId.value}`, null, {
      headers: { Authorization: `Bearer ${token}` }
    })

    if (res?.data) {
      editedMeasurements.value = [...editedMeasurements.value, { ...res.data, Pauses: [] }]
        .sort((a, b) => Number(a.ID) - Number(b.ID))
    }

    await syncEditedStepFromBackend()
    showToast("Замер запущен")
  } catch (err) {
    console.error(err)
    showToast(err?.response?.data?.error || "Ошибка запуска замера")
  }
}

const pauseOrResumeMeasurement = async () => {
  if (!isTimeTrackableType(form.value.type)) return
  const active = activeMeasurement.value
  if (!active) return

  const token = localStorage.getItem("jwt")
  const endpoint = isActiveMeasurementPaused.value ? "resume" : "pause"

  try {
    await api.post(`/processes/measurements/${endpoint}?measurementId=${active.ID}`, null, {
      headers: { Authorization: `Bearer ${token}` }
    })

    await syncEditedStepFromBackend()
    showToast(endpoint === "resume" ? "Замер продолжен" : "Замер на паузе")
  } catch (err) {
    console.error(err)
    showToast("Ошибка обновления замера")
  }
}

const stopMeasurement = async () => {
  if (!isTimeTrackableType(form.value.type)) return
  const active = activeMeasurement.value
  if (!active) return

  const token = localStorage.getItem("jwt")

  try {
    await api.post(`/processes/measurements/finish?measurementId=${active.ID}`, null, {
      headers: { Authorization: `Bearer ${token}` }
    })

    await syncEditedStepFromBackend()
    showToast("Замер остановлен")
  } catch (err) {
    console.error(err)
    showToast("Ошибка остановки замера")
  }
}

const openDeleteMeasurementDialog = (measurementId) => {
  deletingMeasurementId.value = measurementId
  showDeleteMeasurementDialog.value = true
}

const cancelDeleteMeasurement = () => {
  deletingMeasurementId.value = null
  showDeleteMeasurementDialog.value = false
}

const confirmDeleteMeasurement = async () => {
  if (!deletingMeasurementId.value) return

  const token = localStorage.getItem("jwt")
  try {
    await api.delete(`/processes/measurements/${deletingMeasurementId.value}`, {
      headers: { Authorization: `Bearer ${token}` }
    })

    editedMeasurements.value = editedMeasurements.value.filter((m) => Number(m.ID) !== Number(deletingMeasurementId.value))

    await syncEditedStepFromBackend()
    cancelDeleteMeasurement()
    showToast("Замер удален")
  } catch (err) {
    console.error(err)
    showToast("Ошибка удаления замера")
  }
}

const deleteVersion = async () => {
  const token = localStorage.getItem("jwt")
  try {
    await api.delete(`/processes/versions/${props.version.ID}`, { headers: { Authorization: `Bearer ${token}` } })
    showDeleteDialog.value = false
    showToast("Версия удалена")
    emit("refresh")
  } catch (err) {
    console.error(err)
    showToast("Ошибка удаления версии")
  }
}

const formatDateTime = (iso) => {
  if (!iso) return ""
  const d = new Date(iso)
  return `${String(d.getDate()).padStart(2, "0")}.${String(d.getMonth() + 1).padStart(2, "0")}.${d.getFullYear()} ${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`
}
</script>

<style scoped>
.modal.modal-wide.step-edit-modal {
  width: calc(100vw - 240px - 32px);
  max-width: none;
  max-height: calc(100vh - 32px);
  overflow-y: auto;
}

.step-edit-overlay {
  justify-content: center;
  padding: 16px;
}

.modal-columns {
  display: flex;
  gap: 20px;
}

.modal-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.modal-title-row h3 {
  margin: 0;
}

.modal-close-btn {
  border: none;
  background: var(--color-soft-bg);
  border-radius: 6px;
  padding: 6px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.modal-close-btn:hover {
  background: var(--color-muted-bg);
}

.modal-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.field-hint {
  margin-top: -4px;
  color: #6b7280;
  font-size: 12px;
}

.picker-toggle {
  border: 1px solid #d1d5db;
  background: #fff;
  border-radius: 6px;
  padding: 8px 10px;
  text-align: left;
  cursor: pointer;
}

.executors-picker {
  border: 1px solid var(--color-muted-bg);
  border-radius: 8px;
  padding: 8px;
  max-height: 220px;
  overflow-y: auto;
  overflow-x: hidden;
  background: #fafafa;
}

.executors-table-header,
.executors-table-row {
  display: grid;
  grid-template-columns: 80px 1fr 120px;
  gap: 8px;
  align-items: center;
  padding: 6px 4px;
}

.executors-table-compact {
  grid-template-columns: 80px 1fr;
}

.executors-table-header {
  font-weight: 600;
  border-bottom: 1px solid #eee;
}

.workload-input {
  width: 100%;
  max-width: 110px;
}

.executor-sum-hint {
  margin-top: 8px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.executor-sum-hint.valid {
  color: #166534;
}

.executor-sum-hint.invalid {
  color: #b91c1c;
}

.executor-sum-status {
  font-weight: 600;
}

.search-executors {
  padding: 6px 10px;
  border-radius: 6px;
  border: 1px solid #ccc;
  margin-bottom: 8px;
}

.measure-block {
  margin-top: 8px;
  border-top: 1px solid #ececec;
  padding-top: 10px;
}

.edit-main-block {
  border: 1px solid #ececec;
  border-radius: 10px;
  padding: 14px;
  background: #fafafa;
}

.indicators-block {
  margin-top: 14px;
  border: 1px solid #ececec;
  border-radius: 10px;
  padding: 14px;
}

.indicators-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.indicators-header h4 {
  margin: 0;
}

.stats-checkbox {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.stats-pairs {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.stats-pair {
  display: grid;
  grid-template-columns: repeat(2, minmax(180px, 1fr));
  gap: 10px;
}

.stats-pair label {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
}

.stats-pair.single {
  grid-template-columns: minmax(220px, 280px);
}

.measure-controls {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}

.active-hint {
  margin: 0 0 10px;
  font-size: 13px;
  color: #2563eb;
  font-weight: 500;
}

.warning-hint {
  color: #b45309;
}

.measurements-table {
  border: 1px solid #ececec;
  border-radius: 8px;
  overflow: hidden;
}

.measurements-header,
.measurements-row {
  display: grid;
  grid-template-columns: 60px 1.2fr 1.2fr 130px 160px 90px;
  gap: 8px;
  align-items: center;
  padding: 8px 10px;
}

.measurements-header {
  background: #f9fafb;
  font-weight: 600;
  border-bottom: 1px solid #ececec;
}

.measurements-row {
  border-bottom: 1px solid var(--color-soft-bg);
}

.measurements-row:last-child {
  border-bottom: none;
}

.empty {
  color: #888;
  font-style: italic;
  padding: 8px 10px;
}

.modal input[type="text"],
.modal input[type="date"],
.modal input[type="number"],
.modal textarea,
.modal select {
  padding: 10px;
  border-radius: 6px;
  border: 1px solid #ccc;
  transition: border-color 0.2s;
  font-size: 14px;
  font-family: "Inter", sans-serif;
}

.modal input[type="text"]:focus,
.modal select:focus,
.modal textarea:focus,
.modal input[type="number"]:focus,
.search-executors:focus {
  border-color: var(--color-primary);
  outline: none;
}

.version-container {
  margin-top: 20px;
}

.version-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.steps-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
}

.steps-header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.steps-view-tabs {
  display: inline-flex;
  align-items: center;
  background: #f3f4f6;
  border-radius: 8px;
  padding: 3px;
}

.steps-view-btn {
  border: none;
  background: transparent;
  border-radius: 6px;
  padding: 6px 10px;
  font-size: 13px;
  cursor: pointer;
  color: #374151;
}

.steps-view-btn.active {
  background: white;
  color: #111827;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
}

.steps-table {
  margin-top: 10px;
  border: 1px solid var(--color-muted-bg);
  border-radius: 10px;
  overflow: auto;
}

.steps-graph-wrap {
  margin-top: 10px;
  border: 1px solid var(--color-muted-bg);
  border-radius: 10px;
  padding: 12px;
  background: #fafbfc;
}

.steps-graph-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 12px;
}

.graph-canvas-wrap {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: white;
  overflow: auto;
  min-height: 420px;
}

.steps-graph-svg {
  width: 100%;
  min-width: 420px;
  height: 100%;
  min-height: 420px;
}

.graph-node {
  cursor: pointer;
}

.graph-node rect {
  fill: #eef2ff;
  stroke: #c7d2fe;
  stroke-width: 1.5;
  transition: fill 0.15s, stroke-color 0.15s;
}

.graph-node.active rect {
  fill: #e0e7ff;
  stroke: #6366f1;
}

.graph-node:hover rect {
  fill: #e5edff;
  stroke: #818cf8;
}

.graph-node-order {
  fill: #312e81;
  font-size: 13px;
  font-weight: 600;
}

.graph-node-title {
  fill: #111827;
  font-size: 13px;
  font-weight: 600;
}

.graph-node-type {
  fill: #475569;
  font-size: 12px;
}

.graph-info-panel {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: white;
  padding: 12px;
  align-self: start;
  position: sticky;
  top: 10px;
}

.graph-info-panel h4 {
  margin: 0 0 10px;
}

.graph-info-panel p {
  margin: 6px 0;
  line-height: 1.35;
}

.graph-edit-btn {
  margin-top: 10px;
}

.steps-table-grid {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.steps-table-grid th,
.steps-table-grid td {
  border-bottom: 1px solid #eef0f3;
  border-right: 1px solid #eef0f3;
  padding: 8px 10px;
  vertical-align: middle;
  font-size: 14px;
}

.steps-table-grid th:last-child,
.steps-table-grid td:last-child {
  border-right: none;
}

.steps-table-grid th {
  background: #f9fafb;
  font-weight: 600;
  text-align: center;
  white-space: normal;
  line-height: 1.2;
  overflow-wrap: anywhere;
}

.steps-table-grid tbody tr:last-child td {
  border-bottom: none;
}

.steps-cell-center {
  text-align: center;
}

.steps-empty {
  border-top: 1px solid #eef0f3;
}

.steps-table-grid th:nth-child(1) { width: 34px; }
.steps-table-grid th:nth-child(2) { width: 200px; }
.steps-table-grid th:nth-child(3) { width: 130px; }
.steps-table-grid th:nth-child(4) { width: 220px; }
.steps-table-grid th:nth-child(5) { width: 110px; }
.steps-table-grid th:nth-child(6) { width: 120px; }
.steps-table-grid th:nth-child(7) { width: 120px; }
.steps-table-grid th:nth-child(8) { width: 110px; }
.steps-table-grid th:nth-child(9) { width: 86px; }

.steps-table-grid td:nth-child(4) {
  white-space: normal;
  overflow-wrap: anywhere;
  word-break: break-word;
  line-height: 1.3;
}

.steps-table-grid td:nth-child(7),
.steps-table-grid td:nth-child(8) {
  white-space: nowrap;
}

.steps-table-grid td:nth-child(9) .icon-btn {
  padding: 4px;
}

.step-actions-wrap {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.icon-btn {
  border: none;
  background: var(--color-soft-bg);
  border-radius: 6px;
  padding: 6px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.icon-btn.edit:hover {
  background: var(--color-edit-hover);
}

.icon-btn.delete:hover {
  background: var(--color-delete-hover);
}

.add-step-btn {
  padding: 6px 14px;
  background-color: var(--color-primary);
  color: white;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-weight: 500;
}

.add-step-btn:hover {
  background-color: var(--color-primary-hover);
}

.delete-version-btn {
  padding: 6px 12px;
  background-color: var(--color-soft-bg);
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.delete-version-btn:hover {
  background-color: #e0e0e0;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  justify-content: center;
  align-items: center;
}

.modal {
  background: white;
  padding: 20px;
  border-radius: 8px;
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
  color: #fff;
  padding: 12px 18px;
  border-radius: 8px;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
  z-index: 1200;
}

@media (max-width: 1100px) {
  .modal.modal-wide.step-edit-modal {
    width: 95vw;
    max-width: 95vw;
    max-height: calc(100vh - 24px);
  }

  .step-edit-overlay {
    justify-content: center;
    padding: 12px;
  }

  .modal-columns {
    flex-direction: column;
  }

  .steps-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
  }

  .steps-header-actions {
    width: 100%;
    justify-content: space-between;
  }

  .steps-graph-layout {
    grid-template-columns: 1fr;
  }

  .graph-info-panel {
    position: static;
  }

  .steps-graph-svg {
    min-width: 420px;
  }

  .measurements-header,
  .measurements-row {
    grid-template-columns: 50px 1fr 1fr 100px 120px 70px;
  }
}
</style>




