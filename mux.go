package workers

import (
	"sync"
)

// WorkMux is a Beanstalkd Job multiplexer.
// It matches the tube of each incoming job against a list
// of registered tubes and calls the handler of that tube.
type WorkMux struct {
	mu sync.RWMutex
	m  map[string]muxEntry
}

type muxEntry struct {
	h    Handler
	tube string
}

// NewWorkMux allocates and returns a new WorkMux.
func NewWorkMux() *WorkMux {
	return &WorkMux{m: make(map[string]muxEntry)}
}

// Handle registers the job handler for the given tube.
// If a handler already exists for tube, Handle panics.
func (mux *WorkMux) Handle(tube string, handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if tube == "" {
		panic("invalid tube")
	}

	if handler == nil {
		panic("nil handler")
	}

	if _, found := mux.m[tube]; found {
		panic("multiple registrations for " + tube)
	}

	mux.m[tube] = muxEntry{
		h:    handler,
		tube: tube,
	}
}

// Handler returns the handler to use for the given job. If there is no
// registered handler that applies to the job, Handler returns nil.
func (mux *WorkMux) Handler(tube string) Handler {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	if handler, found := mux.m[tube]; found {
		return handler.h
	}

	return nil
}

// Tubes returns a list of tubes handled by the WorkMux.
func (mux *WorkMux) Tubes() []string {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	tubes := make([]string, len(mux.m))
	i := 0

	for k := range mux.m {
		tubes[i] = k
		i++
	}

	return tubes
}

// Work dispatches the job to the proper handler. Makes WorkMux Implements
// the Handler interface. Work panics if no handler is defined to handle the
// job.
func (mux WorkMux) Work(j *Job) {
	h := mux.Handler(j.Tube)

	if h == nil {
		panic("no handler for tube " + j.Tube)
	}

	h.Work(j)
}
