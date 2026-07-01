export interface LeadKanbanColumnStyle {
  headerBg: string
  headerBorder: string
  countBg: string
  countColor: string
}

export interface LeadKanbanColumn {
  id: string
  title: string
  style: LeadKanbanColumnStyle
  showAddLeadButton?: boolean
}

export const LEAD_KANBAN_COLUMNS: LeadKanbanColumn[] = [
  {
    id: 'new',
    title: 'Новый лид',
    showAddLeadButton: true,
    style: {
      headerBg: '#e8f1fc',
      headerBorder: '#c5d9f0',
      countBg: 'rgba(44, 82, 130, 0.12)',
      countColor: '#2c5282',
    },
  },
  {
    id: 'chat',
    title: 'Работа в чатах',
    style: {
      headerBg: '#e6f4f1',
      headerBorder: '#b8e0d8',
      countBg: 'rgba(40, 94, 82, 0.12)',
      countColor: '#285e52',
    },
  },
  {
    id: 'phone',
    title: 'Работа по телефону',
    style: {
      headerBg: '#f0ebfa',
      headerBorder: '#d4c8ee',
      countBg: 'rgba(85, 60, 154, 0.12)',
      countColor: '#553c9a',
    },
  },
  {
    id: 'deal',
    title: 'В сделке',
    style: {
      headerBg: '#eaf5ea',
      headerBorder: '#c5e1c5',
      countBg: 'rgba(47, 102, 47, 0.12)',
      countColor: '#2f662f',
    },
  },
  {
    id: 'failed',
    title: 'Провал лида',
    style: {
      headerBg: '#fcebeb',
      headerBorder: '#f0c8c8',
      countBg: 'rgba(155, 44, 44, 0.12)',
      countColor: '#9b2c2c',
    },
  },
]
