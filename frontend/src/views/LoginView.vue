<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { EyeOffOutline, EyeOutline } from '@vicons/ionicons5'
import { NButton, NIcon, NInput } from 'naive-ui'
import { useAuth } from '@/composables/useAuth'
import { isPhoneFilled, normalizePhone, PHONE_PREFIX } from '@/utils/phone'
import logoUrl from '@/assets/logo.png'

const router = useRouter()
const route = useRoute()
const { login } = useAuth()

const phone = ref(PHONE_PREFIX)
const password = ref('')
const isSubmitting = ref(false)
const errorMessage = ref('')
const showPassword = ref(false)

const canSubmit = computed(() => isPhoneFilled(phone.value) && password.value.trim().length > 0)

function handlePhoneInput(value: string) {
  phone.value = normalizePhone(value)
}

async function handleSubmit() {
  if (!canSubmit.value) return
  isSubmitting.value = true
  errorMessage.value = ''
  try {
    const result = await login(phone.value.trim(), password.value)
    if (!result.ok) {
      errorMessage.value = result.message ?? 'Не удалось войти'
      return
    }
    const redirectTarget = typeof route.query.redirect === 'string' ? route.query.redirect : '/leads'
    await router.replace(redirectTarget)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <main class="login-page">
    <section class="login-card">
      <header class="login-card__header">
        <h1 class="login-card__sr-title">Вход в CRM</h1>
        <img class="login-card__logo" :src="logoUrl" alt="NEIROOT" />
      </header>

      <form class="login-card__form" @submit.prevent="handleSubmit">
        <label class="login-card__field">
          <span class="login-card__label">Телефон</span>
          <NInput
            v-model:value="phone"
            placeholder="+79001234567"
            :maxlength="12"
            size="large"
            @update:value="handlePhoneInput"
          />
        </label>

        <label class="login-card__field">
          <span class="login-card__label">Пароль</span>
          <NInput
            v-model:value="password"
            :type="showPassword ? 'text' : 'password'"
            placeholder="Введите пароль"
            size="large"
          >
            <template #suffix>
              <button
                type="button"
                class="login-card__toggle"
                aria-label="Показать или скрыть пароль"
                @click="showPassword = !showPassword"
              >
                <NIcon :component="showPassword ? EyeOffOutline : EyeOutline" :size="18" />
              </button>
            </template>
          </NInput>
        </label>

        <p v-if="errorMessage" class="login-card__error" role="alert">{{ errorMessage }}</p>

        <NButton
          type="primary"
          color="#1f883d"
          block
          size="large"
          attr-type="submit"
          :disabled="!canSubmit || isSubmitting"
          :loading="isSubmitting"
        >
          Войти
        </NButton>
      </form>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 16px;
  background:
    radial-gradient(ellipse 80% 60% at 50% -10%, rgba(31, 136, 61, 0.1), transparent 55%),
    linear-gradient(180deg, #f8fafc 0%, #eef2f7 100%);
}

.login-card {
  width: min(360px, 100%);
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  box-shadow:
    0 1px 2px rgba(15, 23, 42, 0.04),
    0 12px 32px rgba(15, 23, 42, 0.08);
  overflow: hidden;
}

.login-card__header {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 28px 8px;
}

.login-card__sr-title {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.login-card__logo {
  display: block;
  height: 30px;
  width: auto;
  max-width: 200px;
  object-fit: contain;
  user-select: none;
  pointer-events: none;
}

.login-card__form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px 28px 28px;
}

.login-card__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.login-card__label {
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
}

.login-card__toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  padding: 0;
  color: #64748b;
  cursor: pointer;
  border-radius: 6px;
  transition: color 0.15s ease, background-color 0.15s ease;
}

.login-card__toggle:hover {
  color: #1a202c;
  background: #f1f5f9;
}

.login-card__error {
  margin: 0;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #b91c1c;
  font-size: 13px;
  line-height: 1.4;
}

.login-card__form :deep(.n-button) {
  margin-top: 4px;
  font-weight: 600;
  border-radius: 10px;
}

.login-card__form :deep(.n-input) {
  border-radius: 10px;
}
</style>
