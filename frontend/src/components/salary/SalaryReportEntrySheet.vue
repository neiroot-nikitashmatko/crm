<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { NButton, NDatePicker, NInput, NSelect } from 'naive-ui'
import AppBottomSheet from '@/components/common/AppBottomSheet.vue'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import DateTimeField from '@/components/common/DateTimeField.vue'
import { SALARY_SERVICE_OPTIONS } from '@/constants/salary'
import { useDeals } from '@/composables/useDeals'
import { useSalaryReport } from '@/composables/useSalaryReport'
import type { SalaryReportEntry } from '@/types/salaryReport'
import {
  renderSalaryDealOption,
  salaryDealOptionFullLabel,
  salaryDealOptionNumberLabel,
} from '@/utils/salaryDealLabel'

const show = defineModel<boolean>('show', { required: true })

const props = defineProps<{
  entry: SalaryReportEntry
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const { deals, loadDeals } = useDeals()
const { updateReportEntry } = useSalaryReport()
const isDealsLoading = ref(false)
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
  dealId: null as string | null,
  date: null as number | null,
  service: null as string | null,
  salary: '',
  comment: '',
})

const isSubmitting = ref(false)
const errorMessage = ref('')

onMounted(() => {
  void loadDeals(true)
})

function isClosedDeal(deal: { columnId?: string; status?: string }) {
  return deal.columnId === 'closed' || String(deal.status ?? '').toLowerCase() === 'closed'
}

const closedDealOptions = computed(() => {
  const query = dealSearchQuery.value.trim().toLowerCase()
  const options = deals.value
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

  if (
    props.entry.dealId &&
    !options.some((option) => option.value === props.entry.dealId) &&
    (!query || props.entry.dealNumberLabel.toLowerCase().includes(query))
  ) {
    options.unshift({
      label: props.entry.dealNumberLabel,
      fullLabel: props.entry.dealNumberLabel,
      value: props.entry.dealId,
    })
  }

  return options
})

const serviceOptions = computed(() => {
  const options: Array<{ label: string; value: string }> = SALARY_SERVICE_OPTIONS.map((item) => ({
    label: item.label,
    value: item.value,
  }))

  if (props.entry.service && !options.some((option) => option.value === props.entry.service)) {
    options.unshift({
      label: props.entry.serviceLabel,
      value: props.entry.service,
    })
  }

  return options
})

const canSubmit = computed(
  () =>
    form.dealId !== null &&
    form.date !== null &&
    form.service !== null &&
    form.salary.trim().length > 0,
)

