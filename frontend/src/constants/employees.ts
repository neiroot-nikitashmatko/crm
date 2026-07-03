import type { EmployeePosition, EmployeeRole } from '@/types/employee'

export const EMPLOYEE_POSITION_OPTIONS: Array<{ label: EmployeePosition; value: EmployeePosition }> = [
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

export function normalizeEmployeePosition(position: string): EmployeePosition | null {
  const normalized = position.trim().toLocaleLowerCase('ru-RU')
  if (normalized.includes('мастер')) return 'Мастер'
  if (normalized.includes('менеджер')) return 'Менеджер'
  return null
}

export function getEmployeeRoleLabel(role: string): string {
  if (role === 'admin') return EMPLOYEE_ROLE_LABELS.admin
  if (role === 'manager') return EMPLOYEE_ROLE_LABELS.manager
  return role
}
