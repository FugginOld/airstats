<script>
// @ts-nocheck
    import { onMount, onDestroy } from 'svelte';
    import { themeChanged } from '../../lib/themeStore.js';
    import {
        Chart,
        BarController,
        BarElement,
        CategoryScale,
        LinearScale,
        TimeScale,
        Title,
        Tooltip,
        Legend
    } from 'chart.js';
    import 'chartjs-adapter-date-fns';

    export let type = 'flights'; // 'flights' or 'aircraft'
    export let period = 'day'; // 'year', 'month', or 'day'

    /* Data setup */

    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    const endpoint = `api/stats/charts/${type}/${period}?tz=${encodeURIComponent(timezone)}`;

    async function fetchData() {
        try {
            loading = true;
            const response = await fetch(endpoint);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            const result = await response.json();
            
            if (result.series && result.series.length > 0) {
                const series = result.series[0];
                chartData = {
                    labels: series.points.map(p => new Date(p.x)),
                    dataPoints: series.points.map(p => p.y),
                    label: series.label || (type === 'flights' ? 'Flights Over Time' : 'Aircraft Over Time'),
                    unit: result.x?.unit || (period === 'year' ? 'month' : period === 'month' ? 'day' : 'hour')
                };
            }
            
            error = null;
        } catch (err) {
            error = err.message || 'Failed to load data';
        } finally {
            loading = false;
        }
    }


    /* Chart Setup */

    Chart.register(
        BarController,
        BarElement,
        CategoryScale,
        LinearScale,
        TimeScale,
        Title,
        Tooltip,
        Legend
    );

    let chartCanvas = null;
    let chart = null;
    let loading = true;
    let error = null;
    let chartData = null;
    let hasRenderedOnce = false;

    // set fallback colours
    let CHART_COLOURS = {
        primary: '#570df8', 
        primaryAlpha: 'rgba(87, 13, 248, 0.2)'
    };

    const MONTH_NAMES = [
        'January', 'February', 'March', 'April', 'May', 'June',
        'July', 'August', 'September', 'October', 'November', 'December'
    ];
    
    
    // this is silly, but seems to be the only way to dynamically get the value of daisyui theme colours
    function getDaisyUIColor(className) {
        const temp = document.createElement('div');
        temp.className = className;
        document.body.appendChild(temp);
        const color = getComputedStyle(temp).backgroundColor || getComputedStyle(temp).color;
        document.body.removeChild(temp);
        return color;
    }

    // get colours from daisyui theme
    function updateChartColors() {
        CHART_COLOURS = {
            primary: getDaisyUIColor('bg-primary'),
            primaryAlpha: getDaisyUIColor('bg-primary') + '33',
            baseContent: getDaisyUIColor('bg-base-content')
        };
        
        // update chart colours and redraw
        if (chart && chartData) {
            chart.data.datasets[0].borderColor = CHART_COLOURS.primary;
            chart.data.datasets[0].backgroundColor = CHART_COLOURS.primaryAlpha;
            chart.options.scales.x.ticks.color = CHART_COLOURS.baseContent;
            chart.options.scales.y.ticks.color = CHART_COLOURS.baseContent;
            chart.options.scales.x.border.color = CHART_COLOURS.baseContent;
            chart.options.scales.y.border.color = CHART_COLOURS.baseContent;
            chart.update();
        }
    }

    // Get tooltip title based on period
    function getTooltipTitle(tooltipItems) {
        const index = tooltipItems[0].dataIndex;
        const date = chartData.labels[index];
        
        if (period === 'year') {
            return `${MONTH_NAMES[date.getMonth()]} ${date.getFullYear()}`;
        } else if (period === 'month') {
            return date.toLocaleDateString('en-US', { 
                weekday: 'short', 
                year: 'numeric', 
                month: 'short', 
                day: 'numeric' 
            });
        } else {
            // For day view, show hour
            return date.toLocaleString('en-US', { 
                weekday: 'short',
                month: 'short', 
                day: 'numeric',
                hour: 'numeric',
                minute: '2-digit',
                hour12: true
            });
        }
    }

    // Get x-axis label based on period
    function getXAxisLabel(value, index) {
        const date = chartData.labels[index];
        if (!date) return value;
        
        if (period === 'year') {
            return `${MONTH_NAMES[date.getMonth()].slice(0, 3)} ${date.getFullYear()}`;
        } else if (period === 'month') {
            return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
        } else {
            // For day view, show hour
            return date.toLocaleString('en-US', { 
                hour: 'numeric',
                hour12: true
            });
        }
    }

    function getXAxisTickConfig() {
        const base = {
            color: CHART_COLOURS.baseContent,
            callback: getXAxisLabel
        };
        
        if (period === 'month') {
            return {
                ...base,
                autoSkip: false,
                maxRotation: 45,
                minRotation: 45
            };
        }
        
        return base;
    }
      
    function createChart() {
        if (!chartCanvas || !chartData) return;
        
        // check container exists before creating chart, otherwise theres a weird loading animation
        if (chartCanvas.offsetWidth === 0 || chartCanvas.offsetHeight === 0) {
            requestAnimationFrame(() => createChart());
            return;
        }
        
        // dataset / chart
        const dataset = {
            label: chartData.label,
            data: chartData.dataPoints,
            borderColor: CHART_COLOURS.primary,
            backgroundColor: CHART_COLOURS.primaryAlpha,
            borderWidth: 1,
            barPercentage: 0.90,
            categoryPercentage: 0.95
        };
        
        // tooltip
        const tooltipCallbacks = {
            title: getTooltipTitle,
            label: (context) => {
                const value = context.parsed.y;
                const unit = type === 'flights' ? 'flights' : 'aircraft';
                return value.toLocaleString() + ' ' + unit;
            }
        };
        
        // plugin configuration
        const plugins = {
            title: {
                display: false,
                text: 'Flights'
            },
            legend: {
                display: false
            },
            tooltip: {
                intersect: false,
                mode: 'index',
                callbacks: tooltipCallbacks
            }
        };
        
        // scale configuration
        const scales = {
            x: {
                type: 'category',
                grid: {
                    display: false
                },
                ticks: getXAxisTickConfig(),
                border: {
                    color: CHART_COLOURS.baseContent
                }
            },
            y: {
                beginAtZero: true,
                grid: {
                    display: false
                },
                ticks: {
                    color: CHART_COLOURS.baseContent,
                    callback: (value) => value.toLocaleString()
                },
                border: {
                    color: CHART_COLOURS.baseContent
                }
            }
        };
        
        // chart options
        const options = {
            responsive: true,
            maintainAspectRatio: false,
            animation: true,
            plugins,
            scales
        };
        
        // main chart config
        const config = {
            type: 'bar',
            data: {
                labels: chartData.labels,
                datasets: [dataset]
            },
            options
        };

        if (chart) {
            chart.destroy();
        }
        
        chart = new Chart(chartCanvas, config);
    }
    
    
    $: if (!loading && chartData && chartCanvas) {
        setTimeout(() => createChart(), 250);
    }
    
    let unsubscribe;

    onMount(async () => {
        // get colours for first time
        updateChartColors();
        
        // sub to theme change events
        unsubscribe = themeChanged.subscribe(() => {
            setTimeout(() => {
                updateChartColors();
            }, 50);
        });
        
        // get data
        await fetchData();

        // draw chart
        if (chartData && chartCanvas) {
            setTimeout(() => createChart(), 250);
        }
    });

    onDestroy(() => {
        if (chart) {
            chart.destroy();
            chart = null;
        }
        if (unsubscribe) {
            unsubscribe();
        }
    });
</script>

{#if loading}
    <div class="h-100 flex items-center justify-center">
        <span class="loading loading-bars loading-xl"></span>
    </div>
{:else if error}
    <div class="flex alert alert-error">
        <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span>Something went wrong: {error}</span>
    </div>
{:else if !chartData}
    <div class="alert alert-info">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
        <span>No data available</span>
    </div>
{:else}
    <div class="h-100">
        <canvas bind:this={chartCanvas}></canvas>
    </div>
{/if}
