<script lang="ts">
	import { onMount } from 'svelte';
	import * as echarts from 'echarts';
	import type { ECharts, EChartsOption } from 'echarts';

	export let chartType: 'bar' | 'pie' | 'none';
	export let rows: Record<string, unknown>[];
	export let categoryKey: string;
	export let valueKey: string;

	let el: HTMLDivElement | undefined;
	let chart: ECharts | null = null;

	function buildOption(): EChartsOption | null {
		if (chartType === 'none' || !categoryKey || !valueKey || rows.length === 0) return null;

		const pairs = rows.map((r) => ({
			name: String(r[categoryKey] ?? ''),
			value: Number(r[valueKey])
		}));
		const data = pairs.filter((d) => !Number.isNaN(d.value));
		if (data.length === 0) return null;

		if (chartType === 'pie') {
			return {
				backgroundColor: 'transparent',
				textStyle: { color: 'var(--color-text-secondary, #94a3b8)' },
				tooltip: { trigger: 'item' },
				series: [
					{
						type: 'pie',
						radius: ['36%', '68%'],
						data: data.map((d) => ({ name: d.name, value: d.value }))
					}
				]
			};
		}

		return {
			backgroundColor: 'transparent',
			textStyle: { color: 'var(--color-text-secondary, #94a3b8)' },
			tooltip: { trigger: 'axis' },
			grid: { left: 48, right: 24, top: 24, bottom: 72 },
			xAxis: {
				type: 'category',
				data: data.map((d) => d.name),
				axisLabel: { rotate: data.length > 8 ? 35 : 0, interval: 0 }
			},
			yAxis: { type: 'value' },
			series: [{ type: 'bar', data: data.map((d) => d.value) }]
		};
	}

	function applyOption() {
		if (!chart) return;
		const opt = buildOption();
		if (opt) {
			chart.setOption(opt, true);
			chart.resize();
		} else {
			chart.clear();
		}
	}

	function resize() {
		chart?.resize();
	}

	onMount(() => {
		if (!el) return;
		chart = echarts.init(el, undefined, { renderer: 'canvas' });
		window.addEventListener('resize', resize);
		const ro = new ResizeObserver(() => chart?.resize());
		ro.observe(el);
		applyOption();
		return () => {
			ro.disconnect();
			window.removeEventListener('resize', resize);
			chart?.dispose();
			chart = null;
		};
	});

	$: chartType, rows, categoryKey, valueKey, applyOption();
</script>

<div
	bind:this={el}
	class="h-full min-h-[280px] w-full min-w-0 rounded-lg border border-border bg-bg-secondary/40 {chartType === 'none' ? 'hidden' : ''}"
></div>
