import type { EmployeePosition, EmployeeRole } from '@/types/employee'

export const EMPLOYEE_POSITION_OPTIONS: Array<{ label: EmployeePosition; value: EmployeePosition }> = [
  { label: 'Руководитель', value: 'Руководитель' },
  { label: 'Менеджер', value: 'Менеджер' },
  { label: 'Мастер', value: 'Мастер' },
]

export const EMPLOYEE_ROLE_LABELS: Record<EmployeeRole, string> = {
  admin: 'Администратор',
  manager: 'Пользователь',
}

export const EMPLOYEE_ROLE_OPTIONS: Array<{ label: string; value: EmployeeRole }> = [
  { label: 'Администратор', value: 'admin' },
  { label: 'Пользователь', value: 'manager' },
]

export function isHeadEmployeePosition(position?: string | null): boolean {
  return position?.trim().toLocaleLowerCase('ru-RU') === 'руководитель'
}

export function roleForEmployeePosition(
  position: EmployeePosition | null,
  role: EmployeeRole | null,
): EmployeeRole | null {
  if (isHeadEmployeePosition(position)) return 'admin'
  return role
}

export function normalizeEmployeePosition(position: string): EmployeePosition | null {
  const normalized = position.trim().toLocaleLowerCase('ru-RU')
  if (normalized.includes('руководитель')) return 'Руководитель'
  if (normalized.includes('мастер')) return 'Мастер'
  if (normalized.includes('менеджер')) return 'Менеджер'
  return null
}

export function getEmployeeRoleLabel(role: string): string {
  if (role === 'admin') return EMPLOYEE_ROLE_LABELS.admin
  if (role === 'manager') return EMPLOYEE_ROLE_LABELS.manager
  return role
}

export function getEmployeeInitials(employee: { firstName: string; lastName: string }): string {
  const last = employee.lastName.trim().charAt(0)
  const first = employee.firstName.trim().charAt(0)
  return `${last}${first}`.toUpperCase()
}

export function getAuthorInitials(author: string): string {
  const parts = author.trim().split(/\s+/).filter(Boolean)
  if (parts.length >= 2) {
    return `${parts[0].charAt(0)}${parts[1].charAt(0)}`.toUpperCase()
  }
  if (parts.length === 1) return parts[0].charAt(0).toUpperCase()
  return ''
}

export function getAuthorShortName(author: string): string {
  const parts = author.trim().split(/\s+/).filter(Boolean)
  if (parts.length >= 2) return `${parts[0]} ${parts[1]}`
  return author.trim()
}
