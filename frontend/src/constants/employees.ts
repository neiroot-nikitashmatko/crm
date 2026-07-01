import type { EmployeeRole } from '@/types/employee'

export const EMPLOYEE_ROLE_LABELS: Record<EmployeeRole, string> = {
  admin: 'Администратор',
  manager: 'Пользователь',
}

export function getEmployeeRoleLabel(role: string): string {
  if (role === 'admin') return EMPLOYEE_ROLE_LABELS.admin
  if (role === 'manager') return EMPLOYEE_ROLE_LABELS.manager
  return role
}
