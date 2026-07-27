<template>
    <div class="tab-content">
        <div class="card console-card">

            <div class="card-header">
                <span class="card-title">SERVER CONSOLE</span>
                <div class="console-controls">
                    <input v-model="searchQuery" class="search-input" placeholder="Search logs..." />
                    <button class="icon-btn" @click="downloadLogs" title="Download logs">⬇ log</button>
                    <button class="icon-btn" @click="clearConsole" title="Clear">🗑</button>
                    <button class="icon-btn" @click="scrollToBottom" title="Scroll to bottom">⬇</button>
                    <div class="log-filter">
                        <button v-for="f in logFilters" :key="f" :class="['filter-btn', { active: activeFilter === f }]"
                            @click="activeFilter = f">{{ f }}</button>
                    </div>
                </div>
            </div>

            <div class="console-output" ref="consoleOutput">
                <div v-for="(line, i) in filteredLogs" :key="i" class="log-line" :class="line.type">
                    <span class="log-time">{{ line.time }}</span>
                    <span class="log-level">[{{ line.level }}]</span>
                    <span class="log-msg" v-html="highlightMatch(line.msg)"></span>
                </div>
            </div>

            <div class="console-input-row">
                <span class="console-prompt">❯</span>
                <input v-model="consoleInput" class="console-input" placeholder="Enter server command..."
                    @keydown.enter="sendCommand" @keydown.up.prevent="historyUp" @keydown.down.prevent="historyDown" />
                <button class="btn btn-sm btn-primary" @click="sendCommand">SEND</button>
            </div>

        </div>
    </div>
</template>

<script>
import { api } from '../api.js'
import { store } from '../store.js'

export default {
    name: 'ConsolePage',
    data() {
        return {
            store,
            consoleInput: '',
            activeFilter: 'ALL',
            logFilters: ['ALL', 'INFO', 'WARN', 'ERROR', 'CHAT'],
            cmdHistory: [],
            cmdHistoryIdx: -1,
            searchQuery: '',
        }
    },
    computed: {
        filteredLogs() {
            let logs = this.store.logs
            if (this.activeFilter !== 'ALL') {
                logs = logs.filter(l =>
                    this.activeFilter === 'CHAT' ? l.type === 'chat' : l.level === this.activeFilter
                )
            }
            if (this.searchQuery.trim()) {
                const q = this.searchQuery.toLowerCase()
                logs = logs.filter(l => l.msg.toLowerCase().includes(q) || l.level.toLowerCase().includes(q))
            }
            return logs
        },
    },
    methods: {
        async sendCommand() {
            if (!this.consoleInput.trim()) return
            const cmd = this.consoleInput
            this.cmdHistory.unshift(cmd)
            this.store.addLog('INFO', 'cmd', `> ${cmd}`)
            this.consoleInput = ''
            this.cmdHistoryIdx = -1
            try {
                await api.sendCommand(cmd)
            } catch (e) {
                this.store.addLog('ERROR', 'error', `Command failed: ${e}`)
            }
            this.$nextTick(() => this.scrollToBottom())
        },
        clearConsole() {
            this.store.logs = []
        },
        async downloadLogs() {
            try {
                const blob = await api.downloadServerLogs()
                const url = URL.createObjectURL(blob)
                const a = document.createElement('a')
                a.href = url
                a.download = 'server.log'
                document.body.appendChild(a)
                a.click()
                document.body.removeChild(a)
                URL.revokeObjectURL(url)
            } catch (e) {
                this.store.addLog('ERROR', 'error', `Download failed: ${e}`)
            }
        },
        scrollToBottom() {
            this.$nextTick(() => {
                const el = this.$refs.consoleOutput
                if (el) el.scrollTop = el.scrollHeight
            })
        },
        historyUp() {
            if (this.cmdHistoryIdx < this.cmdHistory.length - 1) {
                this.cmdHistoryIdx++
                this.consoleInput = this.cmdHistory[this.cmdHistoryIdx]
            }
        },
        historyDown() {
            if (this.cmdHistoryIdx > 0) {
                this.cmdHistoryIdx--
                this.consoleInput = this.cmdHistory[this.cmdHistoryIdx]
            } else {
                this.cmdHistoryIdx = -1
                this.consoleInput = ''
            }
        },
        highlightMatch(msg) {
            if (!this.searchQuery.trim()) return this.escapeHtml(msg)
            const q = this.searchQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
            const regex = new RegExp(`(${q})`, 'gi')
            return this.escapeHtml(msg).replace(regex, '<mark class="search-hl">$1</mark>')
        },
        escapeHtml(str) {
            return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
        },
    },
    mounted() {
        this.scrollToBottom()
    },
}
</script>

<style scoped>
.console-card {
    height: calc(100vh - 130px);
    display: flex;
    flex-direction: column;
}

.console-controls {
    display: flex;
    align-items: center;
    gap: 8px;
}

.log-filter {
    display: flex;
    gap: 4px;
}

.filter-btn {
    padding: 3px 8px;
    background: none;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--muted);
    cursor: pointer;
    font-family: 'Share Tech Mono', monospace;
    font-size: 11px;
    transition: all 0.15s;
}

.filter-btn.active {
    border-color: var(--green);
    color: var(--green);
}

.console-output {
    flex: 1;
    overflow-y: auto;
    padding: 12px 16px;
    font-family: 'Share Tech Mono', monospace;
    font-size: 12px;
    line-height: 1.7;
    background: #060810;
}

.log-line {
    display: flex;
    gap: 12px;
}

.log-time {
    color: #374151;
    flex-shrink: 0;
}

.log-level {
    flex-shrink: 0;
    width: 48px;
}

.log-line.info .log-level {
    color: var(--blue);
}

.log-line.warn .log-level {
    color: var(--yellow);
}

.log-line.error .log-level {
    color: var(--red);
}

.log-line.join .log-level {
    color: var(--green);
}

.log-line.chat .log-level {
    color: var(--purple);
}

.log-line.cmd .log-level {
    color: var(--green);
}

.log-line.error .log-msg {
    color: var(--red);
}

.log-line.warn .log-msg {
    color: var(--yellow);
}

.log-line.join .log-msg {
    color: #a7f3d0;
}

.log-msg {
    color: var(--text2);
    word-break: break-all;
}

.console-input-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 16px;
    border-top: 1px solid var(--border);
    background: var(--bg3);
}

.console-prompt {
    color: var(--green);
    font-size: 16px;
}

.console-input {
    flex: 1;
    background: none;
    border: none;
    outline: none;
    color: var(--green);
    font-family: 'Share Tech Mono', monospace;
    font-size: 13px;
    caret-color: var(--green);
}

.console-input::placeholder {
    color: var(--muted);
}

.search-input {
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 4px 10px;
    color: var(--text);
    font-family: 'Share Tech Mono', monospace;
    font-size: 12px;
    width: 160px;
    outline: none;
}

.search-input:focus {
    border-color: var(--green);
}

.search-input::placeholder {
    color: var(--muted);
}

:deep(.search-hl) {
    background: rgba(250, 204, 21, 0.3);
    color: #facc15;
    border-radius: 2px;
    padding: 0 1px;
}
</style>