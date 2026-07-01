import { computed, ref } from 'vue'
import { createEmployee, deleteEmployee, fetchEmployeeById, fetchEmployees, updateEmployee } from '@/api/users'
import type { CreateEmployeeInput, Employee, UpdateEmployeeInput } from '@/types/employee'

const employees = ref<Employee[]>([])
const isLoaded = ref(false)
const isLoading = ref(false)

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
      return employee
    } catch {
      return null
    }
  }

  async function addEmployee(input: CreateEmployeeInput): Promise<Employee> {
    const created = await createEmployee(input)
    employees.value = [...employees.value, created]
    isLoaded.value = true
    return created
  }

  async function removeEmployee(employeeId: string): Promise<void> {
    await deleteEmployee(employeeId)
    employees.value = employees.value.filter((item) => item.id !== employeeId)
  }

  async function editEmployee(employeeId: string, input: UpdateEmployeeInput): Promise<Employee> {
    const updated = await updateEmployee(employeeId, input)
    employees.value = employees.value.map((item) => (item.id === updated.id ? updated : item))
    isLoaded.value = true
    return updated
  }

  return {
    employees: sortedEmployees,
    isLoaded,
    isLoading,
    loadEmployees,
    getEmployee,
    addEmployee,
    removeEmployee,
    editEmployee,
  }
}
