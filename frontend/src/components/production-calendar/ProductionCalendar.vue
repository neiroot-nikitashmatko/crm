<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useDeals } from '@/composables/useDeals'
import type { ProductionCalendarViewMode } from '@/constants/productionCalendar'
import {
  buildCurrentWeekGrid,
  buildMonthGrid,
  formatMonthTitle,
  formatWeekTitle,
  getWeekdayLabels,
  splitIntoWeeks,
} from '@/utils/calendar'
import {
  buildProductionEntriesByDay,
  formatProductionTime,
  getDateKey,
} from '@/utils/productionCalendar'

const props = defineProps<{
  viewMode: ProductionCalendarViewMode
  selectedDate: Date
  employeeFilter?: string | null
}>()

const emit = defineEmits<{
  'open-deal': [dealId: string]
}>()

const { deals, loadDeals } = useDeals()

const today = new Date()

const weekdayLabels = getWeekdayLabels()
const filteredDeals = computed(() => {
  const filter = props.employeeFilter?.trim()
  if (!filter) return deals.value
  return deals.value.filter((deal) => deal.production.employee.trim() === filter)
})
const entriesByDay = computed(() => buildProductionEntriesByDay(filteredDeals.value))

const weeks = computed(() => {
  if (props.viewMode === 'week') {
    return [buildCurrentWeekGrid(props.selectedDate, today)]
  }

  return splitIntoWeeks(buildMonthGrid(props.selectedDate.getFullYear(), props.selectedDate.getMonth(), today))
})

const calendarTitle = computed(() => {
  if (props.viewMode === 'week') {
    return formatWeekTitle(weeks.value[0] ?? [])
  }

  return formatMonthTitle(props.selectedDate.getFullYear(), props.selectedDate.getMonth())
})

onMounted(() => {
  void loadDeals()
})

function getDayKey(date: Date) {
  return getDateKey(date)
}

function getDayEntries(date: Date) {
  return entriesByDay.value.get(getDateKey(date)) ?? []
}

function handleEntryClick(dealId: string) {
  emit('open-deal', dealId)
}
</script>

<template>
  <section
    class="production-calendar"
    :class="{ 'production-calendar--week': viewMode === 'week' }"
  >
    <header class="production-calendar__toolbar">
      <h2 class="production-calendar__month-title">{{ calendarTitle }}</h2>
    </header>

    <div class="production-calendar__weekdays" aria-hidden="true">
      <span
        v-for="weekday in weekdayLabels"
        :key="weekday"
        class="production-calendar__weekday"
      >
        {{ weekday }}
      </span>
    </div>

    <div class="production-calendar__grid" :style="{ '--week-count': weeks.length }">
      <div
        v-for="(week, weekIndex) in weeks"
        :key="`week-${weekIndex}`"
        class="production-calendar__week"
      >
        <article
          v-for="day in week"
          :key="getDayKey(day.date)"
          class="production-calendar__day"
          :class="{
            'production-calendar__day--outside': !day.isCurrentMonth,
            'production-calendar__day--today': day.isToday,
          }"
        >
          <span class="production-calendar__day-number">{{ day.dayNumber }}</span>
          <div class="production-calendar__day-content">
            <button
              v-for="entry in getDayEntries(day.date)"
              :key="entry.dealId"
              type="button"
              class="production-calendar__entry"
              :title="`${entry.nomenclature} · ${formatProductionTime(entry.dueAt)}`"
              @click="handleEntryClick(entry.dealId)"
            >
              <span class="production-calendar__entry-name">{{ entry.nomenclature }}</span>
              <span class="production-calendar__entry-time">{{ formatProductionTime(entry.dueAt) }}</span>
            </button>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>

<style scoped>
.production-calendar {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  height: 100%;
  min-height: 0;
  --day-inset-x: 8px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
  overflow: hidden;
}

.production-calendar__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  padding: 6px 16px;
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
}

.production-calendar__month-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1a202c;
  letter-spacing: -0.02em;
  line-height: 1.25;
}

.production-calendar__weekdays {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  flex-shrink: 0;
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
}

.production-calendar__weekday {
  padding: 6px 12px;
  font-size: 11px;
  font-weight: 600;
  line-height: 1.2;
  color: #64748b;
  text-align: center;
  border-right: 1px solid #e2e8f0;
}

.production-calendar__weekday:last-child {
  border-right: 0;
}

.production-calendar__grid {
  display: grid;
  grid-template-rows: repeat(var(--week-count), minmax(0, 1fr));
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
}

.production-calendar__week {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  min-height: 0;
}

.production-calendar__day {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  min-width: 0;
  min-height: 0;
  padding: clamp(6px, 1.2vh, 10px) var(--day-inset-x);
  border-right: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
  background: #ffffff;
  overflow: hidden;
}

.production-calendar__day:last-child {
  border-right: 0;
}

.production-calendar__week:last-child .production-calendar__day {
  border-bottom: 0;
}

.production-calendar__day--outside {
  background: #f8fafc;
}

.production-calendar__day--today {
  background: #eff6ff;
  box-shadow: inset 0 0 0 2px #93c5fd;
}

.production-calendar__day-number {
  flex-shrink: 0;
  font-size: clamp(13px, 1.6vh, 15px);
  font-weight: 600;
  color: #1a202c;
  line-height: 1.2;
}

.production-calendar__day--outside .production-calendar__day-number {
  color: #94a3b8;
  font-weight: 500;
}

.production-calendar__day--today .production-calendar__day-number {
  color: #1d4ed8;
}

.production-calendar__day-content {
  flex: 1 1 auto;
  width: 100%;
  min-height: 0;
  margin-top: 4px;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 3px;
  overflow-y: auto;
}

.production-calendar__entry {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
  width: 100%;
  padding: 4px 6px;
  border: 0;
  border-radius: 4px;
  background: #e8eef4;
  color: #334155;
  font: inherit;
  font-size: clamp(10px, 1.2vh, 12px);
  line-height: 1.2;
  min-width: 0;
  box-sizing: border-box;
  text-align: left;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    color 0.15s ease;
}

.production-calendar__entry:hover {
  background: #dbe4ee;
  color: #1a202c;
}

.production-calendar__entry:focus-visible {
  outline: 2px solid rgba(31, 136, 61, 0.45);
  outline-offset: 1px;
}

.production-calendar__entry-name {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.production-calendar__entry-time {
  flex-shrink: 0;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.production-calendar--week .production-calendar__day {
  padding: 12px var(--day-inset-x);
}

.production-calendar--week .production-calendar__day-number {
  font-size: 16px;
}

.production-calendar--week .production-calendar__day-content {
  gap: 6px;
  margin-top: 8px;
}

.production-calendar--week .production-calendar__entry {
  padding: 6px 8px;
  font-size: 13px;
  border-radius: 6px;
}
</style>
