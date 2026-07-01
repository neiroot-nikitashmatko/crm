<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NIcon } from 'naive-ui'
import { PencilOutline, TrashOutline } from '@vicons/ionicons5'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import EmployeeDetailsSheet from '@/components/employees/EmployeeDetailsSheet.vue'
import { getEmployeeRoleLabel } from '@/constants/employees'
import { useEmployees } from '@/composables/useEmployees'
import { UsersApiError } from '@/api/users'
import type { Employee } from '@/types/employee'

const { employees, isLoading, loadEmployees, removeEmployee } = useEmployees()

const selectedEmployee = ref<Employee | null>(null)
const isDetailsOpen = ref(false)
const employeeToDelete = ref<Employee | null>(null)
const isDeleteModalOpen = ref(false)
const isDeleting = ref(false)
const errorMessage = ref('')

function openEmployeeDetails(employee: Employee) {
  selectedEmployee.value = employee
  isDetailsOpen.value = true
}

function closeEmployeeDetails() {
  isDetailsOpen.value = false
  selectedEmployee.value = null
}

function handleEditEmployee(employee: Employee) {
  openEmployeeDetails(employee)
}

function handleDeleteEmployee(employee: Employee) {
  employeeToDelete.value = employee
  isDeleteModalOpen.value = true
}

function closeDeleteModal() {
  isDeleteModalOpen.value = false
  employeeToDelete.value = null
}

async function confirmDeleteEmployee() {
  if (!employeeToDelete.value || isDeleting.value) return

  isDeleting.value = true
  errorMessage.value = ''

  try {
    const deletedId = employeeToDelete.value.id
    await removeEmployee(deletedId)

    if (selectedEmployee.value?.id === deletedId) {
      closeEmployeeDetails()
    }

    closeDeleteModal()
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
    closeDeleteModal()
  } finally {
    isDeleting.value = false
  }
}

function getErrorMessage(error: unknown) {
  if (error instanceof UsersApiError) return error.message
  if (error instanceof Error) return error.message
  return 'Не удалось выполнить операцию'
}

onMounted(async () => {
  errorMessage.value = ''
  try {
    await loadEmployees(true)
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  }
})
</script>

<template>
  <div class="employees-list-view">
    <header class="employees-list-view__header">
      <h2 class="employees-list-view__title">Список сотрудников</h2>
      <p class="employees-list-view__description">
        Все сотрудники компании. Карточку можно открыть через кнопку редактирования.
      </p>
    </header>

    <p v-if="errorMessage" class="employees-list-view__error" role="alert">
      {{ errorMessage }}
    </p>

    <section v-if="isLoading" class="employees-list-view__placeholder">
      <p class="employees-list-view__placeholder-text">Загрузка списка сотрудников…</p>
    </section>

    <section v-else-if="employees.length === 0" class="employees-list-view__placeholder">
      <p class="employees-list-view__placeholder-text">
        Пока нет сотрудников. Добавьте первого через пункт меню «Добавить сотрудника».
      </p>
    </section>

    <section v-else class="employees-list-view__table-wrap">
      <div class="employees-list-view__table" role="table">
        <div class="employees-list-view__table-row employees-list-view__table-row--head" role="row">
          <span class="employees-list-view__cell employees-list-view__cell--head" role="columnheader">Имя</span>
          <span class="employees-list-view__cell employees-list-view__cell--head" role="columnheader">Фамилия</span>
          <span class="employees-list-view__cell employees-list-view__cell--head" role="columnheader">Отчество</span>
          <span class="employees-list-view__cell employees-list-view__cell--head employees-list-view__cell--compact" role="columnheader">
            Телефон
          </span>
          <span class="employees-list-view__cell employees-list-view__cell--head employees-list-view__cell--compact" role="columnheader">
            Должность
          </span>
          <span class="employees-list-view__cell employees-list-view__cell--head employees-list-view__cell--compact" role="columnheader">
            Роль
          </span>
          <span
            class="employees-list-view__cell employees-list-view__cell--head employees-list-view__cell--actions"
            role="columnheader"
            aria-hidden="true"
          />
        </div>

        <div
          v-for="employee in employees"
          :key="employee.id"
          class="employees-list-view__table-row"
          role="row"
        >
          <span class="employees-list-view__cell employees-list-view__cell--name">{{ employee.firstName }}</span>
          <span class="employees-list-view__cell employees-list-view__cell--name">{{ employee.lastName }}</span>
          <span class="employees-list-view__cell employees-list-view__cell--name">{{ employee.patronymic }}</span>
          <span class="employees-list-view__cell employees-list-view__cell--compact employees-list-view__cell--phone">
            {{ employee.phone }}
          </span>
          <span class="employees-list-view__cell employees-list-view__cell--compact employees-list-view__cell--position">
            {{ employee.position }}
          </span>
          <span class="employees-list-view__cell employees-list-view__cell--compact employees-list-view__cell--role">
            {{ getEmployeeRoleLabel(employee.role) }}
          </span>
          <div class="employees-list-view__cell employees-list-view__cell--actions">
            <div class="employees-list-view__row-actions">
              <button
                type="button"
                class="employees-list-view__icon-action"
                aria-label="Редактировать сотрудника"
                @click="handleEditEmployee(employee)"
              >
                <NIcon :size="16">
                  <PencilOutline />
                </NIcon>
              </button>
              <button
                type="button"
                class="employees-list-view__icon-action employees-list-view__icon-action--danger"
                aria-label="Удалить сотрудника"
                @click="handleDeleteEmployee(employee)"
              >
                <NIcon :size="16">
                  <TrashOutline />
                </NIcon>
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <EmployeeDetailsSheet
      v-if="selectedEmployee"
      v-model:show="isDetailsOpen"
      :employee="selectedEmployee"
      @close="closeEmployeeDetails"
    />

    <AppModal
      v-model:show="isDeleteModalOpen"
      title="Удаление сотрудника"
      body-variant="center"
      :mask-closable="!isDeleting"
      @close="closeDeleteModal"
    >
      <p class="app-modal__message">
        Вы уверены, что хотете удалить данного сотрудника?
      </p>

      <template #actions>
        <div class="employees-list-view__confirm-actions">
          <AppModalButton :disabled="isDeleting" @click="confirmDeleteEmployee">
            Да
          </AppModalButton>
          <button
            type="button"
            class="employees-list-view__confirm-cancel"
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
.employees-list-view {
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
  min-width: 0;
}

