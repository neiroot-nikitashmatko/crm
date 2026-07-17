<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { EyeOutline, TrashOutline } from '@vicons/ionicons5'
import { NIcon, NSelect } from 'naive-ui'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import DealDetailsSheet from '@/components/deals/DealDetailsSheet.vue'
import SalaryReportEntrySheet from '@/components/salary/SalaryReportEntrySheet.vue'
import { useAuth } from '@/composables/useAuth'
import { useDeals } from '@/composables/useDeals'
import { useEmployees } from '@/composables/useEmployees'
import { useSalaryReport } from '@/composables/useSalaryReport'
import type { SalaryReportEntry } from '@/types/salaryReport'
import { formatSalaryEmployeeName, isSalaryMasterEmployee } from '@/utils/salaryEmployee'

const now = new Date()
const selectedMonth = ref(now.getMonth())
const selectedYear = ref(now.getFullYear())
const selectedEmployeeId = ref<string | null>(null)

const { isAdmin } = useAuth()
const { deals, loadDeals } = useDeals()
const { employees, loadEmployees } = useEmployees()
const { reportEntries, isLoading, loadReportEntries, removeReportEntry } = useSalaryReport()

const isEntrySheetOpen = ref(false)
const selectedEntry = ref<SalaryReportEntry | null>(null)
const entryToDelete = ref<SalaryReportEntry | null>(null)
const isDeleteModalOpen = ref(false)
const isDeleting = ref(false)
const actionErrorMessage = ref('')
const selectedDealId = ref<string | null>(null)

onMounted(() => {
  void loadReportEntries(true)
  if (isAdmin.value) {
    void loadEmployees(true)
  }
})

const periodSelectTheme = {
  peers: {
    InternalSelection: {
      border: '1px solid #d1d9e2',
      borderHover: '1px solid #cbd5e1',
      borderFocus: '1px solid #93c5fd',
      borderActive: '1px solid #93c5fd',
      boxShadowFocus: '0 0 0 3px rgba(147, 197, 253, 0.25)',
      boxShadowActive: '0 0 0 3px rgba(147, 197, 253, 0.25)',
      boxShadowHover: 'none',
      borderRadius: '8px',
      heightMedium: '32px',
      fontSizeMedium: '13px',
      paddingSingle: '0 26px 0 8px',
      arrowSize: '14px',
    },
  },
}

const employeeFilterOptions = computed(() =>
  employees.value
    .filter((employee) => isSalaryMasterEmployee(employee))
    .slice()
    .sort((left, right) => formatSalaryEmployeeName(left).localeCompare(formatSalaryEmployeeName(right), 'ru'))
    .map((employee) => ({
      label: formatSalaryEmployeeName(employee),
      value: employee.id,
    })),
)

const monthOptions = [
  { label: 'Январь', value: 0 },
  { label: 'Февраль', value: 1 },
  { label: 'Март', value: 2 },
  { label: 'Апрель', value: 3 },
  { label: 'Май', value: 4 },
  { label: 'Июнь', value: 5 },
  { label: 'Июль', value: 6 },
  { label: 'Август', value: 7 },
  { label: 'Сентябрь', value: 8 },
  { label: 'Октябрь', value: 9 },
  { label: 'Ноябрь', value: 10 },
  { label: 'Декабрь', value: 11 },
]

const yearOptions = computed(() => {
  const currentYear = now.getFullYear()
  const years = new Set<number>([currentYear - 1, currentYear, currentYear + 1])
  for (const row of reportEntries.value) {
    years.add(new Date(row.date).getFullYear())
  }
  years.add(selectedYear.value)
  return [...years]
    .sort((left, right) => right - left)
    .map((year) => ({ label: String(year), value: year }))
})

const filteredReportRows = computed(() =>
  reportEntries.value.filter((row) => {
    const date = new Date(row.date)
    const matchesPeriod =
      date.getMonth() === selectedMonth.value && date.getFullYear() === selectedYear.value
    if (!matchesPeriod) return false
    if (!isAdmin.value) return true
    if (!selectedEmployeeId.value) return false
    return row.createdBy === selectedEmployeeId.value
  }),
)

