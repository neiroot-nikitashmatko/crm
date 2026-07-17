import { ApiError, requestJson } from '@/api/httpClient'
import { SALARY_SERVICE_OPTIONS } from '@/constants/salary'
import type { CreateSalaryReportEntryInput, SalaryReportEntry } from '@/types/salaryReport'

interface SalaryEntriesListResponse {
  items: Array<Record<string, unknown>>
}

interface SalaryEntryItemResponse {
  item: Record<string, unknown>
}

export class SalaryEntriesApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'SalaryEntriesApiError'
  }
}

function serviceLabelFor(service: string): string {
  const found = SALARY_SERVICE_OPTIONS.find((item) => item.value === service)
  return found?.label ?? service
}

function formatSalaryAmount(value: unknown): string {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return ''
  if (Number.isInteger(numeric)) return String(numeric)
  return String(numeric)
}

function normalizeSalaryEntry(raw: Record<string, unknown>): SalaryReportEntry {
  const service = String(raw.service ?? '')
  return {
    id: String(raw.id ?? ''),
    date: Number(raw.date ?? 0),
    dealId: String(raw.dealId ?? ''),
    dealNumberLabel: String(raw.dealNumberLabel ?? ''),
    service,
    serviceLabel: serviceLabelFor(service),
    salary: formatSalaryAmount(raw.salary),
    comment: String(raw.comment ?? ''),
    createdBy: String(raw.createdBy ?? ''),
  }
}

async function salaryEntriesRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new SalaryEntriesApiError(error.message, error.status)
    }
    throw error
  }
}

export async function fetchSalaryEntries(): Promise<SalaryReportEntry[]> {
  const payload = await salaryEntriesRequestJson<SalaryEntriesListResponse>('/api/v1/salary-entries', {
    method: 'GET',
  })
  return payload.items.map(normalizeSalaryEntry)
}

export async function createSalaryEntry(input: CreateSalaryReportEntryInput): Promise<SalaryReportEntry> {
  const response = await salaryEntriesRequestJson<SalaryEntryItemResponse>('/api/v1/salary-entries', {
    method: 'POST',
    body: JSON.stringify({
      date: input.date,
      dealId: input.dealId,
      service: input.service,
      salary: Number(input.salary),
      comment: input.comment,
      ...(input.employeeId ? { employeeId: input.employeeId } : {}),
    }),
  })
  return normalizeSalaryEntry(response.item)
}

export async function updateSalaryEntry(
  entryId: string,
  input: CreateSalaryReportEntryInput,
): Promise<SalaryReportEntry> {
  const response = await salaryEntriesRequestJson<SalaryEntryItemResponse>(
    `/api/v1/salary-entries/${entryId}`,
    {
      method: 'PATCH',
      body: JSON.stringify({
        date: input.date,
        dealId: input.dealId,
        service: input.service,
        salary: Number(input.salary),
        comment: input.comment,
      }),
    },
  )
  return normalizeSalaryEntry(response.item)
}

export async function deleteSalaryEntry(entryId: string): Promise<void> {
  await salaryEntriesRequestJson<{ ok: boolean }>(`/api/v1/salary-entries/${entryId}`, {
    method: 'DELETE',
  })
}
