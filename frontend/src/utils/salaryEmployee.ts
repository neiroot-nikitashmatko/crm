import type { Employee } from '@/types/employee'

export function formatSalaryEmployeeName(employee: Pick<Employee, 'lastName' | 'firstName'>) {
  return [employee.lastName, employee.firstName]
    .map((part) => part.trim())
    .filter(Boolean)
    .join(' ')
}

export function isSalaryMasterEmployee(employee: Pick<Employee, 'position' | 'isActive'>) {
  return employee.isActive && employee.position.trim().toLowerCase() === 'мастер'
}
