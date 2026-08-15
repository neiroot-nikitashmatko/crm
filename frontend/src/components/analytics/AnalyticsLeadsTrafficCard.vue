<script setup lang="ts">
import { computed, useId } from 'vue'
import type { TrafficSourceMetric } from '@/types/analytics'

const CHART_SIZE = 216
const CHART_CENTER = CHART_SIZE / 2
const RING_RADIUS = 76
const RING_WIDTH = 12
const INNER_RADIUS = 57
const RING_GAP_ANGLE = 12

interface RingSegment extends TrafficSourceMetric {
  path: string
  share: number
}

const props = defineProps<{
  title: string
  metrics: readonly TrafficSourceMetric[]
  nounOne: string
  nounFew: string
  nounMany: string
  loading?: boolean
  legendAriaLabel?: string
}>()

const titleId = useId()

const sortedMetrics = computed(() => {
  const metrics = Array.isArray(props.metrics) ? props.metrics : []
  return [...metrics].sort((left, right) => right.count - left.count)
})

const legendRowCount = computed(() => Math.max(1, Math.ceil(sortedMetrics.value.length / 2)))

const totalCount = computed(() =>
  sortedMetrics.value.reduce((total, metric) => total + metric.count, 0),
)

const ringSegments = computed<RingSegment[]>(() => {
  const metricsWithCount = sortedMetrics.value.filter((metric) => metric.count > 0)
  if (totalCount.value === 0 || metricsWithCount.length === 0) return []

  const availableAngle = 360 - RING_GAP_ANGLE * metricsWithCount.length
  let currentAngle = -90

  return metricsWithCount.map((metric) => {
    const segmentAngle = (metric.count / totalCount.value) * availableAngle
    const startAngle = currentAngle
    const endAngle = startAngle + segmentAngle

    currentAngle = endAngle + RING_GAP_ANGLE

    return {
      ...metric,
      path: getArcPath(startAngle, endAngle),
      share: getShare(metric.count),
    }
  })
})

function getPointOnCircle(angle: number) {
  const radians = (angle * Math.PI) / 180

  return {
    x: CHART_CENTER + RING_RADIUS * Math.cos(radians),
    y: CHART_CENTER + RING_RADIUS * Math.sin(radians),
  }
}

function getArcPath(startAngle: number, endAngle: number) {
  const start = getPointOnCircle(startAngle)
  const end = getPointOnCircle(endAngle)
  const largeArc = endAngle - startAngle > 180 ? 1 : 0

  return [
    `M ${start.x} ${start.y}`,
    `A ${RING_RADIUS} ${RING_RADIUS} 0 ${largeArc} 1 ${end.x} ${end.y}`,
  ].join(' ')
}

function getEntityWord(count: number): string {
  const lastTwoDigits = count % 100
  const lastDigit = count % 10

  if (lastTwoDigits >= 11 && lastTwoDigits <= 14) return props.nounMany
  if (lastDigit === 1) return props.nounOne
  if (lastDigit >= 2 && lastDigit <= 4) return props.nounFew
  return props.nounMany
}

function getShare(count: number): number {
  if (totalCount.value === 0) return 0
  return Math.round((count / totalCount.value) * 100)
}
</script>

<template>
  <section class="analytics-leads-traffic-card" :aria-labelledby="titleId">
    <header class="analytics-leads-traffic-card__header">
      <h2 :id="titleId" class="analytics-leads-traffic-card__title">
        {{ title }}
      </h2>
      <div v-if="$slots['header-actions']" class="analytics-leads-traffic-card__header-actions">
        <slot name="header-actions" />
      </div>
    </header>

    <div
      class="analytics-leads-traffic-card__body"
      :class="{ 'analytics-leads-traffic-card__body--loading': loading }"
    >
      <div class="analytics-leads-traffic-card__chart-wrap">
      <svg
        class="analytics-leads-traffic-card__chart"
        :viewBox="`0 0 ${CHART_SIZE} ${CHART_SIZE}`"
        role="img"
        :aria-label="`Всего ${totalCount} ${getEntityWord(totalCount)}`"
      >
        <circle
          class="analytics-leads-traffic-card__ring-track"
          :cx="CHART_CENTER"
          :cy="CHART_CENTER"
          :r="RING_RADIUS"
        />
        <path
          v-for="segment in ringSegments"
          :key="segment.source"
          class="analytics-leads-traffic-card__ring-segment"
          :d="segment.path"
          :stroke="segment.color"
          :stroke-width="RING_WIDTH"
        >
          <title>
              {{ segment.source }}: {{ segment.count }} {{ getEntityWord(segment.count) }},
              {{ segment.share }}%
          </title>
        </path>
        <circle
          class="analytics-leads-traffic-card__chart-center"
          :cx="CHART_CENTER"
          :cy="CHART_CENTER"
          :r="INNER_RADIUS"
        />
        <circle
          class="analytics-leads-traffic-card__chart-center-dots"
          :cx="CHART_CENTER"
          :cy="CHART_CENTER"
          :r="INNER_RADIUS - 8"
        />
      </svg>

      <div class="analytics-leads-traffic-card__center" aria-hidden="true">
        <strong class="analytics-leads-traffic-card__total">{{ totalCount }}</strong>
        <span class="analytics-leads-traffic-card__total-label">{{ getEntityWord(totalCount) }}</span>
      </div>
      </div>

      <ul class="analytics-leads-traffic-card__legend" :aria-label="legendAriaLabel || 'Источники трафика'">
        <li
          v-for="metric in sortedMetrics"
          :key="metric.source"
          class="analytics-leads-traffic-card__legend-item"
          :aria-label="`${metric.source}: ${metric.count} ${getEntityWord(metric.count)}, ${getShare(metric.count)}%`"
        >
          <span
            class="analytics-leads-traffic-card__legend-color"
            :style="{ backgroundColor: metric.color }"
            aria-hidden="true"
          />
          <span class="analytics-leads-traffic-card__legend-source" :title="metric.source">
            {{ metric.source }}
          </span>
          <span class="analytics-leads-traffic-card__legend-value">
            {{ metric.count }} ({{ getShare(metric.count) }}%)
          </span>
        </li>
      </ul>
    </div>
  </section>
