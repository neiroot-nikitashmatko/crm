import type { DealKanbanColumnId } from '@/types/deal'

export interface DealKanbanColumn {
  id: DealKanbanColumnId
  title: string
  style: {
    headerBg: string
    headerBorder: string
    countBg: string
    countColor: string
  }
}

export const DEAL_KANBAN_COLUMNS: DealKanbanColumn[] = [
  {
    id: 'production',
    title: 'Производство',
    style: {
      headerBg: '#eff6ff',
      headerBorder: '#bfdbfe',
      countBg: 'rgba(37,99,235,0.12)',
      countColor: '#1d4ed8',
    },
  },
  {
    id: 'pickup',
    title: 'Самовывоз',
    style: {
      headerBg: '#ecfdf5',
      headerBorder: '#a7f3d0',
      countBg: 'rgba(5,150,105,0.12)',
      countColor: '#047857',
    },
  },
  {
    id: 'delivery',
    title: 'Доставка',
    style: {
      headerBg: '#f5f3ff',
      headerBorder: '#ddd6fe',
      countBg: 'rgba(124,58,237,0.12)',
      countColor: '#6d28d9',
    },
  },
  {
    id: 'closed',
    title: 'Закрытые сделки',
    style: {
      headerBg: '#f0fdf4',
      headerBorder: '#86efac',
      countBg: 'rgba(22,163,74,0.12)',
      countColor: '#15803d',
    },
  },
  {
    id: 'failed',
    title: 'Проваленные сделки',
    style: {
      headerBg: '#fef2f2',
      headerBorder: '#fecaca',
      countBg: 'rgba(220,38,38,0.12)',
      countColor: '#b91c1c',
    },
  },
]
