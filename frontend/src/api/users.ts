import type { CreateEmployeeInput, Employee, EmployeeRole, UpdateEmployeeInput } from '@/types/employee'
import { ApiError, requestJson } from '@/api/httpClient'

interface UsersListResponse {
  items: Employee[]
}

interface UserItemResponse {
  item: Employee
}

export class UsersApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'UsersApiError'
  }
}

function normalizeEmployee(raw: Employee): Employee {
  return {
    id: String(raw.id),
    firstName: String(raw.firstName ?? ''),
    lastName: String(raw.lastName ?? ''),
    patronymic: String(raw.patronymic ?? ''),
    phone: String(raw.phone ?? ''),
    role: raw.role === 'admin' ? 'admin' : 'manager',
    position: String(raw.position ?? ''),
    birthDate: raw.birthDate ?? null,
    isActive: Boolean(raw.isActive ?? true),
    createdAt: Number(raw.createdAt ?? Date.now()),
    updatedAt: Number(raw.updatedAt ?? Date.now()),
  }
}

async function usersRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new UsersApiError(error.message, error.status)
    }
    throw error
  }
}

export async function fetchEmployees(): Promise<Employee[]> {
  const payload = await usersRequestJson<UsersListResponse>('/api/v1/users', {
    method: 'GET',
  })
  return payload.items.map(normalizeEmployee)
}

export async function fetchEmployeeById(employeeId: string): Promise<Employee> {
  const payload = await usersRequestJson<UserItemResponse>(`/api/v1/users/${employeeId}`, {
    method: 'GET',
  })
  return normalizeEmployee(payload.item)
}

export async function createEmployee(payload: CreateEmployeeInput): Promise<Employee> {
  const response = await usersRequestJson<UserItemResponse>('/api/v1/users', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return normalizeEmployee(response.item)
}

export async function deleteEmployee(employeeId: string): Promise<void> {
  await usersRequestJson<{ ok: boolean }>(`/api/v1/users/${employeeId}`, {
    method: 'DELETE',
  })
}

export async function updateEmployee(employeeId: string, payload: UpdateEmployeeInput): Promise<Employee> {
  const response = await usersRequestJson<UserItemResponse>(`/api/v1/users/${employeeId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
  return normalizeEmployee(response.item)
}

export function formatEmployeeBirthDate(birthDate?: string | null): string {
  if (!birthDate) return '—'
  const [year, month, day] = birthDate.split('-')
  if (!year || !month || !day) return birthDate
  return `${day}.${month}.${year}`
}

export function birthDateFromTimestamp(timestamp: number | null): string | null {
  if (timestamp === null) return null
  const date = new Date(timestamp)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function birthDateToTimestamp(birthDate?: string | null): number | null {
  if (!birthDate) return null
  const [year, month, day] = birthDate.split('-').map((part) => Number(part))
  if (!year || !month || !day) return null
  return new Date(year, month - 1, day).getTime()
}

export function isEmployeeRole(value: string | null): value is EmployeeRole {
  return value === 'admin' || value === 'manager'
}
