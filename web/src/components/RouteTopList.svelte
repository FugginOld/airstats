<script>
    import SkeletonRouteTable from './SkeletonRouteTable.svelte';
    import { refreshRouteData } from '../stores/settings';
    import { createPolledResource } from '../lib/pollResource.js';

    export let endpoint;
    export let title;

    const resource = createPolledResource(endpoint, { refreshMs: 60000, refreshTrigger: refreshRouteData });
    $: data = $resource.data || [];
    $: loading = $resource.loading;
    $: error = $resource.error;
</script>

 {#if loading}
    <SkeletonRouteTable/>
{:else if error}
    <div class="flex alert alert-error">
        <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span>Something went wrong: {error}</span>
    </div>
{:else if data.length === 0}
    <div class="alert alert-info">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
        <span>No data available</span>
    </div>
{:else}
    <ul class="list bg-base-100 rounded-box shadow-md soft-divider">
    <li class="p-4 pb-2 text-xs opacity-60 tracking-wide">{title}</li>

    {#each data as row, index}
        <slot {row} {index} />
    {/each}
    </ul>
{/if}
