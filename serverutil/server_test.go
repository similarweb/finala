package serverutil_test

import (
	"context"
	"finala/serverutil"
	"testing"
	"time"
)

type ServeStruct struct {
	ctx      context.Context
	cancelFn context.CancelFunc
	Request  chan string
	isStop   bool
	init     bool
}

func MockServeStruct() *ServeStruct {
	ctx, cancelFn := context.WithCancel(context.Background())
	return &ServeStruct{
		ctx:      ctx,
		cancelFn: cancelFn,
		Request:  make(chan string),
		isStop:   false,
		init:     false,
	}

}

func (m *ServeStruct) Serve() serverutil.StopFunc {

	stopped := make(chan bool)
	go func() {
		m.init = true
		<-m.ctx.Done()
		m.isStop = true
		stopped <- true
	}()

	return func() {
		m.cancelFn()
		<-stopped
	}
}
func TestServe(t *testing.T) {

	serverStruct := MockServeStruct()

	if serverStruct.init {
		t.Fatalf("unexpected serve init, got %t expected %t", serverStruct.isStop, false)
	}

	runners := serverutil.RunAll(serverStruct)
	time.Sleep(time.Second)
	if !serverStruct.init {
		t.Fatalf("unexpected init serve funtion, got %t expected %t", serverStruct.isStop, true)
	}

	runners.StopFunc()

}
