// Package reflux is IDUNA.GAME's own native Go port of REFLUX, SHANKPIT's real cross-mod/cross-
// object pub/sub layer (founder real-time: "build in REFLUX"). Same real API/ABI shape as
// SHANKPIT/packages/reflux/reflux_runtime.c and ECOWAR/packages/reflux/reflux_runtime.c before
// it -- a single, shared, append-only ring-buffer action log any dispatcher can push into and any
// subscriber can poll from a remembered cursor, the same "shared log, not push callbacks" shape
// those two C ports already prove out, ported natively to Go here (this host's own language)
// rather than compiled through PARENA/cgo the way scrollmod's own vterm_mod.prn is -- REFLUX's
// real implementation is a ~75-line ring buffer with no PARENA-specific logic in it at all (no
// resonance tables, no game-simulation state), so a direct, same-API native port is the honest,
// proportionate choice here, matching SHANKPIT's own "same API, same ABI, own copy" precedent
// for porting this same runtime a second time.
package reflux

import "sync"

// Real, named action types this app's own dispatchers/subscribers agree on by convention, same
// "both sides hardcode the same literal, documented cross-reference" pattern SHANKPIT's own
// reflux_runtime.h already establishes. Payload convention (A, B, C) is per-action-type,
// documented at each dispatch site.

const (
	// Dispatched once, on launch, when the honor code text is first shown in the terminal.
	// Payload: A = HONOR_CODE_VERSION (see honor_code_text.go), B/C unused.
	ActionHonorCodeShown = 1
	// Dispatched when the real device-auth flow starts (POST /auth/device/start succeeds).
	// Payload: A/B/C unused -- the real device_code/user_code are secrets, not logged here.
	ActionDeviceLinkStarted = 2
	// Dispatched once the device-auth exchange succeeds and a real handle is known.
	// Payload: A = 1 (success sentinel; REFLUX's own I32-only ABI has no string payload --
	// same real constraint reflux_runtime.h's own header comment already names), B/C unused.
	ActionDeviceLinked = 3
)

// Action is one entry in the log -- same four-int32 shape RefluxAction (reflux_runtime.h) uses.
type Action struct {
	ActionType int32
	A, B, C    int32
}

// logCapacity mirrors REFLUX_LOG_CAPACITY's own real ring-buffer sizing convention -- generous
// for a single-process terminal app's own lifetime, not tuned against any real measured need.
const logCapacity = 256

// Log is a real ring buffer, same total_dispatched-vs-capacity wraparound logic
// reflux_log_dispatch/reflux_log_at (reflux_runtime.c) already use -- ported field-for-field,
// not reinvented. Not safe for concurrent use without external locking (matching the C original,
// which is also not thread-safe -- SHANKPIT/ECOWAR both only ever call it from one game-tick
// goroutine/thread); IDUNA.GAME's own only dispatcher (the prompt goroutine) and reader (a future
// subscriber) should coordinate via Global's own methods below, which DO lock.
type Log struct {
	actions         [logCapacity]Action
	totalDispatched int
}

func (l *Log) Reset() { *l = Log{} }

func (l *Log) Dispatch(actionType, a, b, c int32) {
	slot := l.totalDispatched % logCapacity
	l.actions[slot] = Action{ActionType: actionType, A: a, B: b, C: c}
	l.totalDispatched++
}

func (l *Log) Length() int {
	if l.totalDispatched < logCapacity {
		return l.totalDispatched
	}
	return logCapacity
}

// At returns the action at the given real (not raw-slot) index, oldest-first, same real
// wraparound resolution reflux_log_at (reflux_runtime.c) uses. ok is false out of range.
func (l *Log) At(index int) (Action, bool) {
	size := l.Length()
	if index < 0 || index >= size {
		return Action{}, false
	}
	if l.totalDispatched <= logCapacity {
		return l.actions[index], true
	}
	oldestSlot := l.totalDispatched % logCapacity
	slot := (oldestSlot + index) % logCapacity
	return l.actions[slot], true
}

// Global is the one, real, per-process log every dispatcher/subscriber in this app shares --
// same "the one, real, per-server-process log" role g_reflux_log (reflux_runtime.c) plays for
// SHANKPIT. Safe for concurrent use (unlike Log's own bare methods above): the prompt goroutine
// dispatches from a background goroutine while the render loop could, in principle, poll from
// the main thread.
var (
	globalMu  sync.Mutex
	globalLog Log
)

func Dispatch(actionType, a, b, c int32) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLog.Dispatch(actionType, a, b, c)
}

func LogSize() int {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalLog.Length()
}

func ActionAt(index int) (Action, bool) {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalLog.At(index)
}
