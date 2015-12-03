package workers

// Handler defines a way for workers to handle jobs for a tube.
// Objects implementing the Handler interface can be registered to
// handle jobs for a particular tube.
type Handler interface {
	Work(*Job)
}

// HandlerFunc type is an adapter to allow the use of
// ordinary functions as Work handlers.  If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler object that calls f.
type HandlerFunc func(*Job)

// Work makes HandlerFunc implement the Handler interface.
func (f HandlerFunc) Work(j *Job) {
	f(j)
}
