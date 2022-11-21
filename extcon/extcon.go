package extcon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

// ExtContext расширенный контекст
type ExtContext struct {
	name       string
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// var id uint64
// var mutexID sync.Mutex

var mutex sync.Mutex
var work bool

// Contexts собраны все контексты
var contexts map[string]*ExtContext

// NewContext создает новый расширенный контекст только с командой завершения
func NewContext(name string) (*ExtContext, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if !work {
		BackgroundInit()
	}
	ec := new(ExtContext)
	ec.name = name
	ctx, cancel := context.WithCancel(context.Background())
	ec.cancelFunc = cancel
	ec.ctx = ctx
	contexts[name] = ec
	return ec, nil
}

// GetName return name context
func (ec *ExtContext) GetName() string {
	return ec.name
}

// Done return chan for cancel
func (ec *ExtContext) Done() <-chan struct{} {
	return ec.ctx.Done()
}

// BackgroundInit инициализируем
func BackgroundInit() {
	// id = 0
	work = true
	contexts = make(map[string]*ExtContext, 0)
}
func StopForName(name string) {
	mutex.Lock()
	ec, is := contexts[name]
	if is {
		ec.cancelFunc()
	}
	mutex.Unlock()

}
func allstop() {
	mutex.Lock()
	work = false
	for _, ec := range contexts {
		ec.cancelFunc()
	}
	mutex.Unlock()
	time.Sleep(5 * time.Second)
}

// SetTimerClock порождает канал где приходят по времения сообщения
func SetTimerClock(step time.Duration) *time.Ticker {
	return time.NewTicker(step)
}

// BackgroundWork обычно вызывается для обслуживания разного
// ПОсле выхода нет контекстов
func BackgroundWork(step time.Duration, stop chan interface{}) {
	if !work {
		BackgroundInit()
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// ,
	// 	syscall.SIGQUIT,
	// 	syscall.SIGINT,
	// 	syscall.SIGTERM,
	// 	syscall.SIGHUP)
	for {
		select {
		case <-stop:
			{
				fmt.Println("Wait make abort...")
				allstop()
				return
			}
		case <-c:
			{
				fmt.Println("Wait make abort...")
				allstop()
				return
			}
		}
	}
}
