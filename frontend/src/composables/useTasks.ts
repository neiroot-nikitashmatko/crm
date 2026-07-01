import { computed, ref } from 'vue'
import {
  completeTask as completeTaskRequest,
  completeTasksByLead as completeTasksByLeadRequest,
  createTask as createTaskRequest,
  fetchTasks,
  patchTask as patchTaskRequest,
} from '@/api/tasks'
import { deleteAttachment, uploadTaskAttachments } from '@/api/attachments'
import { createTaskComment } from '@/api/activities'
import { useAuth } from '@/composables/useAuth'
import { normalizeStoredActivity, normalizeStoredAttachment, type StoredActivity } from '@/types/attachment'
import type { NewTaskPayload, Task, TaskActivityEntry, TaskAttachment, UpdateTaskPayload } from '@/types/task'

const tasks = ref<Task[]>([])
const isLoaded = ref(false)
const isLoading = ref(false)

function normalizeTask(raw: any): Task {
  const dueAtValue = raw.dueAt ?? raw.due_at ?? null
  const createdAtValue = raw.createdAt ?? raw.created_at ?? Date.now()
  const leadId = raw.leadId ?? raw.lead_id
  const dealId = raw.dealId ?? raw.deal_id
  const createdBy = raw.createdBy ?? raw.created_by
  const parsedCreatedAt = new Date(createdAtValue)
  const dueAt =
    dueAtValue === null || dueAtValue === undefined || dueAtValue === ''
      ? null
      : typeof dueAtValue === 'number'
        ? dueAtValue
        : Number(dueAtValue)

  const normalizedActivities = Array.isArray(raw.activities)
    ? raw.activities.map((entry: unknown, index: number) => {
        const normalized = normalizeStoredActivity(entry, String(createdBy ?? 'Система'))
        return {
          id: normalized.id || `${raw.id}-activity-${index}`,
          type: normalized.type,
          author: normalized.author,
          text: normalized.text,
          createdAt: new Date(normalized.createdAt),
        } satisfies TaskActivityEntry
      })
    : []

  const normalizedAttachments = Array.isArray(raw.attachments)
    ? raw.attachments.map((attachment: unknown) => {
        const normalized = normalizeStoredAttachment(attachment, String(createdBy ?? 'Система'))
        return {
          id: normalized.id,
          name: normalized.name,
          size: normalized.size,
          mimeType: normalized.mimeType,
          uploadedBy: normalized.uploadedBy,
          uploadedAt: new Date(normalized.uploadedAt),
        } satisfies TaskAttachment
      })
    : []

  return {
    id: String(raw.id),
    title: String(raw.title ?? ''),
    text: String(raw.text ?? ''),
    dueAt: dueAt !== null && Number.isNaN(dueAt) ? null : dueAt,
    leadId: leadId ? String(leadId) : undefined,
    dealId: dealId ? String(dealId) : undefined,
    createdBy: String(createdBy ?? ''),
    createdAt: Number.isNaN(parsedCreatedAt.getTime()) ? new Date() : parsedCreatedAt,
    status: raw.status === 'completed' ? 'completed' : 'active',
    attachments: normalizedAttachments,
    activities: normalizedActivities,
  }
}

function toTaskActivityEntry(activity: StoredActivity): TaskActivityEntry {
  return {
    id: activity.id,
    type: activity.type,
    author: activity.author,
    text: activity.text,
    createdAt: new Date(activity.createdAt),
  }
}

function prependTaskActivity(taskId: string, activity: StoredActivity) {
  const entry = toTaskActivityEntry(activity)
  tasks.value = tasks.value.map((item) =>
    item.id === taskId ? { ...item, activities: [entry, ...item.activities] } : item,
  )
}

export function useTasks() {
  const { user } = useAuth()
  const hasLoadedOnce = computed(() => isLoaded.value)

  async function loadTasks(force = false) {
    if ((isLoaded.value && !force) || isLoading.value) return
    isLoading.value = true
    try {
      const items = await fetchTasks()
      tasks.value = items.map(normalizeTask)
      isLoaded.value = true
    } finally {
      isLoading.value = false
    }
  }

  function getLeadTasks(leadId: string) {
    return tasks.value.filter((item) => item.leadId === leadId)
  }

  function getDealTasks(dealId: string) {
    return tasks.value.filter((item) => item.dealId === dealId)
  }

  async function addTask(payload: NewTaskPayload) {
    const createdBy = payload.createdBy ?? user.value?.id
    if (!createdBy) {
      throw new Error('Пользователь не авторизован')
    }

    const created = normalizeTask(
      await createTaskRequest({
        title: payload.title,
        text: payload.text,
        dueAt: payload.dueAt,
        leadId: payload.leadId,
        dealId: payload.dealId,
        createdBy,
      }),
    )
    tasks.value.unshift(created)
    return created
  }

  async function updateTask(taskId: string, payload: UpdateTaskPayload) {
    const updated = normalizeTask(await patchTaskRequest(taskId, payload))
    tasks.value = tasks.value.map((item) => (item.id === taskId ? updated : item))
    return updated
  }

  async function completeTask(taskId: string) {
    const updated = normalizeTask(await completeTaskRequest(taskId))
    tasks.value = tasks.value.map((item) => (item.id === taskId ? updated : item))
    return updated
  }

  async function completeLeadTasks(leadId: string) {
    await completeTasksByLeadRequest(leadId)
    tasks.value = tasks.value.map((item) =>
      item.leadId === leadId ? { ...item, status: 'completed' } : item,
    )
  }

  async function addTaskAttachments(taskId: string, files: File[]): Promise<TaskAttachment[]> {
    const { items, activity } = await uploadTaskAttachments(taskId, files)
    const mapped: TaskAttachment[] = items.map((attachment) => ({
      id: attachment.id,
      name: attachment.name,
      size: attachment.size,
      mimeType: attachment.mimeType,
      uploadedBy: attachment.uploadedBy,
      uploadedAt: new Date(attachment.uploadedAt),
    }))
    tasks.value = tasks.value.map((item) =>
      item.id === taskId ? { ...item, attachments: [...mapped, ...item.attachments] } : item,
    )
    if (activity) {
      prependTaskActivity(taskId, activity)
    }
    return mapped
  }

  async function removeTaskAttachment(taskId: string, attachmentId: string): Promise<void> {
    const { activity } = await deleteAttachment(attachmentId)
    tasks.value = tasks.value.map((item) =>
      item.id === taskId
        ? { ...item, attachments: item.attachments.filter((attachment) => attachment.id !== attachmentId) }
        : item,
    )
    if (activity) {
      prependTaskActivity(taskId, activity)
    }
  }

  async function addTaskComment(taskId: string, text: string) {
    const activity = await createTaskComment(taskId, text)
    prependTaskActivity(taskId, activity)
    return toTaskActivityEntry(activity)
  }

  return {
    tasks,
    hasLoadedOnce,
    loadTasks,
    getLeadTasks,
    getDealTasks,
    addTask,
    updateTask,
    completeTask,
    completeLeadTasks,
    addTaskAttachments,
    removeTaskAttachment,
    addTaskComment,
  }
}
