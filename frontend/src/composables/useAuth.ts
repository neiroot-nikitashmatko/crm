import { computed, ref } from 'vue'
import {
  AuthApiError,
  clearPersistedAuthSession,
  loginByPhone,
  persistAuthSession,
  type AuthUser,
} from '@/api/auth'
import { getAuthToken, getAuthUserRaw } from '@/api/session'

const user = ref<AuthUser | null>(null)

const managerSections = new Set([
  'leads',
  'deals',
  'tasks',
  'products-catalog',
  'production-calendar',
])

const masterSections = new Set([
  'leads',
  'deals',
  'production-calendar',
  'salary',
])

function normalizePosition(position?: string): 'manager' | 'master' | '' {
  const normalized = position?.trim().toLocaleLowerCase('ru-RU') ?? ''
  if (normalized.includes('мастер')) return 'master'
  if (normalized.includes('менеджер')) return 'manager'
  return ''
}

function loadStoredUser() {
  if (typeof window === 'undefined') return
  const token = getAuthToken()
  if (!token) {
    user.value = null
    return
  }
  try {
    const raw = getAuthUserRaw()
    if (!raw) return
    const parsed = JSON.parse(raw) as Partial<AuthUser>
    if (!parsed.id || !parsed.phone || !parsed.role) return
    user.value = {
      id: parsed.id,
      phone: parsed.phone,
      role: parsed.role,
      position: parsed.position,
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
  const rawPosition = computed(() => user.value?.position?.trim() ?? '')
  const position = computed(() => normalizePosition(user.value?.position))

  function canAccessSection(sectionName: string): boolean {
    if (isAdmin.value) return true
    if (position.value === 'manager') return managerSections.has(sectionName)
    if (position.value === 'master') return masterSections.has(sectionName)

    // Existing sessions created before position was returned by the API should not lock users out.
    if (rawPosition.value === '') return managerSections.has(sectionName)

    return masterSections.has(sectionName)
  }

  function getDefaultRouteName(): string {
    if (canAccessSection('leads')) return 'leads'
    return 'login'
  }

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
    canAccessSection,
    getDefaultRouteName,
    login,
    logout,
  }
}
