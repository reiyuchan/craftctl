<template>
    <div class="tab-content">

        <!-- Stats Grid -->
        <div class="stats-grid">
            <div class="stat-card" v-for="stat in statCards" :key="stat.label">
                <div class="stat-icon">{{ stat.icon }}</div>
                <div class="stat-info">
                    <span class="stat-value">{{ stat.value }}</span>
                    <span class="stat-label">{{ stat.label }}</span>
                </div>
                <div class="stat-trend" :class="stat.trend">{{ stat.trendVal }}</div>
            </div>
        </div>

        <div class="dashboard-grid">

            <!-- Performance Chart -->
            <div class="card chart-card">
                <div class="card-header">
                    <span class="card-title">PERFORMANCE</span>
                    <div class="chart-tabs">
                        <button v-for="t in ['TPS', 'RAM', 'CPU']" :key="t"
                            :class="['chart-tab', { active: activeChart === t }]" @click="activeChart = t">{{ t
                            }}</button>
                    </div>
                </div>
                <div class="chart-area">
                    <svg viewBox="0 0 400 100" class="perf-chart" preserveAspectRatio="none">
                        <defs>
                            <linearGradient id="chartGrad" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="0%" stop-color="var(--green)" stop-opacity="0.4" />
                                <stop offset="100%" stop-color="var(--green)" stop-opacity="0" />
                            </linearGradient>
                        </defs>
                        <path :d="chartAreaPath" fill="url(#chartGrad)" />
                        <path :d="chartLinePath" fill="none" stroke="var(--green)" stroke-width="2" />
                        <circle v-for="(pt, i) in chartPoints" :key="i" :cx="pt.x" :cy="pt.y" r="3" fill="var(--green)"
                            class="chart-dot" />
                    </svg>
                    <div class="chart-labels">
                        <span v-for="l in ['5m', '4m', '3m', '2m', '1m', 'now']" :key="l" class="chart-time">{{ l }}</span>
                    </div>
                </div>
            </div>

            <!-- Online Players -->
            <div class="card players-card">
                <div class="card-header">
                    <span class="card-title">ONLINE PLAYERS</span>
                    <span class="player-count">{{ store.onlinePlayers.length }}/{{ store.maxPlayers }}</span>
                </div>
                <div class="player-list">
                    <div v-for="p in store.onlinePlayers" :key="p.name" class="player-row">
                        <div class="player-avatar" :style="{ background: p.color }">{{ p.name[0] }}</div>
                        <div class="player-info">
                            <span class="player-name">{{ p.name }}</span>
                            <span class="player-time">{{ p.time }}</span>
                        </div>
                        <div class="player-ping" :class="pingClass(p.ping)">{{ p.ping }}ms</div>
                        <button class="player-action" @click="store.kickPlayer(p.name)" title="Kick">✕</button>
                    </div>
                    <div v-if="store.onlinePlayers.length === 0" class="empty-state">
                        No players online
                    </div>
                </div>
            </div>

        </div>
    </div>
</template>

<script>
import { store } from '../store.js'
import { events } from '../api'