const emptyReportText = computed(() => {
  if (isAdmin.value && !selectedEmployeeId.value) {
    return 'Выберите сотрудника'
  }
  return 'За выбранный период записей нет'
})

const monthSalaryTotal = computed(() =>
  filteredReportRows.value.reduce((sum, row) => sum + (Number(row.salary) || 0), 0),
)

const monthSalaryLabel = computed(() =>
  new Intl.NumberFormat('ru-RU').format(monthSalaryTotal.value),
)

function formatReportDate(timestamp: number) {
  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(timestamp)
}

function handleViewRow(row: SalaryReportEntry) {
  selectedEntry.value = row
  isEntrySheetOpen.value = true
}

async function handleOpenDeal(dealId: string) {
  if (!dealId) return

  if (!deals.value.some((deal) => deal.id === dealId)) {
    await loadDeals(true)
  }

  if (deals.value.some((deal) => deal.id === dealId)) {
    selectedDealId.value = dealId
  }
}

function handleCloseDealSheet() {
  selectedDealId.value = null
}

function handleDeleteRow(row: SalaryReportEntry) {
  entryToDelete.value = row
  isDeleteModalOpen.value = true
}

function closeDeleteModal() {
  if (isDeleting.value) return
  isDeleteModalOpen.value = false
  entryToDelete.value = null
  actionErrorMessage.value = ''
}

async function confirmDeleteEntry() {
  if (!entryToDelete.value || isDeleting.value) return

  const deletedId = entryToDelete.value.id
  isDeleting.value = true
  actionErrorMessage.value = ''
  try {
    await removeReportEntry(deletedId)

    if (selectedEntry.value?.id === deletedId) {
      closeEntrySheet()
    }

    isDeleteModalOpen.value = false
    entryToDelete.value = null
  } catch (error) {
    actionErrorMessage.value =
      error instanceof Error && error.message ? error.message : 'Не удалось удалить запись'
  } finally {
    isDeleting.value = false
  }
}

function closeEntrySheet() {
  isEntrySheetOpen.value = false
  selectedEntry.value = null
}

function handleEntrySaved() {
  closeEntrySheet()
}
</script>

