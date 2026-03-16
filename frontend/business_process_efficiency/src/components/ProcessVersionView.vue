<template>
  <div class="version-container">
    <div class="version-header">
      <h3>Версия {{ version.Version || "?" }}</h3>
      <button class="delete-version-btn" @click="showDeleteDialog = true">Удалить версию</button>
    </div>

    <div class="version-info">
      <p><b>Создана:</b> {{ formatDateTime(version.CreatedAt) }}</p>
    </div>

    <div class="steps-header">
      <h4>Этапы процесса</h4>
      <div class="steps-header-actions">
        <div class="steps-view-tabs">
          <button
            :class="['steps-view-btn', { active: stepsViewMode === 'graph' }]"
            @click="stepsViewMode = 'graph'"
            type="button"
          >
            Граф процесса
          </button>
          <button
            :class="['steps-view-btn', { active: stepsViewMode === 'table' }]"
            @click="stepsViewMode = 'table'"
            type="button"
          >
            Справка
          </button>
        </div>
        <button class="add-step-btn" @click="openCreateStepModal">+ Добавить этап</button>
      </div>
    </div>

    <div v-if="stepsViewMode === 'table'" class="steps-table">
      <div class="steps-reference-title">Справочные данные по этапам</div>
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
            <tr
              v-for="(executorRow, execIndex) in getStepExecutorRows(step)"
              :key="`${step.ID}-${execIndex}`"
              :draggable="execIndex === 0"
              :class="{
                'drag-source-step': execIndex === 0 && isDraggedStep(step.ID),
                'drag-over-step': execIndex === 0 && isDragOverStep(step.ID)
              }"
              @dragstart="execIndex === 0 ? onStepDragStart(step.ID, $event) : null"
              @dragover="execIndex === 0 ? onStepDragOver(step.ID, $event) : null"
              @dragleave="execIndex === 0 ? onStepDragLeave(step.ID) : null"
              @drop="execIndex === 0 ? onStepDrop(step.ID, $event) : null"
              @dragend="execIndex === 0 ? onStepDragEnd() : null"
            >
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
          <div class="graph-toolbar">
            <button class="icon-btn" type="button" @click="zoomGraph(0.1)">+</button>
            <button class="icon-btn" type="button" @click="zoomGraph(-0.1)">-</button>
            <button class="icon-btn" type="button" @click="resetGraphView">Сброс</button>
          </div>
          <svg
            ref="graphSvgRef"
            class="steps-graph-svg"
            :viewBox="`0 0 ${graphSvgWidth} ${graphSvgHeight}`"
            preserveAspectRatio="xMinYMin meet"
            @mousedown="onGraphMouseDown"
            @mousemove="onGraphMouseMove"
            @mouseup="onGraphMouseUp"
            @mouseleave="onGraphMouseUp"
            @wheel.prevent="onGraphWheel"
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

            <g :transform="graphTransform">
              <polyline
                v-for="edge in graphEdges"
                :key="edge.id"
                :points="edge.points"
                fill="none"
                :stroke="edge?.type === 'condition' ? '#f59e0b' : edge?.type === 'parallel' ? '#6366f1' : '#94a3b8'"
                :stroke-width="edge?.type === 'sequence' ? 1.8 : 2.2"
                :stroke-dasharray="edge?.type === 'parallel' ? '6 4' : ''"
                marker-end="url(#stepArrow)"
              />
              <text
                v-for="edge in conditionGraphEdges"
                :key="`label-${edge.id}`"
                :x="edge.labelX"
                :y="edge.labelY"
                class="graph-edge-label"
              >
                {{ edge.label }}
              </text>

              <g
                v-for="node in graphNodes"
                :key="node.id"
                class="graph-node"
                :class="{ active: Number(selectedGraphStep?.ID) === Number(node.id) }"
                @click.stop="selectGraphStep(node.id)"
                @mousedown.stop="onNodeMouseDown(node.id, $event)"
              >
                <circle
                  v-if="isEventType(node.type)"
                  :cx="node.x + GRAPH_NODE_WIDTH / 2"
                  :cy="node.y + GRAPH_NODE_HEIGHT / 2"
                  :r="Math.max(8, Math.min(GRAPH_NODE_WIDTH, GRAPH_NODE_HEIGHT) / 2)"
                />
                <polygon
                  v-else-if="isDiamondType(node.type)"
                  :points="getConditionDiamondPoints(node)"
                />
                <rect
                  v-else
                  :x="node.x"
                  :y="node.y"
                  :width="GRAPH_NODE_WIDTH"
                  :height="GRAPH_NODE_HEIGHT"
                  rx="6"
                  ry="6"
                />
                <text :x="node.x + 6" :y="node.y + 14" class="graph-node-order">{{ node.order }}.</text>
                <text :x="node.x + 26" :y="node.y + 14" class="graph-node-title">{{ shortStepName(node.name) }}</text>
              </g>

              <g
                v-for="node in addableGraphNodes"
                :key="`add-${node.id}`"
                class="graph-node-add"
                @click.stop="openCreateStepModal(node.id)"
              >
                <circle
                  :cx="node.x + GRAPH_NODE_WIDTH / 2"
                  :cy="node.y + GRAPH_NODE_HEIGHT + 18"
                  r="8"
                />
                <text
                  :x="node.x + GRAPH_NODE_WIDTH / 2"
                  :y="node.y + GRAPH_NODE_HEIGHT + 21"
                >
                  +
                </text>
              </g>
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
              <div class="step-name-wrap">
                <input
                  v-model="form.name"
                  type="text"
                  @input="onStepNameInput"
                  @focus="onStepNameFocus"
                  @blur="stepNameSuggestOpen = false"
                />
                <div v-if="stepNameSuggestOpen" class="step-name-suggest">
                  <div v-if="stepNameSuggestLoading" class="step-name-suggest-item muted">Поиск похожих этапов...</div>
                  <button
                    v-for="item in stepNameSuggestions"
                    :key="`suggest-${item.stepId}`"
                    type="button"
                    class="step-name-suggest-item"
                    @mousedown.prevent
                    @click="applyStepNameSuggestion(item)"
                  >
                    <span class="suggest-name">{{ item.stepName }}</span>
                    <span class="suggest-meta">{{ getStepTypeLabel(item.stepType) }}</span>
                  </button>
                  <div v-if="!stepNameSuggestLoading && stepNameSuggestions.length === 0" class="step-name-suggest-item muted">
                    Похожих этапов не найдено
                  </div>
                </div>
              </div>

              <label>Тип *</label>
              <select v-model="form.type" @change="onStepTypeChange">
                <option disabled value="">Выберите тип</option>
                <option v-for="(label, code) in stepTypes" :key="code" :value="code">{{ label }}</option>
              </select>

              <label>Предыдущий этап</label>
              <select v-model.number="form.previousStepId">
                <option :value="0">Нет (начало)</option>
                <option
                  v-for="stepOption in availablePreviousStepForOrder"
                  :key="`order-prev-${stepOption.ID}`"
                  :value="stepOption.ID"
                >
                  {{ getStepDisplayOrder(stepOption, sortedSteps.findIndex((s) => Number(s.ID) === Number(stepOption.ID))) }}.
                  {{ getStepTypeLabel(stepOption.Type) }}: {{ stepOption.Name }}
                </option>
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

              <template v-if="form.type === 'PARALLEL_GATEWAY'">
                <label>Ветви параллели</label>
                <div class="condition-branches">
                  <div class="condition-branch-header">
                    <span>Следующий этап</span>
                    <span></span>
                    <span></span>
                  </div>
                  <div class="condition-branch-row" v-for="(branch, idx) in form.parallelBranches" :key="`pbranch-${idx}`">
                    <select v-model.number="branch.nextStepId">
                      <option :value="0">Выберите этап</option>
                      <option
                        v-for="stepOption in availableParallelNextSteps"
                        :key="`pnext-${stepOption.ID}`"
                        :value="stepOption.ID"
                      >
                        {{ getStepTypeLabel(stepOption.Type) }}: {{ stepOption.Name }}
                      </option>
                    </select>
                    <span></span>
                    <button class="icon-btn delete" type="button" @click="removeParallelBranch(idx)">
                      <Trash :size="14" />
                    </button>
                  </div>
                  <div class="condition-branch-actions">
                    <button class="add-step-btn" type="button" @click="addParallelBranch">+ Добавить ветвь</button>
                  </div>
                </div>
              </template>

              <template v-if="form.type === 'CONDITION'">
                <label>Ветви условия</label>
                <div class="condition-branches">
                  <div class="condition-branch-header">
                    <span>Следующий этап</span>
                    <span>Вероятность (%)</span>
                    <span></span>
                  </div>
                  <div class="condition-branch-row" v-for="(branch, idx) in form.conditionBranches" :key="`branch-${idx}`">
                    <select v-model.number="branch.nextStepId">
                      <option :value="0">Выберите этап</option>
                      <option
                        v-for="stepOption in availableConditionNextSteps"
                        :key="`cond-${stepOption.ID}`"
                        :value="stepOption.ID"
                      >
                        {{ getStepTypeLabel(stepOption.Type) }}: {{ stepOption.Name }}
                      </option>
                    </select>
                    <input type="number" min="0" max="100" step="0.01" v-model.number="branch.probabilityPercent" />
                    <button class="icon-btn delete" type="button" @click="removeConditionBranch(idx)">
                      <Trash :size="14" />
                    </button>
                  </div>
                  <div class="condition-branch-actions">
                    <button class="add-step-btn" type="button" @click="addConditionBranch">+ Добавить ветвь</button>
                    <span class="condition-sum" :class="{ valid: conditionProbabilityIsValid, invalid: !conditionProbabilityIsValid }">
                      Сумма: {{ conditionProbabilitySumDisplay }}%
                    </span>
                  </div>
                </div>
              </template>

              <template v-if="form.type === 'PARALLEL_END' || form.type === 'CONDITION_END'">
                <label>Закрываемый этап</label>
                <select v-model.number="form.closesStepId">
                  <option :value="0">Выберите этап</option>
                  <option
                    v-for="stepOption in availableClosableGatewaySteps"
                    :key="`close-${stepOption.ID}`"
                    :value="stepOption.ID"
                  >
                    {{ getStepTypeLabel(stepOption.Type) }}: {{ stepOption.Name }}
                  </option>
                </select>

                <label>Предыдущие этапы</label>
                <div class="parallel-picker">
                  <label
                    v-for="stepOption in availablePreviousStepsForClosing"
                    :key="`prev-${stepOption.ID}`"
                    class="parallel-option"
                  >
                    <input
                      type="checkbox"
                      :value="stepOption.ID"
                      v-model="form.previousStepIds"
                    />
                    <span>{{ getStepTypeLabel(stepOption.Type) }}: {{ stepOption.Name }}</span>
                  </label>
                  <div v-if="availablePreviousStepsForClosing.length === 0" class="empty">Нет доступных этапов</div>
                </div>
              </template>
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
import { ref, computed, onMounted, watch } from "vue"
import { Pencil, Trash, X } from "lucide-vue-next"
import api from "../api/axios"

