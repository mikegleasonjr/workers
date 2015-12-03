package workers

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kr/beanstalk"
)

// Job is defined and can be instanciated
func TestNewJob(t *testing.T) {
	var job *Job
	job = NewJob(&beanstalk.Conn{}, "tube1", 123, []byte{})
	if job == nil {
		t.Fail()
	}
}

func TestJobCommands(t *testing.T) {
	conn := getTestConn(t)
	defer conn.Close()
	var job *Job

	// Delete
	job = withRandomJob(t, conn)
	if job.Delete() != nil || jobExists(t, conn, job.ID) {
		t.Fail()
	}

	// Release
	job = withRandomReservedJob(t, conn)
	job.Release(0, 0)
	stats, err := conn.StatsJob(job.ID)
	if err != nil || stats["state"] != "ready" {
		t.Fail()
	}

	// Touch
	job = withRandomReservedJob(t, conn)
	statsBefore, err := conn.StatsJob(job.ID)
	time.Sleep(time.Second)
	job.Touch()
	stats, err = conn.StatsJob(job.ID)
	timeLeftBefore, _ := strconv.Atoi(statsBefore["time-left"])
	timeLeft, _ := strconv.Atoi(stats["time-left"])
	if err != nil || timeLeft < timeLeftBefore {
		t.Fail()
	}

	// Bury
	job = withRandomReservedJob(t, conn)
	job.Bury(0)
	stats, err = conn.StatsJob(job.ID)
	if err != nil || stats["state"] != "buried" {
		t.Fail()
	}
}

func TestJobStats(t *testing.T) {
	conn := getTestConn(t)
	defer conn.Close()

	job := withRandomReservedJob(t, conn)
	stats, err := job.Stats()
	if err != nil {
		t.Fail()
	}
	statsOrigi, err := conn.StatsJob(job.ID)
	if err != nil ||
		strconv.Itoa(int(stats.Age.Seconds())) != statsOrigi["age"] ||
		strconv.Itoa(int(stats.TimeLeft.Seconds())) != statsOrigi["time-left"] ||
		strconv.Itoa(int(stats.Priority)) != statsOrigi["pri"] {
		t.Fail()
	}
}

func getTestConn(t *testing.T) *beanstalk.Conn {
	conn, err := beanstalk.Dial("tcp", "localhost:11300")
	if err != nil {
		t.Fail()
	}
	return conn
}

func withRandomJob(t *testing.T, conn *beanstalk.Conn) *Job {
	id, err := conn.Put([]byte{}, 0, 0, time.Minute*5)
	if err != nil {
		t.Fail()
	}
	return NewJob(conn, "default", id, []byte{})
}

func withRandomReservedJob(t *testing.T, conn *beanstalk.Conn) *Job {
	withRandomJob(t, conn)
	id, body, err := conn.Reserve(0)
	if err != nil {
		t.Fail()
	}
	return NewJob(conn, "default", id, body)
}

func jobExists(t *testing.T, conn *beanstalk.Conn, id uint64) bool {
	_, err := conn.Peek(id)
	return err == nil || !strings.HasSuffix(err.Error(), "not found")
}