function fillFormFromEntry(entry: SalaryReportEntry) {
  form.dealId = entry.dealId
  form.date = entry.date
  form.service = entry.service
  form.salary = entry.salary
  form.comment = entry.comment
  dealSearchQuery.value = ''
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

function handleDealSearch(query: string) {
  dealSearchQuery.value = query
}

function keepAllFilteredOptions() {
  return true
}

function handleSalaryInput(value: string) {
  form.salary = value.replace(/\D/g, '')
}

async function handleSubmit() {
  if (
    !canSubmit.value ||
    isSubmitting.value ||
    form.dealId === null ||
    form.date === null ||
    form.service === null
  ) {
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''
  try {
    await updateReportEntry(props.entry.id, {
      date: form.date,
      dealId: form.dealId,
      service: form.service,
      salary: form.salary.trim(),
      comment: form.comment.trim(),
    })
    emit('saved')
    show.value = false
  } catch (error) {
    errorMessage.value =
      error instanceof Error && error.message ? error.message : 'Не удалось сохранить изменения'
  } finally {
    isSubmitting.value = false
  }
}

watch(
  () => [show.value, props.entry] as const,
  ([isOpen, entry]) => {
    if (isOpen) {
      fillFormFromEntry(entry)
      void loadDeals(true).finally(() => {
        isDealsLoading.value = false
      })
      isDealsLoading.value = true
    }
  },
  { immediate: true },
)
</script>

<template>
  <AppBottomSheet
    v-model:show="show"
    title="Редактирование записи"
    body-align="center"
    close-label="Закрыть запись зарплаты"
    @close="emit('close')"
  >
    <form class="salary-report-entry-form" @submit.prevent="handleSubmit">
      <div class="salary-report-entry-form__columns">
        <section class="salary-report-entry-form__section">
          <div class="salary-report-entry-form__field">
            <span class="salary-report-entry-form__label">
              Номер сделки
              <span class="salary-report-entry-form__required" aria-hidden="true">*</span>
            </span>
            <NSelect
              v-model:value="form.dealId"
              class="salary-report-entry-form__control"
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
          </div>

          <div class="salary-report-entry-form__field">
            <span class="salary-report-entry-form__label">
              Дата
              <span class="salary-report-entry-form__required" aria-hidden="true">*</span>
            </span>
            <DateTimeField
              :display-value="formatDate(form.date)"
              :has-value="form.date !== null"
              @open="openDateModal"
              @clear="clearDate"
            />
          </div>
        </section>

        <section class="salary-report-entry-form__section">
          <label class="salary-report-entry-form__field">
            <span class="salary-report-entry-form__label">
              Услуга/Товар
              <span class="salary-report-entry-form__required" aria-hidden="true">*</span>
            </span>
            <NSelect
              v-model:value="form.service"
              class="salary-report-entry-form__control"
              :theme-overrides="fieldSelectTheme"
              :options="serviceOptions"
              clearable
              placeholder="Выберите услугу"
            />
          </label>

          <label class="salary-report-entry-form__field">
            <span class="salary-report-entry-form__label">
              Зарплата сотрудника
              <span class="salary-report-entry-form__required" aria-hidden="true">*</span>
            </span>
            <NInput
              :value="form.salary"
              class="salary-report-entry-form__control"
              :theme-overrides="fieldInputTheme"
              placeholder="Введите сумму"
              inputmode="numeric"
              autocomplete="off"
              @update:value="handleSalaryInput"
            />
          </label>
        </section>
      </div>

      <label class="salary-report-entry-form__field salary-report-entry-form__field--full">
        <span class="salary-report-entry-form__label">Комментарий</span>
        <NInput
          v-model:value="form.comment"
          class="salary-report-entry-form__control"
          :theme-overrides="fieldInputTheme"
          type="textarea"
          :autosize="{ minRows: 4, maxRows: 8 }"
          placeholder="Комментарий (необязательно)"
        />
      </label>
    </form>

    <template #actions>
      <p v-if="errorMessage" class="salary-report-entry-form__error" role="alert">
        {{ errorMessage }}
      </p>
      <AppModalButton :disabled="!canSubmit || isSubmitting" @click="handleSubmit">
        {{ isSubmitting ? 'Сохранение…' : 'Сохранить изменения' }}
      </AppModalButton>
    </template>
  </AppBottomSheet>

  <AppModal
    v-model:show="isDateModalOpen"
    title="Дата завершения сделки"
    width="wide"
    body-variant="date"
    close-label="Закрыть выбор даты"
  >
    <NDatePicker
      v-model:value="dateDraft"
      class="app-modal__date-panel salary-report-entry-form__date-panel"
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
.salary-report-entry-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
  max-width: 880px;
  margin: 0 auto;
}

.salary-report-entry-form__columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  align-items: start;
}

.salary-report-entry-form__section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.salary-report-entry-form__section:first-child {
  padding-right: 32px;
}

.salary-report-entry-form__section + .salary-report-entry-form__section {
  padding-left: 32px;
  border-left: 1px solid #e2e8f0;
}

@media (max-width: 860px) {
  .salary-report-entry-form__columns {
    grid-template-columns: 1fr;
    gap: 0;
  }

  .salary-report-entry-form__section:first-child {
    padding-right: 0;
    padding-bottom: 24px;
  }

  .salary-report-entry-form__section + .salary-report-entry-form__section {
    padding-left: 0;
    padding-top: 24px;
    border-left: none;
    border-top: 1px solid #e2e8f0;
  }
}

.salary-report-entry-form__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.salary-report-entry-form__field--full {
  width: 100%;
}

.salary-report-entry-form__label {
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
}

.salary-report-entry-form__required {
  color: #dc2626;
  font-size: 12px;
  line-height: 1;
  font-weight: 600;
}

.salary-report-entry-form__control {
  width: 100%;
}

.salary-report-entry-form__error {
  margin: 0 0 8px;
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
.salary-report-entry-form__date-panel.n-date-panel .n-date-panel-header,
.n-date-panel.salary-report-entry-form__date-panel .n-date-panel-header {
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

.salary-report-entry-form__date-panel.n-date-panel .n-date-panel-header > *,
.n-date-panel.salary-report-entry-form__date-panel .n-date-panel-header > * {
  display: none;
}
</style>
