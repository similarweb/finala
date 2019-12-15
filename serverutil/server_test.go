package serverutil_test

import (
	"context"
	"finala/serverutil"
	"testing"
	"time"
)

type ServeStruct struct {
	Request chan string
	isStop  bool
	init    bool
}

func MockServeStruct() *ServeStruct {
	return &ServeStruct{
		Request: make(chan string),
		isStop:  false,
		init:    false,
	}

}

func (m *ServeStruct) Serve() serverutil.StopFunc {

	ctx, cancelFn := context.WithCancel(context.Background())
	stopped := make(chan bool)
	go func() {
		m.init = true
		for {
			select {
			case <-ctx.Done():
				m.isStop = true
				stopped <- true
				return
			}
		}
	}()

	return func() {
		cancelFn()
		<-stopped
	}
}
func TestServe(t *testing.T) {

	serverStruct := MockServeStruct()

	if serverStruct.init {
		t.Fatalf("unexpected serve init, got %t expected %t", serverStruct.isStop, false)
	}

	serverutil.RunAll(serverStruct)
	time.Sleep(time.Second)
	if !serverStruct.init {
		t.Fatalf("unexpected init serve funtion, got %t expected %t", serverStruct.isStop, true)
	}

}
