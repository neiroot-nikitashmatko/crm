import { onMounted, onUnmounted, ref } from 'vue'
import { getLeadsKanbanHeight } from '@/constants/layout'

export function useLeadsKanbanLayout() {
  const kanbanHeightPx = ref(getLeadsKanbanHeight())

  function updateKanbanHeight() {
    kanbanHeightPx.value = getLeadsKanbanHeight()
  }

  onMounted(() => {
    window.addEventListener('resize', updateKanbanHeight)
    updateKanbanHeight()
  })

  onUnmounted(() => {
    window.removeEventListener('resize', updateKanbanHeight)
  })

  return {
    kanbanHeightPx,
    updateKanbanHeight,
  }
}
