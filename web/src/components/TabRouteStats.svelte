<script>
    import "/node_modules/flag-icons/css/flag-icons.min.css";
    import { IconChevronCompactRight } from '@tabler/icons-svelte';
    import RouteTopList from './RouteTopList.svelte';
</script>
<div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 mt-10 gap-6">
    <RouteTopList endpoint="api/stats/routes/airlines" title="Top Airlines" let:row>
        <li class="list-row">
             <div>
                <img src="https://doj0yisjozhv1.cloudfront.net/square-logos/{row.airline_icao}.png" width="40" height="40" alt="{row.airline_icao}">
             </div>
            <div class="list-col-grow">
            <div class="font-medium">{row.airline_name}</div>
            <div class="text-xs uppercase font-semibold opacity-60">{row.airline_icao} / {row.airline_iata}</div>
            </div>
            <div class="text-right">
                <div class="font-semibold">{row.flight_count.toLocaleString()}</div>
                <div class="text-xs opacity-60">flights</div>
            </div>
        </li>
    </RouteTopList>

    <RouteTopList endpoint="api/stats/routes/airports-domestic" title="Top Domestic Airports" let:row>
        <li class="list-row">
            <div class="text-4xl font-thin text-accent opacity-80 font-mono">{row.airport_code}</div>
            <div class="list-col-grow">
             <div class="font-medium">{row.airport_name}</div>
            <div class="text-xs uppercase font-semibold opacity-60">{row.airport_country}</div>
            </div>
            <div class="text-right">
                <div class="font-semibold">{row.flight_count.toLocaleString()}</div>
                <div class="text-xs opacity-60">flights</div>
            </div>
        </li>
    </RouteTopList>

    <RouteTopList endpoint="api/stats/routes/airports-international" title="Top International Airports" let:row>
        <li class="list-row">
            <div class="text-4xl font-thin text-accent opacity-80 font-mono">{row.airport_code}</div>
            <div class="list-col-grow">
             <div class="font-medium">{row.airport_name}</div>
            <div class="text-xs uppercase font-semibold opacity-60">{row.airport_country}</div>
            </div>
            <div class="text-right">
                <div class="font-semibold">{row.flight_count.toLocaleString()}</div>
                <div class="text-xs opacity-60">flights</div>
            </div>
        </li>
    </RouteTopList>

    <RouteTopList endpoint="api/stats/routes/routes" title="Top Routes" let:row>
        <li class="list-row">
            <div class="flex items-center justify-center">
                    <div class="text-4xl font-thin text-accent opacity-80 font-mono">{row.origin_iata_code}</div>
                    <IconChevronCompactRight class="text-accent" stroke={2} />
                    <div class="text-4xl font-thin text-accent opacity-80 font-mono">{row.destination_iata_code}</div>
            </div>
            <div class="list-col-grow flex flex-col justify-center">
                <div class="text-xs font-medium">{row.origin_name}</div>
                <div class="text-xs font-medium">{row.destination_name}</div>
            </div>
            <div class="text-right">
                <div class="font-semibold">{row.flight_count.toLocaleString()}</div>
                <div class="text-xs opacity-60">flights</div>
            </div>
        </li>
    </RouteTopList>

    <RouteTopList endpoint="api/stats/routes/countries-origin" title="Top Origin Countries" let:row>
        <li class="list-row">
            <div><span class="text-4xl fi fi-{row.country_iso.toLowerCase()}"></span></div>
            <div class="list-col-grow">
            <div class="font-medium">{row.country_name}</div>
            <div class="text-xs uppercase font-semibold opacity-60">{row.country_iso}</div>
            </div>
            <div class="text-right">
                <div class="font-semibold">{row.flight_count.toLocaleString()}</div>
                <div class="text-xs opacity-60">flights</div>
            </div>
        </li>
    </RouteTopList>

    <RouteTopList endpoint="api/stats/routes/countries-destination" title="Top Destination Countries" let:row>
        <li class="list-row">
            <div><span class="text-4xl fi fi-{row.country_iso.toLowerCase()}"></span></div>
            <div class="list-col-grow">
            <div class="font-medium">{row.country_name}</div>
            <div class="text-xs uppercase font-semibold opacity-60">{row.country_iso}</div>
            </div>
            <div class="text-right">
                <div class="font-semibold">{row.flight_count.toLocaleString()}</div>
                <div class="text-xs opacity-60">flights</div>
            </div>
        </li>
    </RouteTopList>
  </div>
