import { writable } from 'svelte/store';

const { subscribe, set, update } = writable({});

// arbitrary counter that gets incremented on settings save to 
// trigger refresh (there might be a better way...)
export const refreshRouteData = writable(0);
export const refreshInterestingData = writable(0);
export const refreshRecordHolderData = writable(0);

export const settings = {
    subscribe,

    async load() {
        try {
            const response = await fetch('/api/settings');
            if (response.ok) {
                const data = await response.json();
                const settingsObj = {};
                data.forEach(setting => {
                    settingsObj[setting.setting_key] = setting;
                });
                set(settingsObj);
                return settingsObj;
            }
        } catch (error) {
            // settings load failure is non-fatal; UI will use defaults
        }
    },

    async save(updates) {
        try {
            const response = await fetch('/api/settings', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(updates),
            });

            if (response.ok) {
                const data = await response.json();
                const settingsObj = {};
                data.forEach(setting => {
                    settingsObj[setting.setting_key] = setting;
                });
                set(settingsObj);

                // increment counter to trigger refresh
                if (updates.route_table_limit !== undefined) {
                    refreshRouteData.update(n => n + 1);
                }
                if (updates.interesting_table_limit !== undefined) {
                    refreshInterestingData.update(n => n + 1);
                }
                if (updates.record_holder_table_limit !== undefined) {
                    refreshRecordHolderData.update(n => n + 1);
                }

                return true;
            }
            return false;
        } catch (error) {
            return false;
        }
    },

    // get setting
    getValue(key) {
        let value = null;
        const unsubscribe = subscribe(s => {
            value = s[key]?.setting_value;
        });
        unsubscribe();
        return value;
    }
};

settings.load();