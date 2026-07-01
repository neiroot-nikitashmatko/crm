import type { TaskKanbanColumnId } from '@/types/task'

export interface TaskKanbanColumn {
  id: TaskKanbanColumnId
  title: string
  style: {
    headerBg: string
    headerBorder: string
    countBg: string
    countColor: string
  }
}

export const TASK_KANBAN_COLUMNS: TaskKanbanColumn[] = [
  {
    id: 'overdue',
    title: 'Просрочено',
    style: {
      headerBg: '#fef2f2',
      headerBorder: '#fecaca',
      countBg: 'rgba(220,38,38,0.12)',
      countColor: '#b91c1c',
    },
  },
  {
    id: 'today',
    title: 'Сегодня',
    style: {
      headerBg: '#eff6ff',
      headerBorder: '#bfdbfe',
      countBg: 'rgba(37,99,235,0.12)',
      countColor: '#1d4ed8',
    },
  },
  {
    id: 'tomorrow',
    title: 'Завтра',
    style: {
      headerBg: '#e6f4f1',
      headerBorder: '#b8e0d8',
      countBg: 'rgba(40, 94, 82, 0.12)',
      countColor: '#285e52',
    },
  },
  {
    id: 'later',
    title: 'Послезавтра и далее',
    style: {
      headerBg: '#f5f3ff',
      headerBorder: '#ddd6fe',
      countBg: 'rgba(124,58,237,0.12)',
      countColor: '#6d28d9',
    },
  },
  {
    id: 'no-deadline',
    title: 'Без дедлайна',
    style: {
      headerBg: '#f8fafc',
      headerBorder: '#cbd5e1',
      countBg: 'rgba(71,85,105,0.12)',
      countColor: '#334155',
    },
  },
]
