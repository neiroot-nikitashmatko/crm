import { computed, ref } from 'vue'
import {
  createEmployee,
  deleteEmployee,
  deleteEmployeeAvatar,
  fetchEmployeeAvatarBlob,
  fetchEmployeeById,
  fetchEmployees,
  updateEmployee,
  uploadEmployeeAvatar,
} from '@/api/users'
import type { CreateEmployeeInput, Employee, UpdateEmployeeInput } from '@/types/employee'

const employees = ref<Employee[]>([])
const avatarUrls = ref<Record<string, string>>({})
const isLoaded = ref(false)
const isLoading = ref(false)

function revokeAvatarUrl(employeeId: string) {
  const url = avatarUrls.value[employeeId]
  if (!url) return
  URL.revokeObjectURL(url)
  const next = { ...avatarUrls.value }
  delete next[employeeId]
  avatarUrls.value = next
}

function setAvatarUrl(employeeId: string, url: string) {
  avatarMissing.delete(employeeId)
  revokeAvatarUrl(employeeId)
  avatarUrls.value = { ...avatarUrls.value, [employeeId]: url }
}

function setAvatarFromFile(employeeId: string, file: File) {
  setAvatarUrl(employeeId, URL.createObjectURL(file))
}

async function loadAvatar(employee: Employee) {
  if (!employee.hasAvatar) {
    revokeAvatarUrl(employee.id)
    return
  }

  try {
    const blob = await fetchEmployeeAvatarBlob(employee.id)
    setAvatarUrl(employee.id, URL.createObjectURL(blob))
  } catch {
    revokeAvatarUrl(employee.id)
  }
}

async function syncAvatars(list: Employee[]) {
  const withAvatar = new Set(list.filter((item) => item.hasAvatar).map((item) => item.id))

  for (const employeeId of Object.keys(avatarUrls.value)) {
    if (!withAvatar.has(employeeId)) {
      revokeAvatarUrl(employeeId)
    }
  }

  await Promise.all(list.filter((item) => item.hasAvatar).map((item) => loadAvatar(item)))
}

const avatarPending = new Set<string>()
const avatarMissing = new Set<string>()

async function ensureAvatar(employeeId: string) {
  const id = employeeId.trim()
  if (!id || avatarUrls.value[id] || avatarMissing.has(id) || avatarPending.has(id)) return

  const known = employees.value.find((item) => item.id === id)
  if (known && !known.hasAvatar) {
    avatarMissing.add(id)
    return
  }

  avatarPending.add(id)
  try {
    const blob = await fetchEmployeeAvatarBlob(id)
    setAvatarUrl(id, URL.createObjectURL(blob))
  } catch {
    avatarMissing.add(id)
  } finally {
    avatarPending.delete(id)
  }
}

export function useEmployees() {
  const sortedEmployees = computed(() =>
    [...employees.value].sort((left, right) => {
      const byLastName = left.lastName.localeCompare(right.lastName, 'ru')
      if (byLastName !== 0) return byLastName
      const byFirstName = left.firstName.localeCompare(right.firstName, 'ru')
      if (byFirstName !== 0) return byFirstName
      return left.patronymic.localeCompare(right.patronymic, 'ru')
    }),
  )

  async function loadEmployees(force = false) {
    if (isLoading.value) return
    if (isLoaded.value && !force) return

    isLoading.value = true
    try {
      employees.value = await fetchEmployees()
      isLoaded.value = true
    } finally {
      isLoading.value = false
    }

    void syncAvatars(employees.value)
  }

  async function getEmployee(employeeId: string): Promise<Employee | null> {
    const cached = employees.value.find((item) => item.id === employeeId)
    if (cached) return cached

    try {
      const employee = await fetchEmployeeById(employeeId)
      const index = employees.value.findIndex((item) => item.id === employee.id)
      if (index >= 0) {
        employees.value[index] = employee
      } else {
        employees.value.push(employee)
      }
      void loadAvatar(employee)
      return employee
    } catch {
      return null
    }
  }

  async function addEmployee(input: CreateEmployeeInput, avatarFile?: File | null): Promise<Employee> {
    let created = await createEmployee(input)
    if (avatarFile) {
      created = await uploadEmployeeAvatar(created.id, avatarFile)
      setAvatarFromFile(created.id, avatarFile)
    }
    employees.value = [...employees.value, created]
    isLoaded.value = true
    return created
  }

  async function removeEmployee(employeeId: string): Promise<void> {
    await deleteEmployee(employeeId)
    revokeAvatarUrl(employeeId)
    employees.value = employees.value.filter((item) => item.id !== employeeId)
  }

  async function editEmployee(
    employeeId: string,
    input: UpdateEmployeeInput,
    avatar?: { file?: File | null; remove?: boolean },
  ): Promise<Employee> {
    let updated = await updateEmployee(employeeId, input)
    if (avatar?.file) {
      updated = await uploadEmployeeAvatar(updated.id, avatar.file)
      setAvatarFromFile(updated.id, avatar.file)
    } else if (avatar?.remove) {
      await deleteEmployeeAvatar(updated.id)
      updated = { ...updated, hasAvatar: false }
      revokeAvatarUrl(updated.id)
    }
    employees.value = employees.value.map((item) => (item.id === updated.id ? updated : item))
    isLoaded.value = true
    return updated
  }

  return {
    employees: sortedEmployees,
    avatarUrls,
    isLoaded,
    isLoading,
    loadEmployees,
    getEmployee,
    ensureAvatar,
    addEmployee,
    removeEmployee,
    editEmployee,
  }
}