<template>
  <div class="salary-report-view">
    <section class="salary-report-view__panel">
      <header class="salary-report-view__panel-header">
        <h2 class="salary-report-view__title">Отчёт</h2>

        <div class="salary-report-view__actions">
          <p class="salary-report-view__month-total">
            <span class="salary-report-view__month-total-label">Зарплата за месяц:</span>
            <span class="salary-report-view__month-total-value">{{ monthSalaryLabel }} ₽</span>
          </p>

          <template v-if="isAdmin">
            <span class="salary-report-view__actions-divider" aria-hidden="true" />
            <NSelect
              v-model:value="selectedEmployeeId"
              class="salary-report-view__employee-select"
              :theme-overrides="periodSelectTheme"
              :options="employeeFilterOptions"
              :consistent-menu-width="false"
              clearable
              placeholder="Выберите сотрудника"
            />
          </template>

          <span class="salary-report-view__actions-divider" aria-hidden="true" />

          <div class="salary-report-view__period">
            <NSelect
              v-model:value="selectedMonth"
              class="salary-report-view__period-select salary-report-view__period-select--month"
              :theme-overrides="periodSelectTheme"
              :options="monthOptions"
              :consistent-menu-width="false"
            />
            <NSelect
              v-model:value="selectedYear"
              class="salary-report-view__period-select salary-report-view__period-select--year"
              :theme-overrides="periodSelectTheme"
              :options="yearOptions"
              :consistent-menu-width="false"
            />
          </div>
        </div>
      </header>

      <div class="salary-report-view__panel-body">
        <div v-if="isLoading" class="salary-report-view__empty">
          <p class="salary-report-view__empty-text">Загрузка…</p>
        </div>

        <div v-else-if="filteredReportRows.length === 0" class="salary-report-view__empty">
          <p class="salary-report-view__empty-text">{{ emptyReportText }}</p>
        </div>

        <div v-else class="salary-report-view__table-wrap">
          <div class="salary-report-view__table" role="table">
            <div class="salary-report-view__table-row salary-report-view__table-row--head" role="row">
              <span class="salary-report-view__cell salary-report-view__cell--head" role="columnheader">
                Дата
              </span>
              <span
                class="salary-report-view__cell salary-report-view__cell--head salary-report-view__cell--compact"
                role="columnheader"
              >
                Номер сделки
              </span>
              <span class="salary-report-view__cell salary-report-view__cell--head" role="columnheader">
                Услуга/Товар
              </span>
              <span
                class="salary-report-view__cell salary-report-view__cell--head salary-report-view__cell--compact"
                role="columnheader"
              >
                Зарплата сотрудника
              </span>
              <span
                class="salary-report-view__cell salary-report-view__cell--head salary-report-view__cell--actions"
                role="columnheader"
                aria-hidden="true"
              />
            </div>

            <div
              v-for="row in filteredReportRows"
              :key="row.id"
              class="salary-report-view__table-row"
              role="row"
            >
              <span class="salary-report-view__cell salary-report-view__cell--date">
                {{ formatReportDate(row.date) }}
              </span>
              <span class="salary-report-view__cell salary-report-view__cell--compact salary-report-view__cell--deal">
                <button
                  type="button"
                  class="salary-report-view__deal-link"
                  :aria-label="`Открыть сделку ${row.dealNumberLabel}`"
                  @click="handleOpenDeal(row.dealId)"
                >
                  {{ row.dealNumberLabel }}
                </button>
              </span>
              <span class="salary-report-view__cell salary-report-view__cell--service">
                {{ row.serviceLabel }}
              </span>
              <span class="salary-report-view__cell salary-report-view__cell--compact salary-report-view__cell--salary">
                {{ row.salary }}
              </span>
              <div class="salary-report-view__cell salary-report-view__cell--actions">
                <div class="salary-report-view__row-actions">
                  <button
                    type="button"
                    class="salary-report-view__icon-action"
                    aria-label="Просмотреть запись"
                    title="Просмотреть"
                    @click="handleViewRow(row)"
                  >
                    <NIcon :size="16">
                      <EyeOutline />
                    </NIcon>
                  </button>
                  <button
                    type="button"
                    class="salary-report-view__icon-action salary-report-view__icon-action--danger"
                    aria-label="Удалить запись"
                    title="Удалить"
                    @click="handleDeleteRow(row)"
                  >
                    <NIcon :size="16">
                      <TrashOutline />
                    </NIcon>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <SalaryReportEntrySheet
      v-if="selectedEntry"
      v-model:show="isEntrySheetOpen"
      :entry="selectedEntry"
      @close="closeEntrySheet"
      @saved="handleEntrySaved"
    />

    <DealDetailsSheet :deal-id="selectedDealId" @close="handleCloseDealSheet" />

    <AppModal
      v-model:show="isDeleteModalOpen"
      title="Удаление записи"
      body-variant="center"
      @close="closeDeleteModal"
    >
      <p class="app-modal__message">Вы уверены, что хотите удалить данную запись?</p>
      <p v-if="actionErrorMessage" class="salary-report-view__action-error" role="alert">
        {{ actionErrorMessage }}
      </p>

      <template #actions>
        <div class="salary-report-view__confirm-actions">
          <AppModalButton :disabled="isDeleting" @click="confirmDeleteEntry">
            {{ isDeleting ? 'Удаление…' : 'Да' }}
          </AppModalButton>
          <button
            type="button"
            class="salary-report-view__confirm-cancel"
            :disabled="isDeleting"
            @click="closeDeleteModal"
          >
            Нет
          </button>
        </div>
      </template>
    </AppModal>
  </div>
</template>

<style scoped>
.salary-report-view {
  display: flex;
  flex-direction: column;
  width: 100%;
  min-width: 0;
  height: calc(100dvh - 64px - 56px - 48px);
  max-height: calc(100dvh - 64px - 56px - 48px);
}

.salary-report-view__panel {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
  overflow: hidden;
}

.salary-report-view__panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-shrink: 0;
  min-height: 56px;
  padding: 12px 20px;
  border-bottom: 1px solid #e2e8f0;
  box-sizing: border-box;
  background: #ffffff;
}

.salary-report-view__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.salary-report-view__actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 12px;
  flex-shrink: 0;
  min-height: 32px;
}

