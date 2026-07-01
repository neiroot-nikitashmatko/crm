<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { NDatePicker, NInput, NSelect } from 'naive-ui'
import AppBottomSheet from '@/components/common/AppBottomSheet.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import { birthDateFromTimestamp, birthDateToTimestamp, UsersApiError } from '@/api/users'
import { useEmployees } from '@/composables/useEmployees'
import type { Employee, EmployeeRole } from '@/types/employee'
import { isPhoneFilled, normalizePhone } from '@/utils/phone'

const show = defineModel<boolean>('show', { required: true })

const props = defineProps<{
  employee: Employee
}>()

const emit = defineEmits<{
  close: []
  saved: [employee: Employee]
}>()

const { editEmployee } = useEmployees()
const isSubmitting = ref(false)
const errorMessage = ref('')

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

const fieldDatePickerTheme = {
  peers: {
    Input: fieldInputTheme,
  },
}

const employeeForm = reactive({
  firstName: '',
  lastName: '',
  patronymic: '',
  birthDate: null as number | null,
  phone: '',
  password: '',
  position: '',
  role: null as EmployeeRole | null,
})

const roleOptions = [
  { label: 'Администратор', value: 'admin' as const },
  { label: 'Пользователь', value: 'manager' as const },
]

const canSubmit = computed(
  () =>
    employeeForm.firstName.trim().length > 0 &&
    employeeForm.lastName.trim().length > 0 &&
    employeeForm.patronymic.trim().length > 0 &&
    employeeForm.birthDate !== null &&
    isPhoneFilled(employeeForm.phone) &&
    employeeForm.position.trim().length > 0 &&
    employeeForm.role !== null,
)

function fillFormFromEmployee(employee: Employee) {
  employeeForm.firstName = employee.firstName
  employeeForm.lastName = employee.lastName
  employeeForm.patronymic = employee.patronymic
  employeeForm.birthDate = birthDateToTimestamp(employee.birthDate)
  employeeForm.phone = employee.phone
  employeeForm.password = ''
  employeeForm.position = employee.position
  employeeForm.role = employee.role
  errorMessage.value = ''
}

function handlePhoneInput(value: string) {
  employeeForm.phone = normalizePhone(value)
}

function handleSubmit() {
  if (!canSubmit.value || isSubmitting.value || employeeForm.role === null) return

  void submitEmployee()
}

async function submitEmployee() {
  if (employeeForm.role === null) return

  isSubmitting.value = true
  errorMessage.value = ''

  const birthDate = birthDateFromTimestamp(employeeForm.birthDate)
  if (!birthDate) {
    errorMessage.value = 'Укажите дату рождения'
    isSubmitting.value = false
    return
  }

  try {
    const updated = await editEmployee(props.employee.id, {
      firstName: employeeForm.firstName.trim(),
      lastName: employeeForm.lastName.trim(),
      patronymic: employeeForm.patronymic.trim(),
      birthDate,
      phone: employeeForm.phone.trim(),
      password: employeeForm.password,
      position: employeeForm.position.trim(),
      role: employeeForm.role,
    })
    emit('saved', updated)
    show.value = false
  } catch (error) {
    if (error instanceof UsersApiError) {
      errorMessage.value = error.message
    } else if (error instanceof Error) {
      errorMessage.value = error.message
    } else {
      errorMessage.value = 'Не удалось сохранить изменения'
    }
  } finally {
    isSubmitting.value = false
  }
}

watch(
  () => [show.value, props.employee] as const,
  ([isOpen, employee]) => {
    if (isOpen) {
      fillFormFromEmployee(employee)
    }
  },
  { immediate: true },
)
</script>

