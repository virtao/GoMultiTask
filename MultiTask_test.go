package GoMultiTask

import (
	"fmt"
	"runtime"
	"testing"
)

type Context struct {
	Number int
	Str string
}

func TestMultiTask(t *testing.T) {
	testTask := NewMultiTask(1000, callback)
	testTask.Start()

	rFunc := func(name string) {
		const TASK_NUM int = 10000
		for i := 0; i < TASK_NUM; i++ {
			err := testTask.Add(&Context{
				Number: i,
				Str: name,
			})
			if err != nil {
				t.Logf("I am %s, MultiTask add failed, %d task(s) remaining, I will exit.\n", name, TASK_NUM - i)
				return
			}
		}
		testTask.Close()
	}

	go rFunc("Task1")
	go rFunc("Task2")
	go rFunc("Task3")
	go rFunc("Task4")

	FOR:
	for {
		select {
		case res := <-testTask.GetResultChan():
			val := res.(*Context)
			t.Logf("Message : %s\n", val.Str)
		case <-testTask.GetDoneChan():
			t.Logf("All tasks finished.")
			break FOR
		}
	}

	t.Logf("Goroutine number now : %d\n", runtime.NumGoroutine())
}

func callback(in interface{}) (out interface{}) {
	if val, ok := in.(*Context); ok {
		val.Str = fmt.Sprintf("This is %s's %d goroutine.", val.Str, val.Number)
		out = val
	} else {
		val.Str = fmt.Sprintf("Error at %s's %d goroutine.", val.Str, val.Number)
		out = val
	}
	return out
}
