/** Высоты фиксированных блоков layout (px) — для расчёта прокрутки канбана */
export const APP_HEADER_HEIGHT = 64
export const LEADS_SECTION_HEADER_HEIGHT = 56
export const LEADS_KANBAN_PADDING_TOP = 16

export function getLeadsKanbanHeight(): number {
  return (
    window.innerHeight -
    APP_HEADER_HEIGHT -
    LEADS_SECTION_HEADER_HEIGHT
  )
}
