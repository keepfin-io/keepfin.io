<template>
  <div ref="container" class="portfolio-chart">
    <p v-if="zen" class="py-16 text-center text-sm text-muted-foreground">{{ $t('wealth.hiddenValues') }}</p>
    <template v-else-if="paths.length">
      <svg :viewBox="`0 0 ${chartWidth} 250`" class="chart-svg" role="img" :aria-label="$t(mode === 'price' ? 'holdingDetail.chartDescription' : 'portfolio.chartDescription')">
        <g v-for="tick in ticks" :key="tick.y"><line x1="54" :y1="tick.y" :x2="chartWidth - 12" :y2="tick.y" stroke="currentColor" opacity="0.09" stroke-dasharray="3 5" /><text x="0" :y="tick.y + 4" fill="currentColor" class="axis-label">{{ tick.label }}</text></g>
        <path v-for="(path, index) in paths" :key="index" :d="path" fill="none" stroke="var(--brand-accent)" stroke-width="2.5" stroke-linejoin="round" stroke-linecap="round" />
        <circle v-for="(point, index) in isolatedPoints" :key="`dot-${index}`" :cx="point.x" :cy="point.y" r="3.5" fill="var(--brand-accent)" />
        <path v-for="(path, index) in benchmarkPaths" :key="`benchmark-${index}`" :d="path" fill="none" stroke="currentColor" opacity="0.55" stroke-dasharray="6 5" stroke-width="1.75" />
      </svg>
      <div class="chart-dates"><span>{{ points[0]?.date }}</span><span>{{ points[points.length - 1]?.date }}</span></div>
    </template>
    <div v-else class="portfolio-empty"><ChartNoAxesCombined /><p>{{ $t('portfolio.noHistory') }}</p></div>
    <p v-if="mode === 'return'" class="chart-legend">{{ $t('portfolio.chartLegend') }}</p>
    <details v-if="points.length && !zen" class="chart-data">
      <summary>{{ $t('portfolio.viewData') }}<ChevronDown class="size-4" aria-hidden="true" /></summary>
      <div class="chart-data-scroll" tabindex="0" role="region" :aria-label="$t('portfolio.viewData')">
        <table v-if="mode === 'price'" class="price-data"><thead><tr><th>{{ $t('portfolio.date') }}</th><th>{{ $t('holdingDetail.closePrice') }} · {{ currency }}</th></tr></thead><tbody><tr v-for="point in points" :key="point.date"><td>{{ point.date }}</td><td>{{ number(point.price) }}</td></tr></tbody></table>
        <table v-else><thead><tr><th>{{ $t('portfolio.date') }}</th><th>{{ $t('portfolio.value') }}</th><th>{{ $t('portfolio.netFlow') }}</th><th>{{ $t('portfolio.return') }}</th><th>{{ $t('portfolio.benchmark') }}</th></tr></thead><tbody><tr v-for="point in points" :key="point.date"><td>{{ point.date }}</td><td>{{ number(point.value) }}</td><td>{{ number(point.net_flow) }}</td><td>{{ percent(point.return) }}</td><td>{{ percent(point.benchmark) }}</td></tr></tbody></table>
      </div>
    </details>
  </div>
</template>
<script setup>
import { zen } from '@/composables/useZen'
import { computed, ref } from 'vue'
import { useElementSize } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { ChartNoAxesCombined, ChevronDown } from 'lucide-vue-next'
import { formatNumber, formatPercent } from '@/lib/portfolio-format'
const props = defineProps({ points: { type: Array, default: () => [] }, mode: { type: String, default: 'value' }, currency: { type: String, default: '' } })
const { locale } = useI18n()
const container = ref(null)
const { width } = useElementSize(container)
const chartWidth = computed(() => Math.max(200, width.value || 640))
const values = computed(() => props.points.flatMap(p => props.mode === 'return' ? [p.return, p.benchmark] : [p[props.mode]]).filter(valid))
const min = computed(() => Math.min(...values.value, 0))
const max = computed(() => Math.max(...values.value, 1))
function valid(v) { return v != null && Number.isFinite(v) }
function coordinates(value, index) {
  return { x: props.points.length === 1 ? (54 + chartWidth.value - 12) / 2 : 54 + index / (props.points.length - 1) * (chartWidth.value - 66), y: 226 - (value - min.value) / (max.value - min.value || 1) * 206 }
}
function series(key) {
  const lines = []; let line = ''
  props.points.forEach((p, index) => {
    if (!valid(p[key])) { if (line) lines.push(line); line = ''; return }
    const { x, y } = coordinates(p[key], index)
    line += `${line ? 'L' : 'M'}${x.toFixed(1)},${y.toFixed(1)} `
  })
  if (line) lines.push(line)
  return lines
}
const paths = computed(() => series(props.mode))
const isolatedPoints = computed(() => props.points.flatMap((p, index) => valid(p[props.mode]) && !valid(props.points[index - 1]?.[props.mode]) && !valid(props.points[index + 1]?.[props.mode]) ? [coordinates(p[props.mode], index)] : []))
const benchmarkPaths = computed(() => props.mode === 'return' ? series('benchmark') : [])
const ticks = computed(() => [max.value, (max.value + min.value) / 2, min.value].map(v => ({ y: coordinates(v, 0).y, label: new Intl.NumberFormat(locale.value, { notation: 'compact', maximumFractionDigits: 1 }).format(v) + (props.mode === 'return' ? '%' : '') })))
function number(v) { return formatNumber(v, locale.value) }
function percent(v) { return formatPercent(v, locale.value) }
</script>
<style scoped>
.portfolio-chart { min-width: 0; width: 100%; }.chart-svg { display: block; width: 100%; height: 250px; }.axis-label { font-size: 10px; opacity: .6; }
.chart-dates { display: flex; justify-content: space-between; gap: 12px; padding-left: 54px; color: var(--muted-foreground); font-size: 10px; }.chart-legend { margin-top: 18px; color: var(--muted-foreground); font-size: 11px; line-height: 1.7; }
.chart-data { margin-top: 20px; border-top: 1px solid var(--border); }.chart-data summary { display: flex; align-items: center; justify-content: space-between; min-height: 44px; cursor: pointer; list-style: none; font-size: 11px; color: var(--muted-foreground); }.chart-data summary::-webkit-details-marker { display: none; }.chart-data[open] summary svg { transform: rotate(180deg); }
.chart-data-scroll { max-height: 256px; overflow: auto; border-radius: 8px; }.chart-data table { width: 100%; min-width: 520px; font-size: 11px; text-align: left; font-variant-numeric: tabular-nums; }.chart-data th, .chart-data td { padding: 10px; white-space: nowrap; border-bottom: 1px solid var(--border); }.chart-data th { font-weight: 500; color: var(--muted-foreground); }
.chart-data .price-data { min-width: 0; }
</style>
