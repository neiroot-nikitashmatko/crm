<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { CameraOutline, PencilOutline } from '@vicons/ionicons5'
import { NDatePicker, NIcon, NInput, NSelect } from 'naive-ui'
import AppBottomSheet from '@/components/common/AppBottomSheet.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import EmployeeAvatarCropModal from '@/components/employees/EmployeeAvatarCropModal.vue'
import { birthDateFromTimestamp, birthDateToTimestamp, fetchEmployeeAvatarBlob, UsersApiError } from '@/api/users'
import { useEmployees } from '@/composables/useEmployees'
import {
  EMPLOYEE_POSITION_OPTIONS,
  EMPLOYEE_ROLE_OPTIONS,
  isHeadEmployeePosition,
  normalizeEmployeePosition,
  roleForEmployeePosition,
} from '@/constants/employees'
import type { Employee, EmployeePosition, EmployeeRole } from '@/types/employee'
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
const avatarInput = ref<HTMLInputElement | null>(null)
const avatarFile = ref<File | null>(null)
const avatarRemoved = ref(false)
const avatarPreviewUrl = ref('')
const cropSourceUrl = ref('')
const isCropOpen = ref(false)
const cropOpenedFromExisting = ref(false)
let avatarLoadToken = 0

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
  phone: '',
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
    employeeForm.position !== null &&
    selectedRole.value !== null,
)

const avatarInitials = computed(() => {
  const last = employeeForm.lastName.trim().charAt(0)
  const first = employeeForm.firstName.trim().charAt(0)
  return `${last}${first}`.toUpperCase()
})

const cropImageUrl = computed(() => cropSourceUrl.value || avatarPreviewUrl.value)
const cropModalTitle = computed(() =>
  cropOpenedFromExisting.value ? 'Изменить фото профиля' : 'Выберите миниатюру',
)

