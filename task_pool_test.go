package task_pool

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}
func Test_DefaultConfig(t *testing.T) {
	pool := NewTaskPool(DefaultConfig)
	wg := sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		f := func() error {
			fmt.Println("task ...", goid())
			time.Sleep(time.Second)
			wg.Done()
			return nil
		}
		task := &Task{Func: f}
		if err := pool.PushTask(task); err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
}
func Test_TaskPool(t *testing.T) {
	conf := &PoolConfig{
		TaskBufferSize:      10,
		MaxTaskNum:          3,
		SubThreadMinTaskNum: 1,
		StartThreadNum:      1,
		MaxThreadNum:        50,
		ThreadGrowthIntvl:   1,
	}

	pool := NewTaskPool(conf)
	wg := sync.WaitGroup{}

	for i := 0; i < 40; i++ {
		wg.Add(1)
		f := func() error {
			fmt.Println("task ...", goid())
			time.Sleep(time.Second)
			wg.Done()
			return nil
		}
		task := &Task{Func: f}
		pool.PushTask(task)
	}

	fmt.Println("put all task ok")

	time.Sleep(40 * time.Second)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		f := func() error {
			fmt.Println("task ...", goid())
			time.Sleep(time.Second)
			wg.Done()
			return nil
		}
		task := &Task{Func: f}
		pool.PushTask(task)
	}
	wg.Wait()
}
