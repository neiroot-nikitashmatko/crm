export type EmployeeRole = 'admin' | 'manager'

export interface Employee {
  id: string
  firstName: string
  lastName: string
  patronymic: string
  phone: string
  role: EmployeeRole
  position: string
  birthDate?: string | null
  isActive: boolean
  createdAt: number
  updatedAt: number
}

export interface CreateEmployeeInput {
  firstName: string
  lastName: string
  patronymic: string
  birthDate: string
  phone: string
  password: string
  position: string
  role: EmployeeRole
}

export interface UpdateEmployeeInput {
  firstName: string
  lastName: string
  patronymic: string
  birthDate: string
  phone: string
  password: string
  position: string
  role: EmployeeRole
}
