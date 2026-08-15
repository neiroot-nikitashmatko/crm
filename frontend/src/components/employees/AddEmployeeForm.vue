<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { CameraOutline, CloseOutline } from '@vicons/ionicons5'
import { NDatePicker, NIcon, NInput, NSelect } from 'naive-ui'
import AppModalButton from '@/components/common/AppModalButton.vue'
import EmployeeAvatarCropModal from '@/components/employees/EmployeeAvatarCropModal.vue'
import { birthDateFromTimestamp, UsersApiError } from '@/api/users'
import { useEmployees } from '@/composables/useEmployees'
import { EMPLOYEE_POSITION_OPTIONS, EMPLOYEE_ROLE_OPTIONS, isHeadEmployeePosition, roleForEmployeePosition } from '@/constants/employees'
import type { EmployeePosition, EmployeeRole } from '@/types/employee'
import { isPhoneFilled, normalizePhone, PHONE_PREFIX } from '@/utils/phone'

const router = useRouter()
const { addEmployee } = useEmployees()
const isSubmitting = ref(false)
const errorMessage = ref('')
const avatarInput = ref<HTMLInputElement | null>(null)
const avatarFile = ref<File | null>(null)
const avatarPreviewUrl = ref('')
const cropSourceUrl = ref('')
const isCropOpen = ref(false)

const AVATAR_MAX_BYTES = 2 * 1024 * 1024
const AVATAR_MIME_TYPES = ['image/jpeg', 'image/png', 'image/webp']

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
  position: null as EmployeePosition | null,
  role: null as EmployeeRole | null,
})

const isHeadPosition = computed(() => isHeadEmployeePosition(employeeForm.position))
const selectedRole = computed(() => roleForEmployeePosition(employeeForm.position, employeeForm.role))

const canSubmit = computed(
  () =>
    employeeForm.firstName.trim().length > 0 &&
    employeeForm.lastName.trim().length > 0 &&
    employeeForm.patronymic.trim().length > 0 &&
    employeeForm.birthDate !== null &&
    isPhoneFilled(employeeForm.phone) &&
    employeeForm.password.trim().length > 0 &&
    employeeForm.position !== null &&
    selectedRole.value !== null,
)

const avatarInitials = computed(() => {
  const last = employeeForm.lastName.trim().charAt(0)
  const first = employeeForm.firstName.trim().charAt(0)
  return `${last}${first}`.toUpperCase()
})

function resetForm() {
  employeeForm.firstName = ''
  employeeForm.lastName = ''
  employeeForm.patronymic = ''
  employeeForm.birthDate = null
  employeeForm.phone = PHONE_PREFIX
  employeeForm.password = ''
  employeeForm.position = null
  employeeForm.role = null
  clearAvatar()
}

function revokeAvatarPreview() {
  if (!avatarPreviewUrl.value) return
  URL.revokeObjectURL(avatarPreviewUrl.value)
  avatarPreviewUrl.value = ''
}

function revokeCropSource() {
  if (!cropSourceUrl.value) return
  URL.revokeObjectURL(cropSourceUrl.value)
  cropSourceUrl.value = ''
}

function clearAvatar() {
  revokeAvatarPreview()
  avatarFile.value = null
  if (avatarInput.value) avatarInput.value.value = ''
}

function triggerAvatarPick() {
  avatarInput.value?.click()
}

function handleAvatarChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  if (!AVATAR_MIME_TYPES.includes(file.type) || file.size > AVATAR_MAX_BYTES) {
    errorMessage.value = 'Нужно фото JPEG, PNG или WebP до 2 МБ'
    input.value = ''
    return
  }

  errorMessage.value = ''
  revokeCropSource()
  cropSourceUrl.value = URL.createObjectURL(file)
  isCropOpen.value = true
  input.value = ''
}

function handleCropConfirm(file: File) {
  revokeAvatarPreview()
  avatarFile.value = file
  avatarPreviewUrl.value = URL.createObjectURL(file)
}

function handleCropClose() {
  const url = cropSourceUrl.value
  if (!url) return
  window.setTimeout(() => {
    if (cropSourceUrl.value === url) revokeCropSource()
  }, 400)
}

onBeforeUnmount(() => {
  revokeAvatarPreview()
  revokeCropSource()
})

