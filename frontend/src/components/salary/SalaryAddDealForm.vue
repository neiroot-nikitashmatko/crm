<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { NButton, NDatePicker, NInput, NSelect } from 'naive-ui'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import DateTimeField from '@/components/common/DateTimeField.vue'
import { SALARY_SERVICE_OPTIONS } from '@/constants/salary'
import { useAuth } from '@/composables/useAuth'
import { useDeals } from '@/composables/useDeals'
import { useEmployees } from '@/composables/useEmployees'
import { useSalaryReport } from '@/composables/useSalaryReport'
import {
  renderSalaryDealOption,
  salaryDealOptionFullLabel,
  salaryDealOptionNumberLabel,
} from '@/utils/salaryDealLabel'
import { formatSalaryEmployeeName, isSalaryMasterEmployee } from '@/utils/salaryEmployee'

const { isAdmin } = useAuth()
const { deals, loadDeals } = useDeals()
const { employees, loadEmployees } = useEmployees()
const { addReportEntry } = useSalaryReport()
const isDealsLoading = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref('')
const dealSearchQuery = ref('')
const isDateModalOpen = ref(false)
const dateDraft = ref<number | null>(null)

const WORKING_HOURS = [9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]
const TIME_MINUTE_STEP = 5
const TIME_PICKER_PROPS = {
  format: 'HH:mm',
  hours: WORKING_HOURS,
  minutes: TIME_MINUTE_STEP,
  actions: ['confirm'] as Array<'confirm'>,
}

const fieldInputTheme = {
  border: '1px solid #cbd5e1',
  borderHover: '1px solid #cbd5e1',
  borderFocus: '1px solid #93c5fd',
  boxShadowFocus: '0 0 0 3px rgba(147, 197, 253, 0.25)',
  borderRadius: '8px',
  heightMedium: '36px',
  fontSizeMedium: '14px',
}

const fieldSelectTheme = {
  peers: {
    InternalSelection: {
      border: '1px solid #cbd5e1',
      borderHover: '1px solid #cbd5e1',
      borderFocus: '1px solid #93c5fd',
      borderActive: '1px solid #93c5fd',
      boxShadowFocus: '0 0 0 3px rgba(147, 197, 253, 0.25)',
      boxShadowActive: '0 0 0 3px rgba(147, 197, 253, 0.25)',
      boxShadowHover: 'none',
      borderRadius: '8px',
      heightMedium: '36px',
      fontSizeMedium: '14px',
    },
  },
}

const form = reactive({
  employeeId: null as string | null,
  dealId: null as string | null,
  date: null as number | null,
  service: null as string | null,
  salary: '',
  comment: '',
})

const serviceOptions = SALARY_SERVICE_OPTIONS.map((item) => ({
  label: item.label,
  value: item.value,
}))

const employeeOptions = computed(() =>
  employees.value
    .filter((employee) => isSalaryMasterEmployee(employee))
    .slice()
    .sort((left, right) => formatSalaryEmployeeName(left).localeCompare(formatSalaryEmployeeName(right), 'ru'))
    .map((employee) => ({
      label: formatSalaryEmployeeName(employee),
      value: employee.id,
    })),
)

function isClosedDeal(deal: { columnId?: string; status?: string }) {
  return deal.columnId === 'closed' || String(deal.status ?? '').toLowerCase() === 'closed'
}

const closedDealOptions = computed(() => {
  const query = dealSearchQuery.value.trim().toLowerCase()

  return deals.value
    .filter((deal) => isClosedDeal(deal))
    .filter((deal) => {
      if (form.dealId && deal.id === form.dealId) return true
      if (!query) return true
      return String(deal.dealNumber).includes(query)
    })
    .slice()
    .sort((left, right) => Number(right.dealNumber) - Number(left.dealNumber))
    .map((deal) => ({
      label: salaryDealOptionNumberLabel(deal.dealNumber),
      fullLabel: salaryDealOptionFullLabel(deal),
      value: deal.id,
    }))
})

const canSubmit = computed(() => {
  const hasEmployee = !isAdmin.value || form.employeeId !== null
  return (
    hasEmployee &&
    form.dealId !== null &&
    form.date !== null &&
    form.service !== null &&
    form.salary.trim().length > 0
  )
})

onMounted(() => {
  void refreshDeals()
  if (isAdmin.value) {
    void loadEmployees(true)
  }
})

