<script setup lang="ts">
import { computed } from 'vue'

const GAUGE_SIZE = 200
const GAUGE_VIEW_HEIGHT = 148
const GAUGE_CENTER_X = 100
const GAUGE_CENTER_Y = 100
const GAUGE_RADIUS = 74
const GAUGE_STROKE = 14
const GAUGE_START_ANGLE = 150
const GAUGE_SWEEP_ANGLE = 240

const props = defineProps<{
  title: string
  percent: number | null
  hint?: string
  loading?: boolean
  color?: string
}>()

const progressColor = computed(() => props.color || '#3b82f6')

const clampedPercent = computed(() => {
  if (props.percent === null) return 0
  return Math.min(100, Math.max(0, props.percent))
})

const displayPercent = computed(() => {
  if (props.percent === null) return '—'
  const percent = clampedPercent.value
  if (Number.isInteger(percent)) return `${percent}%`
  return `${percent.toLocaleString('ru-RU', {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1,
  })}%`
})

const innerTop = GAUGE_CENTER_Y - GAUGE_RADIUS + GAUGE_STROKE / 2
const openingY =
  GAUGE_CENTER_Y + GAUGE_RADIUS * Math.sin((GAUGE_START_ANGLE * Math.PI) / 180)
const labelCenterY = (innerTop + openingY) / 2

const percentY = computed(() => (props.hint ? labelCenterY - 9 : labelCenterY))
const hintY = labelCenterY + 16

const trackPath = computed(() =>
  describeArc(GAUGE_START_ANGLE, GAUGE_START_ANGLE + GAUGE_SWEEP_ANGLE),
)

const progressPath = computed(() => {
  const sweep = (clampedPercent.value / 100) * GAUGE_SWEEP_ANGLE
  if (sweep <= 0) return ''
  return describeArc(GAUGE_START_ANGLE, GAUGE_START_ANGLE + sweep)
})

function polar(angle: number) {
  const radians = (angle * Math.PI) / 180
  return {
    x: GAUGE_CENTER_X + GAUGE_RADIUS * Math.cos(radians),
    y: GAUGE_CENTER_Y + GAUGE_RADIUS * Math.sin(radians),
  }
}

function describeArc(startAngle: number, endAngle: number) {
  const start = polar(startAngle)
  const end = polar(endAngle)
  const delta = ((endAngle - startAngle) + 360) % 360
  const largeArc = delta > 180 ? 1 : 0

  return `M ${start.x} ${start.y} A ${GAUGE_RADIUS} ${GAUGE_RADIUS} 0 ${largeArc} 1 ${end.x} ${end.y}`
}
</script>

<template>
  <section class="analytics-kpi-card">
    <header class="analytics-kpi-card__header">
      <h2 class="analytics-kpi-card__title">{{ title }}</h2>
      <div v-if="$slots['header-actions']" class="analytics-kpi-card__header-actions">
        <slot name="header-actions" />
      </div>
    </header>

    <div
      class="analytics-kpi-card__gauge"
      :class="{ 'analytics-kpi-card__gauge--loading': loading }"
    >
      <svg
        class="analytics-kpi-card__chart"
        :viewBox="`0 0 ${GAUGE_SIZE} ${GAUGE_VIEW_HEIGHT}`"
        role="img"
        :aria-label="`${title}: ${displayPercent}`"
      >
        <path
          class="analytics-kpi-card__track"
          :d="trackPath"
        />
        <path
          v-if="progressPath"
          class="analytics-kpi-card__progress"
          :d="progressPath"
        />
        <text
          class="analytics-kpi-card__percent"
          :x="GAUGE_CENTER_X"
          :y="percentY"
          dy="0.35em"
          text-anchor="middle"
        >
          {{ displayPercent }}
        </text>
        <text
          v-if="hint"
          class="analytics-kpi-card__hint"
          :x="GAUGE_CENTER_X"
          :y="hintY"
          dy="0.35em"
          text-anchor="middle"
        >
          {{ hint }}
        </text>
      </svg>
    </div>
  </section>
</template>

<style scoped>
.analytics-kpi-card {
  display: flex;
  min-width: 0;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #f6f8fa;
  box-sizing: border-box;
}

.analytics-kpi-card__header {
  position: relative;
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #e2e8f0;
}

.analytics-kpi-card__header:has(.analytics-kpi-card__header-actions) .analytics-kpi-card__title {
  padding-right: 42px;
}

.analytics-kpi-card__title {
  margin: 0;
  min-width: 0;
  flex: 1;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.3;
  color: #1a202c;
}

.analytics-kpi-card__header-actions {
  position: absolute;
  top: 50%;
  right: 20px;
  display: flex;
  flex-shrink: 0;
  align-items: center;
  transform: translateY(-50%);
}

.analytics-kpi-card__gauge--loading {
  opacity: 0.55;
}

.analytics-kpi-card__gauge {
  position: relative;
  width: 100%;
  max-width: 220px;
  margin: 8px auto 4px;
}

.analytics-kpi-card__chart {
  display: block;
  width: 100%;
  height: auto;
  overflow: visible;
}

.analytics-kpi-card__track {
  fill: none;
  stroke: #e2e8f0;
  stroke-width: 14;
  stroke-linecap: round;
}

.analytics-kpi-card__progress {
  fill: none;
  stroke: v-bind(progressColor);
  stroke-width: 14;
  stroke-linecap: round;
}

.analytics-kpi-card__percent {
  fill: #1a202c;
  font-size: 28px;
  font-weight: 700;
  letter-spacing: -0.04em;
  font-variant-numeric: tabular-nums;
}

.analytics-kpi-card__hint {
  fill: #718096;
  font-size: 11px;
  font-weight: 500;
}
</style>