function fillFormFromEmployee(employee: Employee) {
  employeeForm.firstName = employee.firstName
  employeeForm.lastName = employee.lastName
  employeeForm.patronymic = employee.patronymic
  employeeForm.birthDate = birthDateToTimestamp(employee.birthDate)
  employeeForm.phone = employee.phone
  employeeForm.password = ''
  employeeForm.position = normalizeEmployeePosition(employee.position)
  employeeForm.role = roleForEmployeePosition(employeeForm.position, employee.role)
  errorMessage.value = ''
  void loadExistingAvatar(employee)
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

function resetAvatarState() {
  avatarLoadToken += 1
  revokeAvatarPreview()
  revokeCropSource()
  avatarFile.value = null
  avatarRemoved.value = false
  cropOpenedFromExisting.value = false
  isCropOpen.value = false
  if (avatarInput.value) avatarInput.value.value = ''
}

async function loadExistingAvatar(employee: Employee) {
  const token = ++avatarLoadToken
  revokeAvatarPreview()
  revokeCropSource()
  avatarFile.value = null
  avatarRemoved.value = false
  cropOpenedFromExisting.value = false
  isCropOpen.value = false
  if (avatarInput.value) avatarInput.value.value = ''
  if (!employee.hasAvatar) return

  try {
    const blob = await fetchEmployeeAvatarBlob(employee.id)
    if (token !== avatarLoadToken) return
    avatarPreviewUrl.value = URL.createObjectURL(blob)
  } catch {
    if (token !== avatarLoadToken) return
  }
}

function triggerAvatarPick() {
  avatarInput.value?.click()
}

function handleAvatarButtonClick() {
  if (avatarPreviewUrl.value) {
    cropOpenedFromExisting.value = true
    isCropOpen.value = true
    return
  }
  triggerAvatarPick()
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
  if (!isCropOpen.value) {
    cropOpenedFromExisting.value = false
  }
  revokeCropSource()
  cropSourceUrl.value = URL.createObjectURL(file)
  isCropOpen.value = true
  input.value = ''
}

function handleCropConfirm(file: File) {
  const previousPreview = avatarPreviewUrl.value
  avatarFile.value = file
  avatarRemoved.value = false
  avatarPreviewUrl.value = URL.createObjectURL(file)
  cropOpenedFromExisting.value = false

  if (previousPreview && previousPreview !== avatarPreviewUrl.value) {
    window.setTimeout(() => URL.revokeObjectURL(previousPreview), 400)
  }
}

function handleCropReplace() {
  triggerAvatarPick()
}

function handleCropClose() {
  cropOpenedFromExisting.value = false
  const url = cropSourceUrl.value
  if (!url) return
  window.setTimeout(() => {
    if (cropSourceUrl.value === url) revokeCropSource()
  }, 400)
}

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
    const updated = await editEmployee(props.employee.id, {
      firstName: employeeForm.firstName.trim(),
      lastName: employeeForm.lastName.trim(),
      patronymic: employeeForm.patronymic.trim(),
      birthDate,
      phone: employeeForm.phone.trim(),
      password: employeeForm.password,
      position: employeeForm.position,
      role: selectedRole.value,
    }, {
      file: avatarFile.value,
      remove: avatarRemoved.value && !avatarFile.value,
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
      return
    }
    resetAvatarState()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  resetAvatarState()
})
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
      <div class="employee-edit-form__avatar">
        <input
          ref="avatarInput"
          type="file"
          class="employee-edit-form__avatar-input"
          accept="image/jpeg,image/png,image/webp"
          @change="handleAvatarChange"
        />
        <button
          type="button"
          class="employee-edit-form__avatar-button"
          :title="avatarPreviewUrl ? 'Изменить фото профиля' : 'Загрузить фото'"
          :aria-label="avatarPreviewUrl ? 'Изменить фото профиля сотрудника' : 'Загрузить фото сотрудника'"
          @click="handleAvatarButtonClick"
        >
          <img
            v-if="avatarPreviewUrl"
            class="employee-edit-form__avatar-image"
            :src="avatarPreviewUrl"
            alt=""
          />
          <span v-else-if="avatarInitials" class="employee-edit-form__avatar-fallback">
            {{ avatarInitials }}
          </span>
          <NIcon v-else class="employee-edit-form__avatar-icon" :size="20" :component="CameraOutline" />
        </button>
        <button
          v-if="avatarPreviewUrl"
          type="button"
          class="employee-edit-form__avatar-edit"
          title="Изменить фото профиля"
          aria-label="Изменить фото профиля"
          @click="handleAvatarButtonClick"
        >
          <NIcon :size="14" :component="PencilOutline" />
        </button>
        <p class="employee-edit-form__avatar-hint">
          {{ avatarPreviewUrl ? 'Нажмите, чтобы изменить фото профиля' : 'Нажмите, чтобы загрузить фото' }}
        </p>
      </div>

      <EmployeeAvatarCropModal
        v-model:show="isCropOpen"
        :image-url="cropImageUrl"
        :title="cropModalTitle"
        can-replace
        @confirm="handleCropConfirm"
        @replace="handleCropReplace"
        @close="handleCropClose"
      />

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
            <NSelect
              v-model:value="employeeForm.position"
              class="employee-edit-form__control"
              :theme-overrides="fieldSelectTheme"
              :options="EMPLOYEE_POSITION_OPTIONS"
              placeholder="Выберите должность"
              @update:value="handlePositionUpdate"
            />
          </label>

          <p v-if="isHeadPosition" class="employee-edit-form__access-hint">
            Руководителю доступны все разделы системы
          </p>

          <label v-else class="employee-edit-form__field">
            <span class="employee-edit-form__label">
              Роль
              <span class="employee-edit-form__required" aria-hidden="true">*</span>
            </span>
            <NSelect
              v-model:value="employeeForm.role"
              class="employee-edit-form__control"
              :theme-overrides="fieldSelectTheme"
              :options="EMPLOYEE_ROLE_OPTIONS"
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
  gap: 16px;
  width: 100%;
  max-width: 880px;
  margin: 0 auto;
}

.employee-edit-form__avatar {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.employee-edit-form__avatar-input {
  display: none;
}

.employee-edit-form__avatar-button {
  position: relative;
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

.employee-edit-form__avatar-button:hover {
  border-color: #cbd5e1;
  background: #eef2f6;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.08);
}

.employee-edit-form__avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.employee-edit-form__avatar-fallback {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #4a5568;
  line-height: 1;
}

.employee-edit-form__avatar-icon {
  color: #64748b;
}

.employee-edit-form__avatar-edit {
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

.employee-edit-form__avatar-edit:hover {
  border-color: #cbd5e1;
  color: #1a202c;
}

.employee-edit-form__avatar-hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.35;
  color: #718096;
  text-align: center;
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

.employee-edit-form__access-hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.4;
  color: #1f883d;
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