async function refreshDeals() {
  isDealsLoading.value = true
  try {
    await loadDeals(true)
  } finally {
    isDealsLoading.value = false
  }
}

function handleDealSearch(query: string) {
  dealSearchQuery.value = query
}

function keepAllFilteredOptions() {
  return true
}

function formatDate(timestamp: number | null) {
  if (timestamp === null) return 'Выберите дату'
  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(timestamp)
}

function openDateModal() {
  dateDraft.value = form.date ?? Date.now()
  isDateModalOpen.value = true
}

function handleDateConfirm(onConfirm: () => void) {
  onConfirm()
  form.date = dateDraft.value
  isDateModalOpen.value = false
}

function clearDate() {
  form.date = null
  dateDraft.value = null
}

function handleSalaryInput(value: string) {
  form.salary = value.replace(/\D/g, '')
}

function resetForm() {
  form.employeeId = null
  form.dealId = null
  form.date = null
  form.service = null
  form.salary = ''
  form.comment = ''
  dealSearchQuery.value = ''
  dateDraft.value = null
  isDateModalOpen.value = false
  errorMessage.value = ''
}

async function handleSubmit() {
  if (
    !canSubmit.value ||
    isSubmitting.value ||
    form.dealId === null ||
    form.date === null ||
    form.service === null ||
    (isAdmin.value && form.employeeId === null)
  ) {
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''
  try {
    await addReportEntry({
      date: form.date,
      dealId: form.dealId,
      service: form.service,
      salary: form.salary.trim(),
      comment: form.comment.trim(),
      ...(isAdmin.value && form.employeeId ? { employeeId: form.employeeId } : {}),
    })
    resetForm()
  } catch (error) {
    if (error instanceof Error && error.message) {
      errorMessage.value = error.message
    } else {
      errorMessage.value = 'Не удалось сохранить запись'
    }
  } finally {
    isSubmitting.value = false
  }
}

defineExpose({ resetForm })
</script>

<template>
  <form class="salary-add-deal-form" @submit.prevent="handleSubmit">
    <div class="salary-add-deal-form__columns">
      <section class="salary-add-deal-form__section">
        <div v-if="isAdmin" class="salary-add-deal-form__field">
          <span class="salary-add-deal-form__label">
            Сотрудник
            <span class="salary-add-deal-form__required" aria-hidden="true">*</span>
          </span>
          <NSelect
            v-model:value="form.employeeId"
            class="salary-add-deal-form__control"
            :theme-overrides="fieldSelectTheme"
            :options="employeeOptions"
            filterable
            clearable
            placeholder="Выберите сотрудника"
          />
        </div>

        <div class="salary-add-deal-form__field">
          <span class="salary-add-deal-form__label">
            Номер сделки
            <span class="salary-add-deal-form__required" aria-hidden="true">*</span>
          </span>
          <NSelect
            v-model:value="form.dealId"
            class="salary-add-deal-form__control"
            :theme-overrides="fieldSelectTheme"
            :options="closedDealOptions"
            :loading="isDealsLoading"
            :render-option="renderSalaryDealOption"
            filterable
            clearable
            placeholder="Начните вводить номер сделки"
            :filter="keepAllFilteredOptions"
            @search="handleDealSearch"
          />
          <p
            v-if="!isDealsLoading && closedDealOptions.length === 0 && !dealSearchQuery.trim()"
            class="salary-add-deal-form__hint"
          >
            Нет закрытых сделок для выбора
          </p>
        </div>

        <div class="salary-add-deal-form__field">
          <span class="salary-add-deal-form__label">
            Дата
            <span class="salary-add-deal-form__required" aria-hidden="true">*</span>
          </span>
          <DateTimeField
            :display-value="formatDate(form.date)"
            :has-value="form.date !== null"
            @open="openDateModal"
            @clear="clearDate"
          />
        </div>
      </section>

      <section class="salary-add-deal-form__section">
        <label class="salary-add-deal-form__field">
          <span class="salary-add-deal-form__label">
            Услуга/Товар
            <span class="salary-add-deal-form__required" aria-hidden="true">*</span>
          </span>
          <NSelect
            v-model:value="form.service"
            class="salary-add-deal-form__control"
            :theme-overrides="fieldSelectTheme"
            :options="serviceOptions"
            clearable
            placeholder="Выберите услугу"
          />
        </label>

        <label class="salary-add-deal-form__field">
          <span class="salary-add-deal-form__label">
            Зарплата сотрудника
            <span class="salary-add-deal-form__required" aria-hidden="true">*</span>
          </span>
          <NInput
            :value="form.salary"
            class="salary-add-deal-form__control"
            :theme-overrides="fieldInputTheme"
            placeholder="Введите сумму"
            inputmode="numeric"
            autocomplete="off"
            @update:value="handleSalaryInput"
          />
        </label>
      </section>
    </div>

    <label class="salary-add-deal-form__field salary-add-deal-form__field--full">
      <span class="salary-add-deal-form__label">Комментарий</span>
      <NInput
        v-model:value="form.comment"
        class="salary-add-deal-form__control"
        :theme-overrides="fieldInputTheme"
        type="textarea"
        :autosize="{ minRows: 4, maxRows: 8 }"
        placeholder="Комментарий (необязательно)"
      />
    </label>

    <div class="salary-add-deal-form__actions">
      <p v-if="errorMessage" class="salary-add-deal-form__error" role="alert">
        {{ errorMessage }}
      </p>
      <AppModalButton :disabled="!canSubmit || isSubmitting" @click="handleSubmit">
        {{ isSubmitting ? 'Сохранение…' : 'Добавить сделку' }}
      </AppModalButton>
    </div>
  </form>

  <AppModal
    v-model:show="isDateModalOpen"
    title="Дата завершения сделки"
    width="wide"
    body-variant="date"
    close-label="Закрыть выбор даты"
  >
    <NDatePicker
      v-model:value="dateDraft"
      class="app-modal__date-panel salary-add-deal-form__date-panel"
      type="datetime"
      panel
      clearable
      default-time="09:00:00"
      date-format="dd.MM.yyyy"
      time-picker-format="HH:mm"
      format="dd.MM.yyyy HH:mm"
      :time-picker-props="TIME_PICKER_PROPS"
      :actions="['confirm']"
    >
      <template #confirm="{ text, onConfirm }">
        <div class="app-modal__date-confirm">
          <NButton
            class="app-modal__date-confirm-btn"
            type="primary"
            size="small"
            @click="handleDateConfirm(onConfirm)"
          >
            {{ text }}
          </NButton>
        </div>
      </template>
    </NDatePicker>
  </AppModal>
</template>

<style scoped>
.salary-add-deal-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
  width: 100%;
}