.salary-report-view__month-total {
  margin: 0;
  display: inline-flex;
  align-items: baseline;
  gap: 6px;
  white-space: nowrap;
}

.salary-report-view__month-total-label,
.salary-report-view__month-total-value {
  font-size: 13px;
  font-weight: 400;
  color: #64748b;
}

.salary-report-view__month-total-value {
  font-variant-numeric: tabular-nums;
}

.salary-report-view__actions-divider {
  flex-shrink: 0;
  align-self: center;
  width: 1px;
  height: 18px;
  background: #d1d9e2;
}

.salary-report-view__employee-select {
  flex-shrink: 0;
  width: 190px;
}

.salary-report-view__period {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.salary-report-view__period-select {
  flex-shrink: 0;
}

.salary-report-view__period-select--month {
  width: 96px;
}

.salary-report-view__period-select--year {
  width: 68px;
}

.salary-report-view__panel-body {
  flex: 1;
  min-height: 0;
  padding: 20px;
  overflow: auto;
}

.salary-report-view__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 180px;
  border: 1px dashed #cbd5e1;
  border-radius: 12px;
  background: #f8fafc;
}

.salary-report-view__empty-text {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

.salary-report-view__table-wrap {
  min-width: 0;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
  overflow-x: auto;
}

.salary-report-view__table {
  display: grid;
  width: 100%;
  min-width: 760px;
  grid-template-columns:
    max-content
    max-content
    minmax(220px, 1fr)
    max-content
    78px;
}

.salary-report-view__table-row {
  display: contents;
}

.salary-report-view__cell {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  min-height: 48px;
  padding: 12px 14px;
  border-right: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
  font-size: 14px;
  line-height: 1.35;
  color: #1a202c;
  background: #ffffff;
}

.salary-report-view__cell--head {
  min-height: 44px;
  padding: 10px 14px;
  background: #f8fafc;
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.salary-report-view__table-row .salary-report-view__cell:nth-child(5) {
  border-right: 0;
}

.salary-report-view__table-row:last-child .salary-report-view__cell {
  border-bottom: 0;
}

.salary-report-view__cell--compact {
  padding-left: 10px;
  padding-right: 10px;
  white-space: nowrap;
}

.salary-report-view__cell--head.salary-report-view__cell--compact {
  padding-left: 10px;
  padding-right: 10px;
}

.salary-report-view__cell--date,
.salary-report-view__cell--deal {
  white-space: nowrap;
}

.salary-report-view__deal-link {
  margin: 0;
  padding: 0;
  border: none;
  background: transparent;
  color: #1d4ed8;
  font: inherit;
  cursor: pointer;
  text-decoration: none;
}

.salary-report-view__deal-link:hover {
  color: #1e40af;
  text-decoration: underline;
}

.salary-report-view__deal-link:focus-visible {
  outline: 2px solid #93c5fd;
  outline-offset: 2px;
  border-radius: 4px;
}

.salary-report-view__cell--salary {
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.01em;
}

.salary-report-view__cell--service {
  white-space: normal;
}

.salary-report-view__cell--actions {
  justify-content: center;
  width: 78px;
  min-width: 78px;
  max-width: 78px;
  padding: 10px 8px;
  white-space: nowrap;
}

.salary-report-view__row-actions {
  display: inline-flex;
  justify-content: center;
  gap: 6px;
  min-height: 28px;
}

.salary-report-view__icon-action {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #64748b;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.salary-report-view__icon-action:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #334155;
}

.salary-report-view__icon-action--danger:hover {
  color: #dc2626;
}

.salary-report-view__table-row:not(.salary-report-view__table-row--head):hover .salary-report-view__cell {
  background: #f8fafc;
}

.salary-report-view__confirm-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
}

.salary-report-view__action-error {
  margin: 12px 0 0;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #b91c1c;
  font-size: 13px;
  line-height: 1.4;
  text-align: center;
}

.salary-report-view__confirm-cancel {
  min-width: min(100%, 220px);
  padding: 10px 20px;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  background: #ffffff;
  color: #334155;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease;
}

.salary-report-view__confirm-cancel:hover {
  background: #f8fafc;
  border-color: #94a3b8;
}

@media (max-width: 900px) {
  .salary-report-view__panel-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .salary-report-view__actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>
