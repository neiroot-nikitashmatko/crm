import { computed, ref } from 'vue'
import {
  AuthApiError,
  clearPersistedAuthSession,
  loginByPhone,
  persistAuthSession,
  type AuthUser,
} from '@/api/auth'
import { AUTH_USER_STORAGE_KEY, getAuthToken } from '@/api/session'

const user = ref<AuthUser | null>(null)

function loadStoredUser() {
  if (typeof window === 'undefined') return
  const token = getAuthToken()
  if (!token) {
    user.value = null
    return
  }
  try {
    const raw = sessionStorage.getItem(AUTH_USER_STORAGE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw) as Partial<AuthUser>
    if (!parsed.id || !parsed.phone || !parsed.role) return
    user.value = {
      id: parsed.id,
      phone: parsed.phone,
      role: parsed.role,
      firstName: parsed.firstName,
      lastName: parsed.lastName,
      patronymic: parsed.patronymic,
    }
  } catch {
    user.value = null
    clearPersistedAuthSession()
  }
}

loadStoredUser()

export function useAuth() {
  const isAuthenticated = computed(() => user.value !== null && Boolean(getAuthToken()))
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function login(phone: string, password: string): Promise<{ ok: boolean; message?: string }> {
    try {
      const result = await loginByPhone(phone, password)
      user.value = result.user
      persistAuthSession(result.user, result.token)
      return { ok: true }
    } catch (error) {
      if (error instanceof AuthApiError) {
        if (error.status === 401) {
          return { ok: false, message: 'Неверный телефон или пароль' }
        }
        return { ok: false, message: error.message }
      }
      return { ok: false, message: 'Ошибка сети. Попробуйте снова' }
    }
  }

  function logout() {
    user.value = null
    clearPersistedAuthSession()
  }

  return {
    user,
    isAuthenticated,
    isAdmin,
    login,
    logout,
  }
}
