<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NDatePicker, NInput, NSelect } from 'naive-ui'
import AppModalButton from '@/components/common/AppModalButton.vue'
import { birthDateFromTimestamp, UsersApiError } from '@/api/users'
import { useEmployees } from '@/composables/useEmployees'
import type { EmployeeRole } from '@/types/employee'
import { isPhoneFilled, normalizePhone, PHONE_PREFIX } from '@/utils/phone'

const router = useRouter()
const { addEmployee } = useEmployees()
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
  phone: PHONE_PREFIX,
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
    employeeForm.password.trim().length > 0 &&
    employeeForm.position.trim().length > 0 &&
    employeeForm.role !== null,
)

function resetForm() {
  employeeForm.firstName = ''
  employeeForm.lastName = ''
  employeeForm.patronymic = ''
  employeeForm.birthDate = null
  employeeForm.phone = PHONE_PREFIX
  employeeForm.password = ''
  employeeForm.position = ''
  employeeForm.role = null
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
    await addEmployee({
      firstName: employeeForm.firstName.trim(),
      lastName: employeeForm.lastName.trim(),
      patronymic: employeeForm.patronymic.trim(),
      birthDate,
      phone: employeeForm.phone.trim(),
      password: employeeForm.password,
      position: employeeForm.position.trim(),
      role: employeeForm.role,
    })
    resetForm()
    await router.push({ name: 'employees-list' })
  } catch (error) {
    if (error instanceof UsersApiError) {
      errorMessage.value = error.message
    } else if (error instanceof Error) {
      errorMessage.value = error.message
    } else {
      errorMessage.value = 'Не удалось добавить сотрудника'
    }
  } finally {
    isSubmitting.value = false
  }
}

defineExpose({ resetForm })
</script>

<template>
  <form class="add-employee-form" @submit.prevent="handleSubmit">
    <div class="add-employee-form__columns">
      <section class="add-employee-form__section">
        <h3 class="add-employee-form__section-title">Личные данные</h3>

        <label class="add-employee-form__field">
          <span class="add-employee-form__label">
            Фамилия
            <span class="add-employee-form__required" aria-hidden="true">*</span>
          </span>
          <NInput
            v-model:value="employeeForm.lastName"
            class="add-employee-form__control"
            :theme-overrides="fieldInputTheme"
            placeholder="Иванов"
            autocomplete="off"
          />
        </label>

        <label class="add-employee-form__field">
          <span class="add-employee-form__label">
            Имя
            <span class="add-employee-form__required" aria-hidden="true">*</span>
          </span>
          <NInput
            v-model:value="employeeForm.firstName"
            class="add-employee-form__control"
            :theme-overrides="fieldInputTheme"
            placeholder="Иван"
            autocomplete="off"
          />
        </label>

        <label class="add-employee-form__field">
          <span class="add-employee-form__label">
            Отчество
            <span class="add-employee-form__required" aria-hidden="true">*</span>
          </span>
          <NInput
            v-model:value="employeeForm.patronymic"
            class="add-employee-form__control"
            :theme-overrides="fieldInputTheme"
            placeholder="Иванович"
            autocomplete="off"
          />
        </label>

        <label class="add-employee-form__field">
          <span class="add-employee-form__label">
            Дата рождения
            <span class="add-employee-form__required" aria-hidden="true">*</span>
          </span>
          <NDatePicker
            v-model:value="employeeForm.birthDate"
            class="add-employee-form__control"
            :theme-overrides="fieldDatePickerTheme"
            type="date"
            clearable
            format="dd.MM.yyyy"
            placeholder="Выберите дату"
          />
        </label>
      </section>

      <section class="add-employee-form__section">
        <h3 class="add-employee-form__section-title">Доступ в систему</h3>

        <label class="add-employee-form__field">
          <span class="add-employee-form__label">
            Номер телефона
            <span class="add-employee-form__required" aria-hidden="true">*</span>
          </span>
          <NInput
            v-model:value="employeeForm.phone"
            class="add-employee-form__control"
            :theme-overrides="fieldInputTheme"
            placeholder="+79001234567"
            :maxlength="12"
            @update:value="handlePhoneInput"
          />
        </label>

        <label class="add-employee-form__field">
          <span class="add-employee-form__label">
            Пароль для входа
            <span class="add-employee-form__required" aria-hidden="true">*</span>
          </span>
          <NInput
            v-model:value="employeeForm.password"
            class="add-employee-form__control"
            :theme-overrides="fieldInputTheme"
            type="password"
            show-password-on="click"
            placeholder="Введите пароль"
            autocomplete="new-password"
          />
        </label>

        <label class="add-employee-form__field">
          <span class="add-employee-form__label">
            Должность
            <span class="add-employee-form__required" aria-hidden="true">*</span>
          </span>
          <NInput
            v-model:value="employeeForm.position"
            class="add-employee-form__control"
            :theme-overrides="fieldInputTheme"
            placeholder="Например, менеджер по продажам"
            autocomplete="off"
          />
        </label>

        <label class="add-employee-form__field">
          <span class="add-employee-form__label">
            Роль
            <span class="add-employee-form__required" aria-hidden="true">*</span>
          </span>
          <NSelect
            v-model:value="employeeForm.role"
            class="add-employee-form__control"
            :theme-overrides="fieldSelectTheme"
            :options="roleOptions"
            placeholder="Выберите роль"
            clearable
          />
        </label>
      </section>
    </div>

    <div class="add-employee-form__actions">
      <p v-if="errorMessage" class="add-employee-form__error" role="alert">
        {{ errorMessage }}
      </p>
      <AppModalButton :disabled="!canSubmit || isSubmitting" @click="handleSubmit">
        {{ isSubmitting ? 'Добавление…' : 'Добавить сотрудника' }}
      </AppModalButton>
    </div>
  </form>
</template>

<style scoped>
.add-employee-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
  width: 100%;
}

.add-employee-form__columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  align-items: start;
}

.add-employee-form__section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.add-employee-form__section:first-child {
  padding-right: 32px;
}

.add-employee-form__section + .add-employee-form__section {
  padding-left: 32px;
  border-left: 1px solid #e2e8f0;
}

@media (max-width: 860px) {
  .add-employee-form__columns {
    grid-template-columns: 1fr;
    gap: 0;
  }

  .add-employee-form__section:first-child {
    padding-right: 0;
    padding-bottom: 24px;
  }

  .add-employee-form__section + .add-employee-form__section {
    padding-left: 0;
    padding-top: 24px;
    border-left: none;
    border-top: 1px solid #e2e8f0;
  }
}

.add-employee-form__section-title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  text-align: center;
}

.add-employee-form__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.add-employee-form__label {
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
}

.add-employee-form__required {
  color: #dc2626;
  font-size: 12px;
  line-height: 1;
  font-weight: 600;
}

.add-employee-form__control {
  width: 100%;
}

.add-employee-form__actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.add-employee-form__error {
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
