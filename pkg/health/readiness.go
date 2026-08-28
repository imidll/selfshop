package health

import (
	"net/http"
	"sync"
	"sync/atomic"
)

type Checker interface {
	Named() string
	Ready() bool
}

type Readiness struct {
	closed   atomic.Bool
	mu       sync.RWMutex
	checkers []Checker
}

func New(chs ...Checker) *Readiness {
	r := new(Readiness)
	if len(chs) > 0 {
		r.checkers = make([]Checker, len(chs))
		copy(r.checkers, chs)
	}
	return r
}

func (r *Readiness) MarkNotReady() { r.closed.Store(true) }

func (r *Readiness) Register(chs ...Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers = append(r.checkers, chs...)
}

func (r *Readiness) ExistingCheckers() map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string]bool, len(r.checkers))
	for _, ch := range r.checkers {
		out[ch.Named()] = ch.Ready()
	}
	return out
}

func (r *Readiness) Ready() bool {
	if r.closed.Load() {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.checkers) == 0 {
		return false
	}
	for _, ch := range r.checkers {
		if !ch.Ready() {
			return false
		}
	}
	return true
}

func (r *Readiness) AliveHandler(
	w http.ResponseWriter, _ *http.Request,
) {
	w.WriteHeader(http.StatusOK)
}

func (r *Readiness) ReadyHandler(
	w http.ResponseWriter, _ *http.Request,
) {
	if r.Ready() {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}