<template>
  <AppBottomSheet
    v-model:show="show"
    title="Редактирование сотрудника"
    body-align="center"
    close-label="Закрыть карточку сотрудника"
    @close="emit('close')"
  >
    <form class="employee-edit-form" @submit.prevent="handleSubmit">
      <div class="employee-edit-form__columns">
        <section class="employee-edit-form__section">
          <h3 class="employee-edit-form__section-title">Личные данные</h3>

          <label class="employee-edit-form__field">
            <span class="employee-edit-form__label">
              Фамилия
              <span class="employee-edit-form__required" aria-hidden="true">*</span>
            </span>
            <NInput
              v-model:value="employeeForm.lastName"
              class="employee-edit-form__control"
              :theme-overrides="fieldInputTheme"
              placeholder="Иванов"
              autocomplete="off"
            />
          </label>

          <label class="employee-edit-form__field">
            <span class="employee-edit-form__label">
              Имя
              <span class="employee-edit-form__required" aria-hidden="true">*</span>
            </span>
            <NInput
              v-model:value="employeeForm.firstName"
              class="employee-edit-form__control"
              :theme-overrides="fieldInputTheme"
              placeholder="Иван"
              autocomplete="off"
            />
          </label>

          <label class="employee-edit-form__field">
            <span class="employee-edit-form__label">
              Отчество
              <span class="employee-edit-form__required" aria-hidden="true">*</span>
            </span>
            <NInput
              v-model:value="employeeForm.patronymic"
              class="employee-edit-form__control"
              :theme-overrides="fieldInputTheme"
              placeholder="Иванович"
              autocomplete="off"
            />
          </label>

          <label class="employee-edit-form__field">
            <span class="employee-edit-form__label">
              Дата рождения
              <span class="employee-edit-form__required" aria-hidden="true">*</span>
            </span>
            <NDatePicker
              v-model:value="employeeForm.birthDate"
              class="employee-edit-form__control"
              :theme-overrides="fieldDatePickerTheme"
              type="date"
              clearable
              format="dd.MM.yyyy"
              placeholder="Выберите дату"
            />
          </label>
        </section>

        <section class="employee-edit-form__section">
          <h3 class="employee-edit-form__section-title">Доступ в систему</h3>

          <label class="employee-edit-form__field">
            <span class="employee-edit-form__label">
              Номер телефона
              <span class="employee-edit-form__required" aria-hidden="true">*</span>
            </span>
            <NInput
              v-model:value="employeeForm.phone"
              class="employee-edit-form__control"
              :theme-overrides="fieldInputTheme"
              placeholder="+79001234567"
              :maxlength="12"
              @update:value="handlePhoneInput"
            />
          </label>

          <label class="employee-edit-form__field">
            <span class="employee-edit-form__label">Новый пароль для входа</span>
            <NInput
              v-model:value="employeeForm.password"
              class="employee-edit-form__control"
              :theme-overrides="fieldInputTheme"
              type="password"
              show-password-on="click"
              placeholder="Оставьте пустым, чтобы не менять"
              autocomplete="new-password"
            />
          </label>

          <label class="employee-edit-form__field">
            <span class="employee-edit-form__label">
              Должность
              <span class="employee-edit-form__required" aria-hidden="true">*</span>
            </span>
            <NInput
              v-model:value="employeeForm.position"
              class="employee-edit-form__control"
              :theme-overrides="fieldInputTheme"
              placeholder="Например, менеджер по продажам"
              autocomplete="off"
            />
          </label>

          <label class="employee-edit-form__field">
            <span class="employee-edit-form__label">
              Роль
              <span class="employee-edit-form__required" aria-hidden="true">*</span>
            </span>
            <NSelect
              v-model:value="employeeForm.role"
              class="employee-edit-form__control"
              :theme-overrides="fieldSelectTheme"
              :options="roleOptions"
              placeholder="Выберите роль"
            />
          </label>
        </section>
      </div>

      <p v-if="errorMessage" class="employee-edit-form__error" role="alert">
        {{ errorMessage }}
      </p>
    </form>

    <template #actions>
      <AppModalButton :disabled="!canSubmit || isSubmitting" @click="handleSubmit">
        {{ isSubmitting ? 'Сохранение…' : 'Сохранить изменения' }}
      </AppModalButton>
    </template>
  </AppBottomSheet>
</template>

<style scoped>
.employee-edit-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
  max-width: 880px;
  margin: 0 auto;
}

.employee-edit-form__columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  align-items: start;
}

.employee-edit-form__section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.employee-edit-form__section:first-child {
  padding-right: 32px;
}

.employee-edit-form__section + .employee-edit-form__section {
  padding-left: 32px;
  border-left: 1px solid #e2e8f0;
}

@media (max-width: 860px) {
  .employee-edit-form__columns {
    grid-template-columns: 1fr;
    gap: 0;
  }

  .employee-edit-form__section:first-child {
    padding-right: 0;
    padding-bottom: 24px;
  }

  .employee-edit-form__section + .employee-edit-form__section {
    padding-left: 0;
    padding-top: 24px;
    border-left: none;
    border-top: 1px solid #e2e8f0;
  }
}

.employee-edit-form__section-title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  text-align: center;
}

.employee-edit-form__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.employee-edit-form__label {
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
}

.employee-edit-form__required {
  color: #dc2626;
  font-size: 12px;
  line-height: 1;
  font-weight: 600;
}

.employee-edit-form__control {
  width: 100%;
}

.employee-edit-form__error {
  margin: 0;
  width: 100%;
  max-width: 420px;
  margin-inline: auto;
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
