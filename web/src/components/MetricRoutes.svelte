<script>
    import { IconRoute, IconWorld, IconBuildingAirport } from '@tabler/icons-svelte';
    import NumberFlow from '@number-flow/svelte'
    import SkeletonMetrics from './SkeletonMetrics.svelte';
    import { createPolledResource } from '../lib/pollResource.js';

    const resource = createPolledResource('api/stats/routes/metrics', { refreshMs: 2000 });
</script>
{#if $resource.loading}
<div class="flex justify-center gap-10">
    <SkeletonMetrics />
</div>
{:else if $resource.error}
    <div class="alert alert-error">
        <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span>Something went wrong: {$resource.error}</span>
    </div>
{:else if $resource.data.length === 0}
    <div class="alert alert-info">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
        <span>No data available</span>
    </div>
{:else}
    <div class="p-2 stats stats-vertical lg:stats-horizontal bg-base-100 shadow-sm rounded-xl hover:shadow-md transition-all duration-200">
        <div class="stat">
            <div class="stat-figure">
                <div class="icon icon-tabler icons-tabler-outline">
                    <IconRoute stroke={2}/>
                </div>
            </div>
            <div class="stat-title">Flights with routes</div>
            <div class="stat-value"><NumberFlow willChange={true} respectMotionPreference={false} value={$resource.data.total_routes} /></div>
            <div class="stat-desc">all time</div>
        </div>
        <div class="stat">
            <div class="stat-figure">
                <div class="icon icon-tabler icons-tabler-outline">
                    <IconWorld stroke={2}/>
                </div>
            </div>
            <div class="stat-title">Unique Countries</div>
            <div class="stat-value"><NumberFlow willChange={true} respectMotionPreference={false} value={$resource.data.unqiue_countries} /></div>
            <div class="stat-desc">all time</div>
        </div>
        <div class="stat">
            <div class="stat-figure">
                <div class="icon icon-tabler icons-tabler-outline">
                    <IconBuildingAirport stroke={2}/>
                </div>
            </div>
            <div class="stat-title">Unique Airports</div>
            <div class="stat-value"><NumberFlow willChange={true} respectMotionPreference={false} value={$resource.data.unique_airports} /></div>
            <div class="stat-desc">all time</div>
        </div>
    </div>
{/if}
