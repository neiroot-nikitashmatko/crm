import { ref } from 'vue'
import {
  createSalaryEntry,
  deleteSalaryEntry,
  fetchSalaryEntries,
  updateSalaryEntry,
} from '@/api/salaryEntries'
import type { CreateSalaryReportEntryInput, SalaryReportEntry } from '@/types/salaryReport'

const reportEntries = ref<SalaryReportEntry[]>([])
const isLoaded = ref(false)
const isLoading = ref(false)

export function useSalaryReport() {
  async function loadReportEntries(force = false) {
    if (isLoading.value) return
    if (isLoaded.value && !force) return

    isLoading.value = true
    try {
      reportEntries.value = await fetchSalaryEntries()
      isLoaded.value = true
    } finally {
      isLoading.value = false
    }
  }

  async function addReportEntry(input: CreateSalaryReportEntryInput): Promise<SalaryReportEntry> {
    const created = await createSalaryEntry(input)
    reportEntries.value = [created, ...reportEntries.value]
    isLoaded.value = true
    return created
  }

  async function updateReportEntry(
    entryId: string,
    input: CreateSalaryReportEntryInput,
  ): Promise<SalaryReportEntry> {
    const updated = await updateSalaryEntry(entryId, input)
    reportEntries.value = reportEntries.value.map((row) => (row.id === updated.id ? updated : row))
    return updated
  }

  async function removeReportEntry(entryId: string): Promise<void> {
    await deleteSalaryEntry(entryId)
    reportEntries.value = reportEntries.value.filter((row) => row.id !== entryId)
  }

  return {
    reportEntries,
    isLoaded,
    isLoading,
    loadReportEntries,
    addReportEntry,
    updateReportEntry,
    removeReportEntry,
  }
}
