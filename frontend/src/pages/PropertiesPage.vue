<template>
    <div class="tab-content">
        <div class="props-header">
            <div class="props-title-row">
                <h2 class="props-title">SERVER PROPERTIES</h2>
                <button class="btn btn-primary" :disabled="saving" @click="save">
                    {{ saving ? 'SAVING...' : 'SAVE' }}
                </button>
            </div>
            <input v-model="filter" type="text" class="filter-input" placeholder="Search properties..." />
        </div>

        <div v-if="loading" class="props-loading">Loading properties...</div>

        <div v-else-if="filteredKeys.length === 0" class="props-empty">No properties found</div>

        <div v-else class="props-grid">
            <div v-for="key in filteredKeys" :key="key" class="prop-row">
                <div class="prop-info">
                    <span class="prop-key">{{ key }}</span>
                </div>
                <div class="prop-control">
                    <label v-if="isBool(key)" class="toggle">
                        <input type="checkbox" :checked="props[key] === 'true'"
                            @change="setProp(key, ($event.target as HTMLInputElement).checked ? 'true' : 'false')" />
                        <span class="toggle-track">
                            <span class="toggle-thumb"></span>
                        </span>
                    </label>
                    <input v-else-if="isNumber(key)" type="number" :value="props[key]"
                        @input="setProp(key, ($event.target as HTMLInputElement).value)" class="prop-input sm" />
                    <input v-else type="text" :value="props[key]"
                        @input="setProp(key, ($event.target as HTMLInputElement).value)" class="prop-input" />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../api'

const emit = defineEmits<{ (e: 'toast', payload: { msg: string; type: string }): void }>()

const props = ref<Record<string, string>>({})
const loading = ref(true)
const saving = ref(false)
const filter = ref('')

const keys = computed(() => Object.keys(props.value).sort())

const filteredKeys = computed(() => {
    if (!filter.value) return keys.value
    const q = filter.value.toLowerCase()
    return keys.value.filter(k =>
        k.toLowerCase().includes(q) || (props.value[k] ?? '').toLowerCase().includes(q)
    )
})

const boolKeys = new Set([
    'allow-nether', 'spawn-monsters', 'spawn-animals', 'spawn-npcs',
    'online-mode', 'pvp', 'hardcore', 'white-list', 'enable-command-block',
    'force-gamemode', 'prevent-proxy-connections', 'enforce-whitelist',
    'spawn-protection', 'enable-rcon', 'enable-query', 'allow-flight',
    'broadcast-console-to-ops', 'broadcast-rcon-to-ops', 'sync-chunk-writes',
    'use-native-transport', 'hide-online-players', 'enforce-secure-profile',
    'log-ips', 'require-resource-pack', 'send-old-pack-style',
])

const numberKeys = new Set([
    'server-port', 'max-players', 'view-distance', 'simulation-distance',
    'max-world-size', 'rate-limit', 'entity-broadcast-range-percentage',
    'max-chained-neighbor-updates', 'network-compression-threshold',
    'rcon.port', 'query.port', 'player-idle-timeout',
    'spawn-protection', 'op-permission-level',
])

function isBool(key: string): boolean {
    const v = props.value[key]
    if (boolKeys.has(key)) return true
    return v === 'true' || v === 'false'
}

function isNumber(key: string): boolean {
    if (numberKeys.has(key)) return true
    const v = props.value[key]
    return v !== '' && !isNaN(Number(v))
}

function setProp(key: string, value: string) {
    props.value[key] = value
}

async function load() {
    loading.value = true
    try {
        const data = await fetch('/api/server/properties').then(r => {
            if (!r.ok) throw new Error('Failed to load')
            return r.json()
        })
        const result: Record<string, string> = {}
        for (const [k, v] of Object.entries(data)) {
            result[k] = String(v ?? '')
        }
        props.value = result
    } catch (e: any) {
        emit('toast', { msg: `Failed to load properties: ${e.message ?? e}`, type: 'danger' })
    } finally {
        loading.value = false
    }
}

async function save() {
    saving.value = true
    try {
        await fetch('/api/server/properties', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(props.value),
        }).then(r => {
            if (!r.ok) throw new Error('Failed to save')
        })
        emit('toast', { msg: 'Properties saved', type: 'success' })
    } catch (e: any) {
        emit('toast', { msg: `Save failed: ${e.message ?? e}`, type: 'danger' })
    } finally {
        saving.value = false
    }
}

onMounted(load)
</script>

<style scoped>
.props-header {
    margin-bottom: 20px;
}

.props-title-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
}

.props-title {
    font-family: 'VT323', monospace;
    font-size: 20px;
    letter-spacing: 2px;
}

.filter-input {
    width: 100%;
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 8px 12px;
    color: var(--text);
    font-family: 'Share Tech Mono', monospace;
    font-size: 13px;
    outline: none;
}

.filter-input:focus {
    border-color: var(--green);
}

.props-loading,
.props-empty {
    text-align: center;
    padding: 40px 20px;
    color: var(--muted);
    font-size: 14px;
}

.props-grid {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
}

.prop-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 16px;
    border-bottom: 1px solid rgba(30, 45, 61, 0.4);
}

.prop-row:last-child {
    border-bottom: none;
}

.prop-key {
    font-size: 13px;
    color: var(--text2);
}

.prop-control {
    flex-shrink: 0;
    margin-left: 16px;
}

.prop-input {
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 5px 10px;
    color: var(--text);
    font-family: 'Share Tech Mono', monospace;
    font-size: 13px;
    width: 200px;
    outline: none;
}

.prop-input.sm {
    width: 100px;
}

.prop-input:focus {
    border-color: var(--green);
}
</style>
