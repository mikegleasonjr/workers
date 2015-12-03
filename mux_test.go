package workers

import (
	"testing"
)

// WorkMux is defined and can be instanciated
func TestNewWorkMux(t *testing.T) {
	var mux *WorkMux
	mux = NewWorkMux()
	if mux == nil {
		t.Fail()
	}
}

func TestHandleInvalidTube(t *testing.T) {
	defer func() {
		if err := recover(); err == nil || err != "invalid tube" {
			t.Fail()
		}
	}()
	mux := NewWorkMux()
	mux.Handle("", HandlerFunc(func(*Job) {}))
}

func TestHandleNilHandler(t *testing.T) {
	defer func() {
		if err := recover(); err == nil || err != "nil handler" {
			t.Fail()
		}
	}()
	mux := NewWorkMux()
	mux.Handle("tube1", nil)
}

func TestHandleDuplicate(t *testing.T) {
	defer func() {
		if err := recover(); err == nil || err != "multiple registrations for tube1" {
			t.Fail()
		}
	}()
	mux := NewWorkMux()
	mux.Handle("tube1", HandlerFunc(func(*Job) {}))
	mux.Handle("tube1", HandlerFunc(func(*Job) {}))
}

// Checks for unknown registered handler
func TestHandlerUnknown(t *testing.T) {
	mux := NewWorkMux()
	if mux.Handler("tube1") != nil {
		t.Fail()
	}
}

// Checks for a known registered handler
func TestHandlerRegistered(t *testing.T) {
	mux := NewWorkMux()
	handled := false
	mux.Handle("tube1", HandlerFunc(func(*Job) { handled = true }))
	mux.Handler("tube1").Work(nil)
	if !handled {
		t.Fail()
	}
}

func TestEmptyTubes(t *testing.T) {
	mux := NewWorkMux()
	if len(mux.Tubes()) != 0 {
		t.Fail()
	}
}

func TestTubes(t *testing.T) {
	contains := func(s []string, e string) bool {
		for _, a := range s {
			if a == e {
				return true
			}
		}
		return false
	}

	mux := NewWorkMux()
	mux.Handle("tube1", HandlerFunc(func(*Job) {}))
	mux.Handle("tube2", HandlerFunc(func(*Job) {}))
	mux.Handle("tubeN", HandlerFunc(func(*Job) {}))
	tubes := mux.Tubes()
	if len(tubes) != 3 || !contains(tubes, "tube1") || !contains(tubes, "tube2") || !contains(tubes, "tubeN") {
		t.Fail()
	}
}

func TestWork(t *testing.T) {
	handled := false
	mux := NewWorkMux()
	mux.Handle("tube1", HandlerFunc(func(*Job) {}))
	mux.Handle("tube2", HandlerFunc(func(*Job) { handled = true }))
	mux.Handle("tubeN", HandlerFunc(func(*Job) {}))
	mux.Work(&Job{Tube: "tube2"})
	if !handled {
		t.Fail()
	}
}

func TestWorkUnhandledTube(t *testing.T) {
	defer func() {
		if err := recover(); err == nil || err != "no handler for tube tubeX" {
			t.Fail()
		}
	}()
	mux := NewWorkMux()
	mux.Work(&Job{Tube: "tubeX"})
}