function handlePositionUpdate(value: EmployeePosition | null) {
  if (isHeadEmployeePosition(value)) {
    employeeForm.role = 'admin'
  }
}

function handlePhoneInput(value: string) {
  employeeForm.phone = normalizePhone(value)
}

function handleSubmit() {
  if (!canSubmit.value || isSubmitting.value || employeeForm.position === null || selectedRole.value === null) return

  void submitEmployee()
}

async function submitEmployee() {
  if (employeeForm.position === null || selectedRole.value === null) return

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
      position: employeeForm.position,
      role: selectedRole.value,
    }, avatarFile.value)
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
    <div class="add-employee-form__avatar">
      <input
        ref="avatarInput"
        type="file"
        class="add-employee-form__avatar-input"
        accept="image/jpeg,image/png,image/webp"
        @change="handleAvatarChange"
      />
      <button
        type="button"
        class="add-employee-form__avatar-button"
        title="Загрузить фото"
        aria-label="Загрузить фото сотрудника"
        @click="triggerAvatarPick"
      >
        <img
          v-if="avatarPreviewUrl"
          class="add-employee-form__avatar-image"
          :src="avatarPreviewUrl"
          alt=""
        />
        <span v-else-if="avatarInitials" class="add-employee-form__avatar-fallback">
          {{ avatarInitials }}
        </span>
        <NIcon v-else class="add-employee-form__avatar-icon" :size="20" :component="CameraOutline" />
      </button>
      <button
        v-if="avatarPreviewUrl"
        type="button"
        class="add-employee-form__avatar-remove"
        title="Убрать фото"
        aria-label="Убрать фото"
        @click="clearAvatar"
      >
        <NIcon :size="14" :component="CloseOutline" />
      </button>
      <p class="add-employee-form__avatar-hint">Загрузите фото сотрудника</p>
    </div>

    <EmployeeAvatarCropModal
      v-model:show="isCropOpen"
      :image-url="cropSourceUrl"
      @confirm="handleCropConfirm"
      @close="handleCropClose"
    />
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
          <NSelect
            v-model:value="employeeForm.position"
            class="add-employee-form__control"
            :theme-overrides="fieldSelectTheme"
            :options="EMPLOYEE_POSITION_OPTIONS"
            placeholder="Выберите должность"
            clearable
            @update:value="handlePositionUpdate"
          />
        </label>

        <p v-if="isHeadPosition" class="add-employee-form__access-hint">
          Руководителю доступны все разделы системы
        </p>

        <label v-else class="add-employee-form__field">
          <span class="add-employee-form__label">
            Роль
            <span class="add-employee-form__required" aria-hidden="true">*</span>
          </span>
          <NSelect
            v-model:value="employeeForm.role"
            class="add-employee-form__control"
            :theme-overrides="fieldSelectTheme"
            :options="EMPLOYEE_ROLE_OPTIONS"
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
  gap: 16px;
  width: 100%;
}

.add-employee-form__avatar {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.add-employee-form__avatar-input {
  display: none;
}

.add-employee-form__avatar-button {
  width: 64px;
  height: 64px;
  padding: 0;
  border: 1px solid #e2e8f0;
  border-radius: 50%;
  overflow: hidden;
  background: #f6f8fa;
  color: #64748b;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease,
    box-shadow 0.15s ease;
}

.add-employee-form__avatar-button:hover {
  border-color: #cbd5e1;
  background: #eef2f6;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.08);
}

.add-employee-form__avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.add-employee-form__avatar-fallback {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #4a5568;
  line-height: 1;
}

.add-employee-form__avatar-icon {
  color: #64748b;
}

.add-employee-form__avatar-remove {
  position: absolute;
  top: 0;
  right: calc(50% - 40px);
  width: 22px;
  height: 22px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #e2e8f0;
  border-radius: 50%;
  background: #ffffff;
  color: #64748b;
  cursor: pointer;
}

.add-employee-form__avatar-remove:hover {
  border-color: #cbd5e1;
  color: #1a202c;
}

.add-employee-form__avatar-hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.35;
  color: #718096;
  text-align: center;
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

.add-employee-form__access-hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.4;
  color: #1f883d;
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