.salary-add-deal-form__columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  align-items: start;
}

.salary-add-deal-form__section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.salary-add-deal-form__section:first-child {
  padding-right: 32px;
}

.salary-add-deal-form__section + .salary-add-deal-form__section {
  padding-left: 32px;
  border-left: 1px solid #e2e8f0;
}

@media (max-width: 860px) {
  .salary-add-deal-form__columns {
    grid-template-columns: 1fr;
    gap: 0;
  }

  .salary-add-deal-form__section:first-child {
    padding-right: 0;
    padding-bottom: 24px;
  }

  .salary-add-deal-form__section + .salary-add-deal-form__section {
    padding-left: 0;
    padding-top: 24px;
    border-left: none;
    border-top: 1px solid #e2e8f0;
  }
}

.salary-add-deal-form__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.salary-add-deal-form__field--full {
  width: 100%;
}

.salary-add-deal-form__label {
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
}

.salary-add-deal-form__required {
  color: #dc2626;
  font-size: 12px;
  line-height: 1;
  font-weight: 600;
}

.salary-add-deal-form__control {
  width: 100%;
}

.salary-add-deal-form__hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.4;
  color: #94a3b8;
}

.salary-add-deal-form__actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.salary-add-deal-form__error {
  margin: 0;
  width: 100%;
  max-width: 420px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #b91c1c;
  font-size: 13px;
  line-height: 1.4;
  text-align: center;
}
</style>

<style>
/* Модалка телепортируется в body — нужен unscoped селектор */
.salary-add-deal-form__date-panel.n-date-panel .n-date-panel-header,
.n-date-panel.salary-add-deal-form__date-panel .n-date-panel-header {
  visibility: hidden;
  pointer-events: none;
  height: 8px;
  min-height: 8px;
  max-height: 8px;
  padding: 0;
  overflow: hidden;
  border-bottom: none;
  box-sizing: border-box;
}

.salary-add-deal-form__date-panel.n-date-panel .n-date-panel-header > *,
.n-date-panel.salary-add-deal-form__date-panel .n-date-panel-header > * {
  display: none;
}
</style>
