import { writable } from 'svelte/store';

/**
 * Poll `endpoint` on an interval, returning a store of
 * { data, loading, error }. Polling starts on first subscribe and stops on
 * last unsubscribe (Svelte's native store start/stop lifecycle) — no
 * onMount/onDestroy needed in the consuming component.
 *
 * If `refreshTrigger` (a Svelte store) is given, a change to its value also
 * triggers an immediate re-fetch — matching the existing
 * "$: if ($refreshXData) fetchData()" settings-driven refresh pattern. The
 * trigger's first (subscribe-time) callback is skipped, since every Svelte
 * store fires its subscriber immediately with the current value.
 */
export function createPolledResource(endpoint, { refreshMs = 60000, refreshTrigger } = {}) {
	return writable({ data: null, loading: true, error: null }, (set, update) => {
		async function fetchData() {
			try {
				const response = await fetch(endpoint);
				if (!response.ok) {
					throw new Error(`${response.status}`);
				}
				const data = await response.json();
				set({ data, loading: false, error: null });
			} catch (err) {
				update((state) => ({ ...state, loading: false, error: err.message }));
			}
		}

		fetchData();
		const interval = setInterval(fetchData, refreshMs);

		let unsubscribeTrigger;
		if (refreshTrigger) {
			let skippedInitial = false;
			unsubscribeTrigger = refreshTrigger.subscribe(() => {
				if (!skippedInitial) {
					skippedInitial = true;
					return;
				}
				fetchData();
			});
		}

		return () => {
			clearInterval(interval);
			if (unsubscribeTrigger) {
				unsubscribeTrigger();
			}
		};
	});
}
