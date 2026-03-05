package ipc

import (
	"sync"
	"time"
)

// defaultLogBufferSize is the number of log entries retained in the ring buffer.
const defaultLogBufferSize = 500

// StatsTracker tracks sync statistics and state for the IPC status provider.
type StatsTracker struct {
	mu             sync.RWMutex
	state          string // "idle", "syncing", "error"
	lastSync       time.Time
	lastError      string
	notesSynced    int
	tagsSynced     int
	queueProcessed int
	lastDurationMs int64

	logBuf  []LogEntry
	logSize int

	syncTrigger chan struct{}
}

// NewStatsTracker creates a new StatsTracker with the given log buffer size.
func NewStatsTracker(logBufferSize int) *StatsTracker {
	if logBufferSize <= 0 {
		logBufferSize = defaultLogBufferSize
	}
	return &StatsTracker{
		state:       "idle",
		logBuf:      make([]LogEntry, 0, logBufferSize),
		logSize:     logBufferSize,
		syncTrigger: make(chan struct{}, 1),
	}
}

// GetStatus returns the current status for IPC clients.
func (st *StatsTracker) GetStatus() StatusResponse {
	st.mu.RLock()
	defer st.mu.RUnlock()

	var lastSyncStr string
	if !st.lastSync.IsZero() {
		lastSyncStr = st.lastSync.UTC().Format(time.RFC3339)
	}

	return StatusResponse{
		State:     st.state,
		LastSync:  lastSyncStr,
		LastError: st.lastError,
		Stats: SyncStats{
			NotesSynced:    st.notesSynced,
			TagsSynced:     st.tagsSynced,
			QueueProcessed: st.queueProcessed,
			LastDurationMs: st.lastDurationMs,
		},
	}
}

// TriggerSync sends a non-blocking signal to trigger an immediate sync.
func (st *StatsTracker) TriggerSync() {
	select {
	case st.syncTrigger <- struct{}{}:
	default:
		// Already triggered, skip.
	}
}

// SyncTriggered returns the channel that receives sync trigger signals.
func (st *StatsTracker) SyncTriggered() <-chan struct{} {
	return st.syncTrigger
}

// GetLogs returns the last n log entries from the ring buffer.
func (st *StatsTracker) GetLogs(n int) []LogEntry {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if n <= 0 || len(st.logBuf) == 0 {
		return nil
	}

	if n > len(st.logBuf) {
		n = len(st.logBuf)
	}

	start := len(st.logBuf) - n
	result := make([]LogEntry, n)
	copy(result, st.logBuf[start:])
	return result
}

// SetSyncing marks the bridge as currently syncing.
func (st *StatsTracker) SetSyncing() {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.state = "syncing"
}

// SetIdle marks the bridge as idle after a successful sync.
func (st *StatsTracker) SetIdle() {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.state = "idle"
	st.lastSync = time.Now()
}

// SetError marks the bridge as in error state.
func (st *StatsTracker) SetError(errMsg string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.state = "error"
	st.lastError = errMsg
}

// RecordSync records stats from a completed sync cycle.
func (st *StatsTracker) RecordSync(notesSynced, tagsSynced, queueProcessed int, durationMs int64) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.notesSynced += notesSynced
	st.tagsSynced += tagsSynced
	st.queueProcessed += queueProcessed
	st.lastDurationMs = durationMs
}

// AddLog appends a log entry to the ring buffer.
func (st *StatsTracker) AddLog(entry LogEntry) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if len(st.logBuf) >= st.logSize {
		// Shift buffer: drop oldest entry.
		copy(st.logBuf, st.logBuf[1:])
		st.logBuf[len(st.logBuf)-1] = entry
	} else {
		st.logBuf = append(st.logBuf, entry)
	}
}
