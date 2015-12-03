package workers

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/kr/beanstalk"
)

func Example() {
	mux := NewWorkMux()

	mux.Handle("tube1", HandlerFunc(func(job *Job) {
		fmt.Printf("processing job %d with content %v\n", job.ID, job.Body)
		job.Delete()
	}))

	mux.Handle("tube2", HandlerFunc(func(job *Job) {
		job.Release(0, 0)
	}))

	ConnectAndWork("tcp", "localhost:11300", mux)
}

func TestStopClient(t *testing.T) {
	client := &Client{
		Network: "tcp",
		Addr:    "localhost:11300",
		Handler: HandlerFunc(func(job *Job) {
		}),
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		client.Stop()
	}()

	err := client.ConnectAndWork()
	if err != ErrClientHasQuit {
		t.Fail()
	}
}

func TestClientStopsOnSIGTERM(t *testing.T) {
	go func() {
		time.Sleep(100 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	err := ConnectAndWork("tcp", "localhost:11300", HandlerFunc(func(job *Job) {}))
	if err != ErrClientHasQuit {
		t.Fail()
	}
}

func TestClientStopsOnSIGINT(t *testing.T) {
	go func() {
		time.Sleep(100 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	err := ConnectAndWork("tcp", "localhost:11300", HandlerFunc(func(job *Job) {}))
	if err != ErrClientHasQuit {
		t.Fail()
	}
}

func TestReserveIsParallelAndWaits(t *testing.T) {
	count := int32(0)
	tubeName := strconv.Itoa(int(time.Now().Unix()))
	start := time.Now()

	mux := NewWorkMux()
	mux.Handle(tubeName, HandlerFunc(func(job *Job) {
		time.Sleep(time.Second)
		atomic.AddInt32(&count, 1)
		job.Delete()
	}))

	go func() {
		conn, _ := beanstalk.Dial("tcp", "localhost:11300")
		tube := &beanstalk.Tube{Conn: conn, Name: tubeName}
		tube.Put([]byte("job1"), 0, 0, time.Minute)
		tube.Put([]byte("job2"), 0, 0, time.Minute)
		tube.Put([]byte("job3"), 0, 0, time.Minute)
		tube.Put([]byte("job4"), 0, 0, time.Minute)
		tube.Put([]byte("job5"), 0, 0, time.Minute)
		time.Sleep(time.Millisecond * 1100)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	ConnectAndWork("tcp", "localhost:11300", mux)

	if count != 5 || time.Since(start) > time.Duration(time.Millisecond*2200) {
		t.Fail()
	}
}