export default {
    name: 'DashboardPage',
    data() {
        return {
            store,
            activeChart: 'CPU',
            cleanup: null,
        }
    },
    computed: {
        statCards() {
            const s = this.store.currentStats
            return [
                { icon: '⚡', label: 'CPU Usage', value: s.cpu.toFixed(1) + '%', trend: 'neutral', trendVal: '' },
                { icon: '🧠', label: 'RAM Usage', value: (s.ram / (1024 * 1024 * 1024)).toFixed(2) + ' GB', trend: 'neutral', trendVal: s.ramPercent.toFixed(1) + '%' },
                { icon: '🧵', label: 'Threads', value: String(s.threads), trend: 'neutral', trendVal: '' },
                { icon: '⏱', label: 'Status', value: this.store.serverStatus === 'running' ? 'Online' : 'Offline', trend: this.store.serverStatus === 'running' ? 'up' : 'down', trendVal: this.store.serverStatus === 'running' ? '●' : '○' },
            ]
        },
        chartPoints() {
            const data = this.store.chartData[this.activeChart]
            if (!data || data.length === 0) return []
            const max = this.activeChart === 'TPS' ? 20 : this.activeChart === 'RAM' ? 8 : 100
            return data.map((v, i) => ({
                x: (i / Math.max(data.length - 1, 1)) * 400,
                y: 100 - (v / max) * 90,
            }))
        },
        chartLinePath() {
            return this.chartPoints.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ')
        },
        chartAreaPath() {
            return this.chartPoints.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ')
                + ' L 400 100 L 0 100 Z'
        },
    },
    async mounted() {
        await this.store.fetchServerStats()
        this.cleanup = events.onServerStats((data) => {
            this.store.currentStats = data
            this.store.chartData.CPU.push(data.cpu)
            this.store.chartData.RAM.push(data.ram / (1024 * 1024 * 1024))
            this.store.chartData.TPS.push(20)
            if (this.store.chartData.CPU.length > 60) {
                this.store.chartData.CPU.shift()
                this.store.chartData.RAM.shift()
                this.store.chartData.TPS.shift()
            }
        })
    },
    beforeUnmount() {
        if (this.cleanup) this.cleanup()
    },
    methods: {
        pingClass(ping) {
            if (ping < 50) return 'ping-good'
            if (ping < 100) return 'ping-ok'
            return 'ping-bad'
        },
    },
}
</script>

<style scoped>
.stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
    margin-bottom: 20px;
}

.stat-card {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 16px;
    display: flex;
    align-items: center;
    gap: 14px;
    transition: border-color 0.2s;
}

.stat-card:hover {
    border-color: var(--green2);
}

.stat-icon {
    font-size: 28px;
}

.stat-value {
    display: block;
    font-family: 'VT323', monospace;
    font-size: 28px;
    line-height: 1;
    color: var(--text);
}

.stat-label {
    display: block;
    font-size: 11px;
    color: var(--muted);
    margin-top: 2px;
}

.stat-trend {
    margin-left: auto;
    font-size: 11px;
}

.stat-trend.up {
    color: var(--green);
}

.stat-trend.down {
    color: var(--red);
}

.dashboard-grid {
    display: grid;
    grid-template-columns: 1.6fr 1fr;
    gap: 16px;
}

/* Chart */
.chart-tabs {
    display: flex;
    gap: 4px;
}

.chart-tab {
    padding: 3px 10px;
    background: none;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--muted);
    cursor: pointer;
    font-family: 'Share Tech Mono', monospace;
    font-size: 11px;
    transition: all 0.15s;
}

.chart-tab.active {
    border-color: var(--green);
    color: var(--green);
}

.chart-area {
    padding: 16px 16px 8px;
}

.perf-chart {
    width: 100%;
    height: 100px;
    display: block;
}

.chart-dot {
    transition: r 0.2s;
}

.chart-labels {
    display: flex;
    justify-content: space-between;
    padding: 4px 0 0;
}

.chart-time {
    font-size: 10px;
    color: var(--muted);
}

/* Players panel */
.player-count {
    font-size: 13px;
    color: var(--text2);
}

.player-list {
    padding: 8px 0;
}

.player-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 16px;
    transition: background 0.1s;
}

.player-row:hover {
    background: var(--bg3);
}

.player-name {
    display: block;
    font-size: 13px;
}

.player-time {
    display: block;
    font-size: 11px;
    color: var(--muted);
}

.player-info {
    flex: 1;
}

.player-ping {
    font-size: 12px;
    padding: 2px 7px;
    border-radius: 3px;
}

.ping-good {
    color: var(--green);
    background: rgba(74, 222, 128, 0.1);
}

.ping-ok {
    color: var(--yellow);
    background: rgba(251, 191, 36, 0.1);
}

.ping-bad {
    color: var(--red);
    background: rgba(248, 113, 113, 0.1);
}

.player-action {
    background: none;
    border: none;
    color: var(--muted);
    cursor: pointer;
    padding: 4px 6px;
    font-size: 12px;
    transition: color 0.15s;
}

.player-action:hover {
    color: var(--red);
}

.empty-state {
    padding: 24px;
    text-align: center;
    color: var(--muted);
}
</style>