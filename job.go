package workers

import (
	"strconv"
	"time"

	"github.com/kr/beanstalk"
)

// Job represents a job received by a worker.
type Job struct {
	ID   uint64
	Tube string
	Body []byte
	conn *beanstalk.Conn
}

// JobStats represents statistical information about a job.
type JobStats struct {
	Priority uint32
	Age      time.Duration
	TimeLeft time.Duration
}

// NewJob creates a Job.
func NewJob(conn *beanstalk.Conn, tube string, id uint64, body []byte) *Job {
	return &Job{
		ID:   id,
		Tube: tube,
		Body: body,
		conn: conn,
	}
}

// Delete deletes the current job.
// It removes the job from the server entirely.
func (j *Job) Delete() error {
	return j.conn.Delete(j.ID)
}

// Release releases the current job. Release puts the reserved job back
// into the ready queue (and marks its state as ready) to be run by any client.
func (j *Job) Release(pri uint32, delay time.Duration) error {
	return j.conn.Release(j.ID, pri, delay)
}

// Touch touches the current job. It allows the worker to request more
// time to work on the job.
func (j *Job) Touch() error {
	return j.conn.Touch(j.ID)
}

// Bury buries the current job. Bury puts the job into the "buried" state.
// Buried jobs are put into a FIFO linked list and will not be touched by
// the server again until a client kicks them manually.
func (j *Job) Bury(pri uint32) error {
	return j.conn.Bury(j.ID, pri)
}

// Stats gives statistical information about the current job.
func (j *Job) Stats() (*JobStats, error) {
	m, err := j.conn.StatsJob(j.ID)
	if err != nil {
		return nil, err
	}

	pri, err := strconv.Atoi(m["pri"])
	if err != nil {
		return nil, err
	}

	age, err := strconv.Atoi(m["age"])
	if err != nil {
		return nil, err
	}

	left, err := strconv.Atoi(m["time-left"])
	if err != nil {
		return nil, err
	}

	return &JobStats{
		Priority: uint32(pri),
		Age:      time.Duration(time.Duration(age) * time.Second),
		TimeLeft: time.Duration(time.Duration(left) * time.Second),
	}, nil
}