.employees-list-view__header {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.employees-list-view__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.employees-list-view__description {
  margin: 0;
  font-size: 14px;
  line-height: 1.45;
  color: #64748b;
}

.employees-list-view__error {
  margin: 0;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #b91c1c;
  font-size: 13px;
  line-height: 1.4;
}

.employees-list-view__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 240px;
  padding: 32px 24px;
  border: 1px dashed #cbd5e1;
  border-radius: 12px;
  background: #f8fafc;
}

.employees-list-view__placeholder-text {
  margin: 0;
  max-width: 420px;
  font-size: 15px;
  line-height: 1.5;
  color: #64748b;
  text-align: center;
}

.employees-list-view__table-wrap {
  min-width: 0;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
}

.employees-list-view__table {
  display: grid;
  width: 100%;
  grid-template-columns:
    minmax(max-content, 1fr)
    minmax(max-content, 1fr)
    minmax(max-content, 1fr)
    max-content
    max-content
    max-content
    max-content;
}

.employees-list-view__table-row {
  display: contents;
}

.employees-list-view__cell {
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

.employees-list-view__cell--head {
  min-height: 44px;
  padding: 10px 14px;
  background: #f8fafc;
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.employees-list-view__table-row .employees-list-view__cell:nth-child(7) {
  border-right: 0;
}

.employees-list-view__table-row:last-child .employees-list-view__cell {
  border-bottom: 0;
}

.employees-list-view__cell--name {
  white-space: nowrap;
}

.employees-list-view__cell--compact {
  padding-left: 10px;
  padding-right: 10px;
  white-space: nowrap;
}

.employees-list-view__cell--head.employees-list-view__cell--compact {
  padding-left: 10px;
  padding-right: 10px;
}

.employees-list-view__cell--phone {
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.01em;
}

.employees-list-view__cell--role {
  color: #475569;
}

.employees-list-view__cell--actions {
  justify-content: center;
  padding: 10px;
  white-space: nowrap;
}

.employees-list-view__row-actions {
  display: inline-flex;
  justify-content: center;
  gap: 6px;
}

.employees-list-view__icon-action {
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

.employees-list-view__icon-action:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #334155;
}

.employees-list-view__icon-action--danger:hover {
  color: #dc2626;
}

.employees-list-view__table-row:not(.employees-list-view__table-row--head):hover .employees-list-view__cell {
  background: #f8fafc;
}

.employees-list-view__confirm-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
}

.employees-list-view__confirm-cancel {
  min-width: min(100%, 220px);
  padding: 10px 20px;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  background: #ffffff;
  color: #475569;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.35;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.employees-list-view__confirm-cancel:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #94a3b8;
  color: #334155;
}

.employees-list-view__confirm-cancel:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
