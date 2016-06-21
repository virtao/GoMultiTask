package examples

import (
	"sync"
	"errors"
)

type MultiTaskCallback func(interface{})interface{}

type MultiTask struct {
	callbackFunc MultiTaskCallback
	in chan interface{}
	out chan interface{}
	closeRoutines chan bool
	taskClosed chan bool
	routineNum int
	wg sync.WaitGroup
	isStart bool
	isStartLock sync.Mutex
}

func NewMultiTask(routineNum int, callbackFunc MultiTaskCallback) (ret *MultiTask) {
	ret = &MultiTask{
		callbackFunc: callbackFunc,
		routineNum: routineNum,
	}

	return ret
}

func (this *MultiTask) Start() {
	this.isStartLock.Lock()
	defer this.isStartLock.Unlock()
	if this.isStart {
		return
	}

	this.closeRoutines = make(chan bool)
	this.taskClosed = make(chan bool)
	this.in = make(chan interface{})
	this.out = make(chan interface{})

	this.wg.Add(this.routineNum)
	for i := 0; i < this.routineNum; i++ {
		go this.handle()
	}

	this.isStart = true
}

func (this *MultiTask) Close() {
	this.isStartLock.Lock()
	defer this.isStartLock.Unlock()

	if !this.isStart {
		return
	}
	close(this.closeRoutines)
	this.wg.Wait()
	close(this.taskClosed)
	close(this.in)
	// 这里先不要关闭out通道，因为一般情况下主线程会同时监听out和taskClosed两个通道，
	// 如果这里关闭out，那么主线程有可能监听不到taskClosed关闭的消息，
	// 反而监听到out关闭的消息（此时out通道会传递空值，代表自己已关闭），导致逻辑错误。
	//close(this.out)
	this.isStart = false
}

func (this *MultiTask) Add(data interface{}) (err error) {
	this.isStartLock.Lock()
	defer this.isStartLock.Unlock()
	if this.isStart {
		this.in <-data
	} else {
		err = errors.New("任务处理已关闭，需要重新启动。")
	}
	return err
}

func (this *MultiTask) GetDoneChan() <-chan bool {
	return this.taskClosed
}

func (this *MultiTask) GetResultChan() <-chan interface{} {
	return this.out
}

func (this *MultiTask) handle() {
	for {
		select {
		case inData := <-this.in:
			this.out <- this.callbackFunc(inData)
		case <-this.closeRoutines:
			this.wg.Done()
			return
		}
	}


}