const props = defineProps({ version: Object, processId: Number })
const emit = defineEmits(["refresh"])

const showDeleteDialog = ref(false)
const showDeleteMeasurementDialog = ref(false)
const showDeleteStepDialog = ref(false)
const modalOpen = ref(false)
const showExecutorPicker = ref(false)
const stepsViewMode = ref("graph")
const editMode = ref(false)
const editingStepId = ref(null)
const editedMeasurements = ref([])
const deletingMeasurementId = ref(null)
const deletingStepId = ref(null)
const selectedGraphStepId = ref(null)
const createAfterStepId = ref(null)
const draggedStepId = ref(null)
const dragOverStepId = ref(null)
const isReorderingSteps = ref(false)
const stepNameSuggestions = ref([])
const stepNameSuggestOpen = ref(false)
const stepNameSuggestLoading = ref(false)
let stepNameSuggestTimer = null
const graphSvgRef = ref(null)
const graphScale = ref(1)
const graphPan = ref({ x: 0, y: 0 })
const nodeOverrides = ref({})
const draggingNodeState = ref(null)
const panningState = ref(null)

const form = ref({
  name: "",
  type: "",
  description: "",
  previousStepId: 0,
  closesStepId: 0,
  previousStepIds: [],
  executorIds: [],
  executorPercents: {},
  parallelBranches: [],
  conditionBranches: [],
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
  CONDITION: "Условие",
  PARALLEL_GATEWAY: "Параллельный шлюз",
  PARALLEL_END: "Конец параллели",
  CONDITION_END: "Конец условия"
}
const TIME_TRACKABLE_STEP_TYPES = ["SUBPROCESS", "OPERATION"]
const isTimeTrackableType = (type) => TIME_TRACKABLE_STEP_TYPES.includes(type)
const isExecutorAllowedType = (type) => TIME_TRACKABLE_STEP_TYPES.includes(type)
const GRAPH_NODE_WIDTH = 74
const GRAPH_NODE_HEIGHT = 20
const GRAPH_COLS = 1
const GRAPH_GAP_X = 44
const GRAPH_GAP_Y = 52
const GRAPH_PADDING_X = 24
const GRAPH_PADDING_Y = 20

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

const clearStepNameSuggestTimer = () => {
  if (stepNameSuggestTimer) {
    clearTimeout(stepNameSuggestTimer)
    stepNameSuggestTimer = null
  }
}

const fetchStepNameSuggestions = async () => {
  const q = String(form.value.name || "").trim()
  if (q.length < 3) {
    stepNameSuggestions.value = []
    stepNameSuggestOpen.value = false
    return
  }

  const token = localStorage.getItem("jwt")
  if (!token) return

  stepNameSuggestLoading.value = true
  try {
    const res = await api.get("/processes/steps/suggest", {
      params: { q, limit: 5, excludeProcessId: props.processId },
      headers: { Authorization: `Bearer ${token}` }
    })
    stepNameSuggestions.value = Array.isArray(res?.data) ? res.data : []
    stepNameSuggestOpen.value = true
  } catch (err) {
    console.error(err)
    stepNameSuggestions.value = []
    stepNameSuggestOpen.value = false
  } finally {
    stepNameSuggestLoading.value = false
  }
}

const onStepNameInput = () => {
  clearStepNameSuggestTimer()
  stepNameSuggestTimer = setTimeout(() => {
    fetchStepNameSuggestions()
  }, 300)
}

const onStepNameFocus = () => {
  if (String(form.value.name || "").trim().length >= 3) {
    fetchStepNameSuggestions()
  }
}

const applyStepNameSuggestion = (item) => {
  form.value.name = item?.stepName || form.value.name
  stepNameSuggestOpen.value = false
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
    previousStepId: 0,
    closesStepId: 0,
    previousStepIds: [],
    executorIds: [],
    executorPercents: {},
    parallelBranches: [],
    conditionBranches: [],
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
  createAfterStepId.value = null
  stepNameSuggestions.value = []
  stepNameSuggestOpen.value = false
  stepNameSuggestLoading.value = false
  clearStepNameSuggestTimer()
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

const graphStepById = computed(() => {
  const map = new Map()
  for (const step of sortedSteps.value) {
    map.set(Number(step.ID), step)
  }
  return map
})

const graphOrderIndexByStepId = computed(() => {
  const map = new Map()
  sortedSteps.value.forEach((step, idx) => map.set(Number(step.ID), idx))
  return map
})

const graphLinks = computed(() => {
  const edges = []
  const hasEdge = new Set()
  const byId = graphStepById.value
  const orderIndex = graphOrderIndexByStepId.value

  const addEdge = (fromId, toId, type, label = "") => {
    const from = Number(fromId)
    const to = Number(toId)
    if (!Number.isFinite(from) || !Number.isFinite(to) || from === to) return
    if (!byId.has(from) || !byId.has(to)) return
    const key = `${type}:${from}->${to}`
    if (hasEdge.has(key)) return
    hasEdge.add(key)
    edges.push({ id: key, fromId: from, toId: to, type, label })
  }

  const incomingCount = new Map()
  const bumpIncoming = (stepId) => {
    const id = Number(stepId)
    if (!Number.isFinite(id)) return
    incomingCount.set(id, Number(incomingCount.get(id) || 0) + 1)
  }

  // 1) Explicit previous links have top priority.
  for (const step of sortedSteps.value) {
    const stepId = Number(step.ID)
    const previous = Array.isArray(step.PreviousSteps) ? step.PreviousSteps : []
    for (const prev of previous) {
      const prevId = Number(prev.PreviousStepID)
      addEdge(prevId, stepId, "sequence")
      bumpIncoming(stepId)
    }
  }

  // 2) Typed branch links (condition / parallel).
  for (const step of sortedSteps.value) {
    if (step.Type === "CONDITION") {
      const branches = Array.isArray(step.ConditionBranches) ? step.ConditionBranches : []
      for (const branch of branches) {
        const nextId = Number(branch.NextStepID)
        const probability = Number(branch.ProbabilityPercent)
        const label = Number.isFinite(probability) ? `${probability.toFixed(0)}%` : ""
        addEdge(step.ID, nextId, "condition", label)
        bumpIncoming(nextId)
      }
    }

    const parallels = Array.isArray(step.ParallelBranches) ? step.ParallelBranches : []
    const fromIndex = orderIndex.get(Number(step.ID))
    for (const parallel of parallels) {
      const toId = Number(parallel.NextStepID)
      const toIndex = orderIndex.get(toId)
      if (!Number.isFinite(fromIndex) || !Number.isFinite(toIndex)) continue
      if (toIndex <= fromIndex) continue
      addEdge(step.ID, toId, "parallel")
      bumpIncoming(toId)
    }

    if (step.Type === "PARALLEL_END" || step.Type === "CONDITION_END") {
      const previous = Array.isArray(step.PreviousSteps) ? step.PreviousSteps : []
      for (const prev of previous) {
        const prevId = Number(prev.PreviousStepID)
        addEdge(prevId, step.ID, "sequence")
        bumpIncoming(step.ID)
      }
    }
  }

  // 3) Fallback linear links only for steps without explicit incoming links.
  for (let i = 0; i < sortedSteps.value.length - 1; i++) {
    const from = sortedSteps.value[i]
    const to = sortedSteps.value[i + 1]
    const toId = Number(to.ID)
    if (Number(incomingCount.get(toId) || 0) > 0) continue

    const branches = Array.isArray(from.ConditionBranches) ? from.ConditionBranches : []
    const parallelBranches = Array.isArray(from.ParallelBranches) ? from.ParallelBranches : []
    if (from.Type === "CONDITION" && branches.length > 0) continue
    if (from.Type === "PARALLEL_GATEWAY" && parallelBranches.length > 0) continue

    addEdge(from.ID, to.ID, "sequence")
    bumpIncoming(toId)
  }

  return edges
})

const graphColumnByStepId = computed(() => {
  const columns = new Map()
  const orderIndex = graphOrderIndexByStepId.value

  for (const step of sortedSteps.value) {
    const stepId = Number(step.ID)
    if (!Number.isFinite(stepId)) continue
    if (!columns.has(stepId)) columns.set(stepId, 0)

    const outgoing = graphLinks.value.filter((e) => e.fromId === stepId)
    if (outgoing.length <= 0) continue

    const baseCol = Number(columns.get(stepId) || 0)
    if (outgoing.length === 1) {
      const toId = outgoing[0].toId
      if (!columns.has(toId)) columns.set(toId, baseCol)
      continue
    }

    const sortedOutgoing = [...outgoing].sort((a, b) => {
      const ai = Number(orderIndex.get(a.toId) ?? 0)
      const bi = Number(orderIndex.get(b.toId) ?? 0)
      return ai - bi
    })

    const startShift = -((sortedOutgoing.length - 1) / 2)
    sortedOutgoing.forEach((edge, idx) => {
      const toId = edge.toId
      if (columns.has(toId)) return
      columns.set(toId, baseCol + startShift + idx)
    })
  }

  return columns
})

const graphAutoNodes = computed(() => {
  const columns = graphColumnByStepId.value
  const rawCols = sortedSteps.value.map((step) => Number(columns.get(Number(step.ID)) || 0))
  const minCol = rawCols.length > 0 ? Math.min(...rawCols) : 0

  return sortedSteps.value.map((step, index) => {
    const row = Math.floor(index / GRAPH_COLS)
    const colRaw = Number(columns.get(Number(step.ID)) || 0)
    const col = colRaw - minCol
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

watch(
  graphAutoNodes,
  (nodes) => {
    const next = {}
    for (const node of nodes) {
      const id = Number(node.id)
      const existing = nodeOverrides.value[id]
      next[id] = existing && Number.isFinite(existing.x) && Number.isFinite(existing.y)
        ? existing
        : { x: node.x, y: node.y }
    }
    nodeOverrides.value = next
  },
  { immediate: true }
)

const graphNodes = computed(() =>
  graphAutoNodes.value.map((node) => {
    const override = nodeOverrides.value[Number(node.id)]
    if (!override) return node
    return {
      ...node,
      x: Number.isFinite(override.x) ? override.x : node.x,
      y: Number.isFinite(override.y) ? override.y : node.y
    }
  })
)

const addableGraphNodes = computed(() =>
  graphNodes.value.filter((node) => !isTerminalType(node?.type))
)

const graphEdges = computed(() => {
  const nodeById = new Map(graphNodes.value.map((n) => [Number(n.id), n]))
  return graphLinks.value
    .map((link) => {
      const fromNode = nodeById.get(Number(link.fromId))
      const toNode = nodeById.get(Number(link.toId))
      if (!fromNode || !toNode) return null

      const fromX = fromNode.x + GRAPH_NODE_WIDTH / 2
      const fromY = fromNode.y + GRAPH_NODE_HEIGHT
      const toX = toNode.x + GRAPH_NODE_WIDTH / 2
      const toY = toNode.y
      const midY = Math.round((fromY + toY) / 2)
      return {
        id: link.id,
        type: link.type,
        label: link.label || "",
        points: `${fromX},${fromY} ${fromX},${midY} ${toX},${midY} ${toX},${toY}`,
        labelX: toX + 6,
        labelY: midY - 4
      }
    })
    .filter(Boolean)
})

const conditionGraphEdges = computed(() =>
  graphEdges.value.filter((edge) => edge && edge.type === "condition" && edge.label)
)

const graphSvgWidth = computed(() => {
  const nodes = graphNodes.value
  if (nodes.length === 0) return GRAPH_PADDING_X * 2 + GRAPH_NODE_WIDTH
  const minX = Math.min(...nodes.map((n) => n.x))
  const maxX = Math.max(...nodes.map((n) => n.x))
  return maxX - minX + GRAPH_NODE_WIDTH + GRAPH_PADDING_X * 2
})

const graphSvgHeight = computed(() => {
  const rows = Math.max(1, Math.ceil(sortedSteps.value.length / GRAPH_COLS))
  return GRAPH_PADDING_Y * 2 + rows * GRAPH_NODE_HEIGHT + Math.max(0, rows - 1) * GRAPH_GAP_Y
})

const graphTransform = computed(() => `translate(${graphPan.value.x} ${graphPan.value.y}) scale(${graphScale.value})`)

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
  if (str.length <= 8) return str
  return `${str.slice(0, 7)}…`
}

const isEventType = (type) => type === "START" || type === "END"
const isTerminalType = (type) => type === "END"
const isDiamondType = (type) =>
  type === "CONDITION" || type === "PARALLEL_GATEWAY" || type === "CONDITION_END" || type === "PARALLEL_END"

const getConditionDiamondPoints = (node) => {
  const cx = node.x + GRAPH_NODE_WIDTH / 2
  const cy = node.y + GRAPH_NODE_HEIGHT / 2
  const rx = GRAPH_NODE_WIDTH / 2
  const ry = GRAPH_NODE_HEIGHT / 2
  return `${cx},${cy - ry} ${cx + rx},${cy} ${cx},${cy + ry} ${cx - rx},${cy}`
}

const clampGraphScale = (value) => Math.min(2.4, Math.max(0.5, value))

const zoomGraph = (delta) => {
  graphScale.value = clampGraphScale(graphScale.value + delta)
}

const resetGraphView = () => {
  graphScale.value = 1
  graphPan.value = { x: 0, y: 0 }
}

const toGraphCoords = (event) => {
  const svg = graphSvgRef.value
  if (!svg || !svg.getScreenCTM) return { x: 0, y: 0 }

  const point = svg.createSVGPoint()
  point.x = event.clientX
  point.y = event.clientY
  const ctm = svg.getScreenCTM()
  if (!ctm) return { x: 0, y: 0 }

  const svgPoint = point.matrixTransform(ctm.inverse())
  return {
    x: (svgPoint.x - graphPan.value.x) / graphScale.value,
    y: (svgPoint.y - graphPan.value.y) / graphScale.value
  }
}

const onNodeMouseDown = (nodeId, event) => {
  if (event.button !== 0) return
  const node = graphNodes.value.find((n) => Number(n.id) === Number(nodeId))
  if (!node) return
  const point = toGraphCoords(event)
  draggingNodeState.value = {
    id: Number(nodeId),
    offsetX: point.x - node.x,
    offsetY: point.y - node.y
  }
}

const onGraphMouseDown = (event) => {
  if (event.button !== 0) return
  if (draggingNodeState.value) return
  panningState.value = {
    startX: event.clientX,
    startY: event.clientY,
    panX: graphPan.value.x,
    panY: graphPan.value.y
  }
}

const onGraphMouseMove = (event) => {
  if (draggingNodeState.value) {
    const point = toGraphCoords(event)
    const nodeId = Number(draggingNodeState.value.id)
    const nextX = point.x - draggingNodeState.value.offsetX
    const nextY = point.y - draggingNodeState.value.offsetY
    nodeOverrides.value = {
      ...nodeOverrides.value,
      [nodeId]: { x: nextX, y: nextY }
    }
    return
  }

  if (panningState.value) {
    const dx = event.clientX - panningState.value.startX
    const dy = event.clientY - panningState.value.startY
    graphPan.value = {
      x: panningState.value.panX + dx,
      y: panningState.value.panY + dy
    }
  }
}

const onGraphMouseUp = () => {
  draggingNodeState.value = null
  panningState.value = null
}

const onGraphWheel = (event) => {
  const delta = event.deltaY < 0 ? 0.1 : -0.1
  zoomGraph(delta)
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

const openCreateStepModal = (afterStepId = null) => {
  resetForm()
  createAfterStepId.value = Number.isFinite(Number(afterStepId)) ? Number(afterStepId) : null
  form.value.previousStepId = createAfterStepId.value || 0
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
    previousStepId: 0,
    closesStepId: Number(step.ClosesStepID) || 0,
    previousStepIds: (step.PreviousSteps || []).map((p) => Number(p.PreviousStepID)).filter(Number.isFinite),
    executorIds: isExecutorAllowedType(step.Type)
      ? (step.StepExecutors || []).map((se) => Number(se.EmployeeID)).filter(Number.isFinite)
      : [],
    executorPercents: isExecutorAllowedType(step.Type)
      ? Object.fromEntries((step.StepExecutors || []).map((se) => [Number(se.EmployeeID), Number(se.WorkloadPercent || 0)]))
      : {},
    parallelBranches: step.Type === "PARALLEL_GATEWAY"
      ? (step.ParallelBranches || []).map((b) => ({
          nextStepId: Number(b.NextStepID) || 0
        }))
      : [],
    conditionBranches: step.Type === "CONDITION"
      ? (step.ConditionBranches || []).map((b) => ({
          nextStepId: Number(b.NextStepID) || 0,
          probabilityPercent: Number(b.ProbabilityPercent) || 0
        }))
      : [],
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

  const explicitPrevious = (step.PreviousSteps || []).map((p) => Number(p.PreviousStepID)).find(Number.isFinite)
  if (Number.isFinite(explicitPrevious) && explicitPrevious > 0) {
    form.value.previousStepId = explicitPrevious
  } else {
    const currentIndex = sortedSteps.value.findIndex((s) => Number(s.ID) === Number(step.ID))
    form.value.previousStepId = currentIndex > 0 ? Number(sortedSteps.value[currentIndex - 1].ID) || 0 : 0
  }

  executorSearch.value = ""
  showExecutorPicker.value = false
  modalOpen.value = true
  syncEditedStepFromBackend()
}

const getStepTypeLabel = (type) => stepTypes[type] || type || "—"

function getStepDisplayOrder(step, fallbackIndex) {
  const order = Number(step?.StepOrder)
  if (Number.isFinite(order) && order > 0) return order
  return fallbackIndex + 1
}

const isDraggedStep = (stepId) => Number(draggedStepId.value) === Number(stepId)
const isDragOverStep = (stepId) => Number(dragOverStepId.value) === Number(stepId)

const onStepDragStart = (stepId, event) => {
  if (isReorderingSteps.value) {
    event.preventDefault()
    return
  }
  draggedStepId.value = Number(stepId)
  dragOverStepId.value = null
  if (event?.dataTransfer) {
    event.dataTransfer.effectAllowed = "move"
    event.dataTransfer.setData("text/plain", String(stepId))
  }
}

const onStepDragOver = (stepId, event) => {
  if (isReorderingSteps.value) return
  if (!draggedStepId.value || Number(draggedStepId.value) === Number(stepId)) return
  event.preventDefault()
  if (event?.dataTransfer) {
    event.dataTransfer.dropEffect = "move"
  }
  dragOverStepId.value = Number(stepId)
}

const onStepDragLeave = (stepId) => {
  if (Number(dragOverStepId.value) === Number(stepId)) {
    dragOverStepId.value = null
  }
}

const onStepDragEnd = () => {
  draggedStepId.value = null
  dragOverStepId.value = null
}

const onStepDrop = async (targetStepId, event) => {
  event.preventDefault()

  const sourceStepId = Number(draggedStepId.value)
  const toStepId = Number(targetStepId)
  onStepDragEnd()

  if (!Number.isFinite(sourceStepId) || !Number.isFinite(toStepId) || sourceStepId === toStepId) {
    return
  }

  const orderedIds = sortedSteps.value.map((s) => Number(s.ID)).filter(Number.isFinite)
  const fromIndex = orderedIds.findIndex((id) => id === sourceStepId)
  const toIndex = orderedIds.findIndex((id) => id === toStepId)

  if (fromIndex < 0 || toIndex < 0) return

  const nextOrder = [...orderedIds]
  const [moved] = nextOrder.splice(fromIndex, 1)
  nextOrder.splice(toIndex, 0, moved)

  const unchanged = nextOrder.every((id, idx) => id === orderedIds[idx])
  if (unchanged) return

  const token = localStorage.getItem("jwt")
  if (!token) return

  try {
    isReorderingSteps.value = true
    await api.post(
      "/processes/steps/reorder",
      {
        processVersionId: props.version.ID,
        orderedStepIds: nextOrder
      },
      { headers: { Authorization: `Bearer ${token}` } }
    )
    showToast("Порядок этапов обновлен")
    emit("refresh")
  } catch (err) {
    console.error(err)
    showToast("Ошибка изменения порядка этапов")
  } finally {
    isReorderingSteps.value = false
  }
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
  if (form.value.type !== "PARALLEL_GATEWAY") {
    form.value.parallelBranches = []
  }
  if (form.value.type !== "CONDITION") {
    form.value.conditionBranches = []
  }
  if (form.value.type !== "PARALLEL_END" && form.value.type !== "CONDITION_END") {
    form.value.closesStepId = 0
    form.value.previousStepIds = []
  }
}

const addParallelBranch = () => {
  if (!Array.isArray(form.value.parallelBranches)) form.value.parallelBranches = []
  form.value.parallelBranches.push({ nextStepId: 0 })
}

const removeParallelBranch = (idx) => {
  form.value.parallelBranches.splice(idx, 1)
}

const addConditionBranch = () => {
  if (!Array.isArray(form.value.conditionBranches)) form.value.conditionBranches = []
  form.value.conditionBranches.push({ nextStepId: 0, probabilityPercent: 0 })
}

const removeConditionBranch = (idx) => {
  form.value.conditionBranches.splice(idx, 1)
}

const availableParallelNextSteps = computed(() => {
  return sortedSteps.value.filter((s) =>
    Number(s.ID) !== Number(editingStepId.value || 0)
  )
})

const availableConditionNextSteps = computed(() => {
  return sortedSteps.value.filter((s) => Number(s.ID) !== Number(editingStepId.value || 0))
})

const availablePreviousStepForOrder = computed(() =>
  sortedSteps.value.filter(
    (s) => Number(s.ID) !== Number(editingStepId.value || 0) && !isTerminalType(s.Type)
  )
)

const availableClosableGatewaySteps = computed(() => {
  if (form.value.type === "PARALLEL_END") {
    return sortedSteps.value.filter((s) => s.Type === "PARALLEL_GATEWAY")
  }
  if (form.value.type === "CONDITION_END") {
    return sortedSteps.value.filter((s) => s.Type === "CONDITION")
  }
  return []
})

const availablePreviousStepsForClosing = computed(() =>
  sortedSteps.value.filter((s) => Number(s.ID) !== Number(editingStepId.value || 0))
)

const conditionProbabilitySum = computed(() => {
  return (form.value.conditionBranches || [])
    .map((b) => Number(b.probabilityPercent))
    .reduce((acc, v) => acc + (Number.isFinite(v) ? v : 0), 0)
})

const conditionProbabilitySumDisplay = computed(() => conditionProbabilitySum.value.toFixed(2))
const conditionProbabilityIsValid = computed(() => Math.abs(conditionProbabilitySum.value - 100) <= 0.0001)

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

const fetchCurrentVersionStepIds = async (token) => {
  const res = await api.get(`/processes/${props.processId}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  const versions = Array.isArray(res?.data?.Versions) ? res.data.Versions : []
  const version = versions.find((v) => Number(v.ID) === Number(props.version.ID))
  const steps = Array.isArray(version?.Steps) ? [...version.Steps] : []
  steps.sort((a, b) => {
    const ao = Number(a?.StepOrder)
    const bo = Number(b?.StepOrder)
    if (Number.isFinite(ao) && Number.isFinite(bo) && ao !== bo) return ao - bo
    return Number(a?.ID || 0) - Number(b?.ID || 0)
  })
  return steps.map((s) => Number(s.ID)).filter(Number.isFinite)
}

const applyStepOrderByPrevious = async (stepId, previousStepId, token) => {
  const currentIds = await fetchCurrentVersionStepIds(token)
  if (!currentIds.includes(Number(stepId))) return

  const nextOrder = currentIds.filter((id) => id !== Number(stepId))
  const prevId = Number(previousStepId) || 0
  if (prevId > 0 && nextOrder.includes(prevId)) {
    const idx = nextOrder.findIndex((id) => id === prevId)
    nextOrder.splice(idx + 1, 0, Number(stepId))
  } else {
    nextOrder.unshift(Number(stepId))
  }

  await api.post(
    "/processes/steps/reorder",
    {
      processVersionId: props.version.ID,
      orderedStepIds: nextOrder
    },
    { headers: { Authorization: `Bearer ${token}` } }
  )
}

const saveStep = async () => {
  if (!form.value.name || !form.value.type) {
    alert("Заполните обязательные поля")
    return
  }

  const canUseTimeSettings = isTimeTrackableType(form.value.type)
  const canUseExecutors = isExecutorAllowedType(form.value.type)
  const isCondition = form.value.type === "CONDITION"
  const isParallelGateway = form.value.type === "PARALLEL_GATEWAY"
  const isClosingStep = form.value.type === "PARALLEL_END" || form.value.type === "CONDITION_END"
  if (!canUseTimeSettings) {
    form.value.useStatistics = false
  }
  if (!canUseExecutors) {
    form.value.executorIds = []
    form.value.executorPercents = {}
  }
  if (!isParallelGateway) {
    form.value.parallelBranches = []
  }
  if (!isCondition) {
    form.value.conditionBranches = []
  }
  if (!isClosingStep) {
    form.value.closesStepId = 0
    form.value.previousStepIds = []
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
  const parallelBranches = isParallelGateway
    ? (form.value.parallelBranches || []).map((b) => ({ nextStepId: Number(b.nextStepId) || 0 }))
    : []
  const conditionBranches = isCondition
    ? (form.value.conditionBranches || []).map((b) => ({
        nextStepId: Number(b.nextStepId) || 0,
        probabilityPercent: Number(b.probabilityPercent) || 0
      }))
    : []
  const closesStepId = isClosingStep ? Number(form.value.closesStepId) || 0 : 0
  const previousStepIds = isClosingStep
    ? (form.value.previousStepIds || []).map(Number).filter((id) => Number.isFinite(id) && id > 0)
    : []
  const regularPreviousStepIds = !isClosingStep && Number(form.value.previousStepId) > 0
    ? [Number(form.value.previousStepId)]
    : []

  if (isParallelGateway) {
    if (parallelBranches.length > 0) {
      const hasEmptyNext = parallelBranches.some((b) => !Number.isFinite(b.nextStepId) || b.nextStepId <= 0)
      if (hasEmptyNext) {
        showToast("Выберите следующий этап для каждой ветви параллели")
        return
      }
      const uniqueNext = new Set(parallelBranches.map((b) => b.nextStepId))
      if (uniqueNext.size !== parallelBranches.length) {
        showToast("Этапы в ветвях параллели не должны повторяться")
        return
      }
    }
  }

  if (isCondition) {
    if (conditionBranches.length > 0) {
      const hasEmptyNext = conditionBranches.some((b) => !Number.isFinite(b.nextStepId) || b.nextStepId <= 0)
      if (hasEmptyNext) {
        showToast("Выберите следующий этап для каждой ветви условия")
        return
      }
      const uniqueNext = new Set(conditionBranches.map((b) => b.nextStepId))
      if (uniqueNext.size !== conditionBranches.length) {
        showToast("Этапы в ветвях условия не должны повторяться")
        return
      }
      const invalidProb = conditionBranches.some((b) => !Number.isFinite(b.probabilityPercent) || b.probabilityPercent < 0 || b.probabilityPercent > 100)
      if (invalidProb) {
        showToast("Вероятность ветви условия должна быть в диапазоне 0..100")
        return
      }
      if (!conditionProbabilityIsValid.value) {
        showToast("Сумма вероятностей ветвей условия должна быть ровно 100%")
        return
      }
    }
  }
  if (isClosingStep) {
    if (!Number.isFinite(closesStepId) || closesStepId <= 0) {
      showToast("Укажите, какой шлюз закрывает этот этап")
      return
    }
    if (previousStepIds.length === 0) {
      showToast("Укажите предыдущие этапы для закрывающего этапа")
      return
    }
  }
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
      const updatedStepId = Number(editingStepId.value)
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
          ClosesStepId: isClosingStep ? closesStepId : null,
          PreviousStepIds: isClosingStep ? previousStepIds : regularPreviousStepIds,
          ExecutorLoads: canUseExecutors
            ? executorLoads.map((x) => ({
                employeeId: x.employeeId,
                workloadPercent: x.workloadPercent
              }))
            : [],
          ParallelBranches: parallelBranches.map((b) => ({ nextStepId: b.nextStepId })),
          ConditionBranches: conditionBranches.map((b) => ({
            nextStepId: b.nextStepId,
            probabilityPercent: b.probabilityPercent
          })),
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
      await applyStepOrderByPrevious(updatedStepId, form.value.previousStepId, token)
      showToast("Этап обновлен")
    } else {
      const createRes = await api.post(
        "/processes/steps",
        {
          processVersionId: props.version.ID,
          name: form.value.name,
          type: form.value.type,
          description: form.value.description,
          closesStepId: isClosingStep ? closesStepId : null,
          previousStepIds: isClosingStep ? previousStepIds : regularPreviousStepIds,
          executorIds: canUseExecutors ? executorIds : [],
          parallelBranches: parallelBranches.map((b) => ({ nextStepId: b.nextStepId })),
          conditionBranches: conditionBranches.map((b) => ({
            nextStepId: b.nextStepId,
            probabilityPercent: b.probabilityPercent
          }))
        },
        { headers: { Authorization: `Bearer ${token}` } }
      )

      const createdStepId = Number(createRes?.data?.ID)
      if (Number.isFinite(createdStepId) && createdStepId > 0) {
        await applyStepOrderByPrevious(createdStepId, form.value.previousStepId, token)
      }

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

.step-name-wrap {
  position: relative;
}

.step-name-suggest {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  z-index: 50;
  background: #fff;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  box-shadow: 0 8px 16px rgba(15, 23, 42, 0.08);
  max-height: 220px;
  overflow-y: auto;
}

.step-name-suggest-item {
  width: 100%;
  border: none;
  background: transparent;
  text-align: left;
  padding: 8px 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
}

.step-name-suggest-item:hover {
  background: #f8fafc;
}

.step-name-suggest-item.muted {
  color: #6b7280;
  cursor: default;
}

.suggest-name {
  color: #111827;
}

.suggest-meta {
  color: #6b7280;
  font-size: 12px;
  white-space: nowrap;
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

.parallel-picker {
  border: 1px solid var(--color-muted-bg);
  border-radius: 8px;
  background: #fafafa;
  padding: 8px;
  max-height: 160px;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.parallel-option {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.condition-branches {
  border: 1px solid var(--color-muted-bg);
  border-radius: 8px;
  background: #fafafa;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.condition-branch-header,
.condition-branch-row {
  display: grid;
  grid-template-columns: 1fr 130px 40px;
  gap: 8px;
  align-items: center;
}

.condition-branch-header {
  font-size: 12px;
  font-weight: 600;
  color: #6b7280;
}

.condition-branch-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.condition-sum {
  font-size: 12px;
  font-weight: 600;
}

.condition-sum.valid {
  color: #166534;
}

.condition-sum.invalid {
  color: #b91c1c;
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

.steps-reference-title {
  padding: 10px 12px;
  font-size: 13px;
  font-weight: 600;
  color: #4b5563;
  border-bottom: 1px solid #eef0f3;
  background: #f9fafb;
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
  position: relative;
}

.graph-toolbar {
  position: sticky;
  top: 8px;
  left: 8px;
  z-index: 3;
  display: inline-flex;
  gap: 6px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 4px;
  margin: 8px;
}

.steps-graph-svg {
  width: 100%;
  min-width: 420px;
  height: 100%;
  min-height: 420px;
}

.graph-node {
  cursor: move;
}

.graph-node rect,
.graph-node circle,
.graph-node polygon {
  fill: #eef2ff;
  stroke: #c7d2fe;
  stroke-width: 1.5;
  transition: fill 0.15s, stroke-color 0.15s;
}

.graph-node.active rect,
.graph-node.active circle,
.graph-node.active polygon {
  fill: #e0e7ff;
  stroke: #6366f1;
}

.graph-node:hover rect,
.graph-node:hover circle,
.graph-node:hover polygon {
  fill: #e5edff;
  stroke: #818cf8;
}

.graph-node-add {
  cursor: pointer;
}

.graph-node-add circle {
  fill: #ffffff;
  stroke: #6366f1;
  stroke-width: 1.4;
}

.graph-node-add text {
  fill: #4338ca;
  font-size: 12px;
  text-anchor: middle;
  font-weight: 700;
}

.graph-node-add:hover circle {
  fill: #eef2ff;
}

.graph-node-order {
  fill: #312e81;
  font-size: 9px;
  font-weight: 600;
}

.graph-node-title {
  fill: #111827;
  font-size: 9px;
  font-weight: 600;
}

.graph-node-type {
  fill: #475569;
  font-size: 12px;
}

.graph-edge-label {
  fill: #b45309;
  font-size: 9px;
  font-weight: 600;
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

.steps-table-grid tbody tr[draggable="true"] {
  cursor: grab;
}

.steps-table-grid tbody tr[draggable="true"]:active {
  cursor: grabbing;
}

.drag-source-step td {
  opacity: 0.65;
}

.drag-over-step td {
  background: #eef2ff;
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




