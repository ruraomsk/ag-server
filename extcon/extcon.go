package extcon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

//ExtContext расширенный контекст
type ExtContext struct {
	name       string
	ctx        context.Context
	cancelFunc context.CancelFunc
}

var id uint64
var mutexID sync.Mutex

var mutex sync.Mutex
var work bool

//Contexts собраны все контексты
var contexts map[string]*ExtContext

//NewContext создает новый расширенный контекст только с командой завершения
func NewContext(name string) (*ExtContext, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if !work {
		return nil, fmt.Errorf("stoped context system")
	}
	ec := new(ExtContext)
	ec.name = name
	ctx, cancel := context.WithCancel(context.Background())
	ec.cancelFunc = cancel
	ec.ctx = ctx
	contexts[name] = ec
	return ec, nil
}

//GetName return name context
func (ec *ExtContext) GetName() string {
	return ec.name
}

//Done return chan for cancel
func (ec *ExtContext) Done() <-chan struct{} {
	return ec.ctx.Done()
}

//BackgroundInit инициализируем
func BackgroundInit() {
	id = 0
	work = true
	contexts = make(map[string]*ExtContext, 0)
}

func allstop() {
	mutex.Lock()
	work = false
	for _, ec := range contexts {
		ec.cancelFunc()
	}
	mutex.Unlock()
	time.Sleep(10 * time.Second)
}

//SetTimerClock порождает канал где приходят по времения сообщения
func SetTimerClock(step time.Duration) chan int {
	timer := make(chan int)
	go func() {
		for true {
			time.Sleep(step)
			timer <- 1
		}
	}()
	return timer
}

//BackgroundWork обычно вызывается для обслуживания разного
// ПОсле выхода нет контекстов
func BackgroundWork(step time.Duration, stop chan int) {
	if !work {
		BackgroundInit()
	}
	timer := make(chan int)
	go func() {
		for true {
			time.Sleep(step)
			timer <- 1
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for true {
		select {
		case <-stop:
			{
				allstop()
				return
			}
		case <-c:
			{
				allstop()
				return
			}
		}
	}
}