</template>

<style scoped>
.analytics-leads-traffic-card {
  width: 100%;
  min-width: 0;
  height: 100%;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #f6f8fa;
  box-sizing: border-box;
}

.analytics-leads-traffic-card__header {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 20px;
  border-bottom: 1px solid #e2e8f0;
}

.analytics-leads-traffic-card__header:has(.analytics-leads-traffic-card__header-actions) .analytics-leads-traffic-card__title {
  padding-right: 42px;
}

.analytics-leads-traffic-card__title {
  margin: 0;
  min-width: 0;
  flex: 1;
  font-size: 18px;
  font-weight: 600;
  line-height: 1.35;
  color: #1a202c;
}

.analytics-leads-traffic-card__header-actions {
  position: absolute;
  top: 50%;
  right: 20px;
  display: flex;
  flex-shrink: 0;
  align-items: center;
  transform: translateY(-50%);
}

.analytics-leads-traffic-card__body {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 20px 20px;
}

.analytics-leads-traffic-card__body--loading {
  opacity: 0.55;
  pointer-events: none;
}

.analytics-leads-traffic-card__chart-wrap {
  position: relative;
  width: 216px;
  max-width: 100%;
}

.analytics-leads-traffic-card__chart {
  display: block;
  width: 100%;
  height: auto;
  overflow: visible;
}

.analytics-leads-traffic-card__center {
  position: absolute;
  top: 23.61%;
  left: 23.61%;
  display: flex;
  width: 52.78%;
  height: 52.78%;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  pointer-events: none;
}

.analytics-leads-traffic-card__total {
  font-size: 32px;
  font-weight: 700;
  line-height: 1;
  letter-spacing: -0.04em;
  font-variant-numeric: tabular-nums;
  color: #1a202c;
}

.analytics-leads-traffic-card__total-label {
  font-size: 11px;
  font-weight: 500;
  line-height: 1;
  color: #718096;
}

.analytics-leads-traffic-card__ring-track {
  fill: none;
  stroke: #e7ebef;
  stroke-width: 1;
}

.analytics-leads-traffic-card__ring-segment {
  fill: none;
  stroke-linecap: round;
}

.analytics-leads-traffic-card__chart-center {
  fill: #ffffff;
  stroke: #e2e8f0;
  stroke-width: 1;
}

.analytics-leads-traffic-card__chart-center-dots {
  fill: none;
  stroke: #cbd5e1;
  stroke-width: 1;
  stroke-dasharray: 1 8;
  stroke-linecap: round;
  opacity: 0.8;
}

.analytics-leads-traffic-card__legend {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  grid-template-rows: repeat(v-bind(legendRowCount), auto);
  grid-auto-flow: column;
  align-self: stretch;
  gap: 10px 16px;
  margin: 10px 0 0;
  padding: 0;
  list-style: none;
}

.analytics-leads-traffic-card__legend-item {
  display: grid;
  grid-template-columns: 8px minmax(0, 1fr) auto;
  align-items: center;
  gap: 7px;
  min-width: 0;
  font-size: 12px;
  line-height: 1.35;
}

.analytics-leads-traffic-card__legend-color {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.analytics-leads-traffic-card__legend-source {
  overflow: hidden;
  color: #4a5568;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.analytics-leads-traffic-card__legend-value {
  color: #1a202c;
  font-size: 11px;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  white-space: nowrap;
}

@media (max-width: 480px) {

  .analytics-leads-traffic-card__body {
    padding-right: 16px;
    padding-left: 16px;
  }

  .analytics-leads-traffic-card__legend {
    grid-template-columns: minmax(0, 1fr);
    grid-template-rows: none;
    grid-auto-flow: row;
  }
}
</style>
