<script>
  import { IconSettings } from '@tabler/icons-svelte';
  import ThemeSelector from './components/ThemeSelector.svelte';
  import AboveTimeline from './components/AboveTimeline.svelte';
  import TabRouteStats from './components/TabRouteStats.svelte';
  import TabMotionStats from './components/TabMotionStats.svelte';
  import TabInterestingStats from './components/TabInterestingStats.svelte';
  import TabActivity from './components/TabActivity.svelte';
  import Footer from './components/Footer.svelte';
  import Settings from './components/Settings.svelte';

  let activeTab = 'activity';
  let tabsElement;

  const tabs = [
    { name: 'activity', label: 'Activity', component: TabActivity },
    { name: 'route-stat', label: 'Route Information', component: TabRouteStats },
    { name: 'interesting-stat', label: 'Interesting Aircraft', component: TabInterestingStats },
    { name: 'motion-stat', label: 'Record Holders', component: TabMotionStats }
  ];

  function setActiveTab(tabName) {
    activeTab = tabName;
    if (tabsElement) {
      const yOffset = -60;
      const y = tabsElement.getBoundingClientRect().top + window.pageYOffset + yOffset;
      window.scrollTo({ top: y, behavior: 'smooth' });
    }
  }

  function openSettingsModal() {
      document.getElementById("settings-modal").showModal();
  }

</script>


<div class="navbar bg-base-100 shadow-sm">
  <div class="navbar-start">
  </div>
  <div class="navbar-center">
    <h1 class="text-4xl font-normal text-primary drop-shadow-[0_0_15px_rgba(59,130,246,0.5)]">
      Skystats
    </h1>
  </div>
  <div class="navbar-end">
    <div class="mr-4">
      <ThemeSelector/>
    </div>
    <button
        class="btn btn-ghost btn-circle"
        on:click={() => openSettingsModal()}
    >
        <IconSettings class="h-5 w-5" />
    </button>
  </div>
</div>

<div class="container max-w-8xl mx-auto p-8">
  <div class="grid grid-cols-1 mt-10 mb-15">
    <AboveTimeline />
  </div>

  <!-- tabs -->
  <div bind:this={tabsElement} class="tabs mb-6 flex justify-center">
    {#each tabs as tab}
      <button class="mr-4
                    { activeTab === tab.name ?
                      'badge badge-lg badge-primary tab-active text-white' :
                      'badge badge-lg badge-primary badge-outline'
                    }"
      on:click={() => setActiveTab(tab.name)}>
      {tab.label}
      </button>
    {/each}
  </div>

  <!-- tab content -->
  <div style="min-height: 1000px;">
    {#each tabs as tab}
      <div class="{activeTab === tab.name ? 'block fade-in' : 'hidden'}">
        <svelte:component this={tab.component} />
      </div>
    {/each}
  </div>

  <Footer />
</div>

<Settings />

<style>

  .fade-in {
    animation: fadeIn 0.5s ease-in;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  /* Override the divider for DaisyUI list component, as its stopped working in recent versions */
  :global(.soft-divider > :not(:last-child).list-row)::after,
  :global(.soft-divider > :not(:last-child) .list-row)::after {
    opacity: 0.05 !important;
  }
</style>
