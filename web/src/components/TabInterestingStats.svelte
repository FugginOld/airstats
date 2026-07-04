<script>
    import InterestingAircraft from './InterestingAircraft.svelte';
    import { IconTank, IconBuildings, IconShield, IconCurrencyDollar } from '@tabler/icons-svelte';
</script>

<div class="grid grid-cols-1 lg:grid-cols-2 mt-10 gap-6">
    <InterestingAircraft endpoint='api/stats/interesting/military' title="Military Aircraft" icon={IconTank} aircraftType="mil" />
    <InterestingAircraft endpoint='api/stats/interesting/government' title="Government Aircraft" icon={IconBuildings} aircraftType="gov" />
    <InterestingAircraft endpoint='api/stats/interesting/police' title="Police Aircraft" icon={IconShield} aircraftType="pol" />
    <InterestingAircraft endpoint='api/stats/interesting/civilian' title="Civilian Aircraft" icon={IconCurrencyDollar} aircraftType="civ" />
</div>
