package workers

import (
	"errors"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kr/beanstalk"
)

// ErrClientHasQuit is returned by Client when it is quitting
var ErrClientHasQuit = errors.New("client has quit")

// Client defines parameters for running an beanstalk client.
type Client struct {
	Network string
	Addr    string
	Handler Handler
	mu      sync.Mutex // guards stop
	stop    chan error
}

// ConnectAndWork connects on the c.Network and c.Addr and then
// calls Reserve to handle jobs on the beanstalk instance.
func (c *Client) ConnectAndWork() error {
	conn, err := net.Dial(c.Network, c.Addr)

	if err != nil {
		return err
	}

	return c.Reserve(conn)
}

// ConnectAndWork creates a client, connects to the beanstalk instance and
// reserves jobs to be processed by Handler.
func ConnectAndWork(network string, addr string, handler Handler) error {
	client := &Client{Network: network, Addr: addr, Handler: handler}
	return client.ConnectAndWork()
}

// Reserve accepts incoming jobs on the beanstalk.Conn conn, creating a
// new service goroutine for each. The service goroutines read the job and
// then call c.Handler to process them.
func (c *Client) Reserve(conn io.ReadWriteCloser) error {
	c.mu.Lock()
	c.stop = make(chan error)
	c.mu.Unlock()
	bs := beanstalk.NewConn(conn)
	tubes := c.tubes(bs)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go c.quitOnSignal(wg)

	defer bs.Close()
	defer wg.Wait()

	for {
		wait := time.Second // how long to sleep when no jobs in queues

		for name, tube := range tubes {
			id, body, err := tube.Reserve(0 /* don't block others */)
			if err == nil {
				wait = 0 // drain the queue as fast as possible
				wg.Add(1)
				go c.work(wg, NewJob(bs, name, id, body))
			} else if !isTimeoutOrDeadline(err) {
				c.Stop()
				return err
			}
			select {
			case <-c.stop:
				return ErrClientHasQuit
			default:
			}
		}

		select {
		case <-c.stop:
			return ErrClientHasQuit
		case <-time.After(wait):
		}
	}
}

// Stop stops reserving jobs and wait for current workers to finish their job.
func (c *Client) Stop() {
	c.mu.Lock()
	close(c.stop)
	c.mu.Unlock()
}

func (c *Client) tubes(conn *beanstalk.Conn) map[string]*beanstalk.TubeSet {
	names := []string{"default"}

	if mux, isMux := c.Handler.(*WorkMux); isMux {
		names = mux.Tubes()
	}

	tubes := make(map[string]*beanstalk.TubeSet, len(names))
	for _, name := range names {
		tubes[name] = beanstalk.NewTubeSet(conn, name)
	}

	return tubes
}

func (c *Client) work(wg *sync.WaitGroup, j *Job) {
	defer wg.Done()
	c.Handler.Work(j)
}

func (c *Client) quitOnSignal(wg *sync.WaitGroup) {
	defer wg.Done()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-c.stop:
	case <-sigchan:
		c.Stop()
	}
}

func isTimeoutOrDeadline(err error) bool {
	if connerr, isConnErr := err.(beanstalk.ConnError); isConnErr {
		return connerr.Op == "reserve-with-timeout" &&
			(connerr.Err == beanstalk.ErrTimeout || connerr.Err == beanstalk.ErrDeadline)
	}

	return false
}
