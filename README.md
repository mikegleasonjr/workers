# workers

[![Build Status](https://travis-ci.org/mikegleasonjr/workers.svg?branch=master)](https://travis-ci.org/mikegleasonjr/workers)

A simple beanstalk client library to consume jobs written in go. Heavily inspired from the standard `net/http` package.

## Install

```
$ go get github.com/mikegleasonjr/workers
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/mikegleasonjr/workers"
)

func main() {
	mux := workers.NewWorkMux()

	mux.Handle("tube1", workers.HandlerFunc(func(job *workers.Job) {
		fmt.Println("deleting job:", job.ID, job.Tube)
		job.Delete()
	}))

	mux.Handle("tube2", workers.HandlerFunc(func(job *workers.Job) {
		job.Bury(1000)
	}))

	workers.ConnectAndWork("tcp", "127.0.0.1:11300", mux)
}
```

Or if you would like to consume jobs only on the `default` tube:

```go
package main

import (
	"fmt"
	"github.com/mikegleasonjr/workers"
)

func main() {
	workers.ConnectAndWork("tcp", "127.0.0.1:11300", workers.HandlerFunc(func(job *workers.Job) {
		fmt.Println("deleting job:", job.ID, job.Tube)
		job.Delete()
	}))
}
```

## Job Handlers

Jobs are serviced each in their own goroutines. Jobs are handled in parallel as fast as they are reserved from the server.

You can handle jobs by providing an object implementing the `Handler` interface:

```go
type Handler interface {
	Work(*Job)
}
```
Or use the `HandlerFunc` adapter as seen in the examples above.

## Stopping workers

The client will disconnect itself from the beanstalk server and return upon receiving a `SIGINT` or a `SIGTERM` signal, waiting for current jobs to be handled.
