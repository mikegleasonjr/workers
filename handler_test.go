package workers

import "testing"

// Handler interface is defined and has a Work method
func TestHandlerInterface(t *testing.T) {
	var h Handler
	h = &testHandler{}
	h.Work(nil)
}

// HandlerFunc is defined and has a Work method that calls the
// function passed to HandlerFunc
func TestHandlerFunc(t *testing.T) {
	var h Handler
	called := false

	h = HandlerFunc(func(*Job) {
		called = true
	})

	h.Work(nil)

	if !called {
		t.Fail()
	}
}

type testHandler struct{}

func (h *testHandler) Work(*Job) {}
