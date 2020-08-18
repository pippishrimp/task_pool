package task_pool

import (
	"errors"
	"sync/atomic"
	"time"
)

var DefaultConfig = &PoolConfig{
	TaskBufferSize:      1024,
	MaxTaskNum:          896, // 总长度的7/8
	SubThreadMinTaskNum: 0,
	StartThreadNum:      1,
	MaxThreadNum:        50,
	ThreadGrowthIntvl:   1,
	TimeOut:             1,
}

type PoolConfig struct {
	TaskBufferSize      int // 任务channel长度
	MaxTaskNum          int // 启动协程处理最大任务数
	SubThreadMinTaskNum int // 当任务数小于等于该值，子协程退出
	StartThreadNum      int // 启动协程数
	MaxThreadNum        int // 最大协程数
	ThreadGrowthIntvl   int // 协程增长时间间隔，单位秒
	TimeOut             int // push 超时时间, 单位s
}

type TaskFunc func() error
type Task struct {
	Func TaskFunc
}

type TaskPool struct {
	buffer             chan *Task
	conf               *PoolConfig
	currThreadNum      int32
	lastGrowThreadTime int64
}

func NewTaskPool(conf *PoolConfig) *TaskPool {
	if conf == nil {
		conf = DefaultConfig
	}

	pool := &TaskPool{
		conf:   conf,
		buffer: make(chan *Task, conf.TaskBufferSize),
	}

	pool.correcConfig()

	for i := 0; i < conf.StartThreadNum; i++ {
		go pool.process(false)
	}

	return pool
}

func (t *TaskPool) PushTask(task *Task) error {
	if task != nil {
		return t.pushTask(task)
	}

	return errors.New("taks is nil")
}

func (t *TaskPool) Close() {
	close(t.buffer)
}

func (t *TaskPool) pushTask(task *Task) error {
	timeout := time.After(time.Duration(t.conf.TimeOut) * time.Second)
	select {
	case t.buffer <- task:
		currBufferLen := len(t.buffer)
		currThreadNum := atomic.LoadInt32(&t.currThreadNum)
		threadGrowthIntvl := int(time.Now().Unix() - t.lastGrowThreadTime)
		if currBufferLen > t.conf.MaxTaskNum && int(currThreadNum) < t.conf.MaxThreadNum && threadGrowthIntvl > t.conf.ThreadGrowthIntvl {
			go t.process(true)
			atomic.AddInt32(&t.currThreadNum, 1)
			t.lastGrowThreadTime = time.Now().Unix()
		}
	case <-timeout:
		return errors.New("push timeout")
	}
	return nil
}

// 纠正配置参数，对不合理的参数修改为默认配置
func (t *TaskPool) correcConfig() {
	if t.conf.TaskBufferSize == 0 {
		t.conf.TaskBufferSize = DefaultConfig.TaskBufferSize
	}
	if t.conf.MaxTaskNum == 0 {
		t.conf.MaxTaskNum = DefaultConfig.MaxTaskNum
	}

	if t.conf.TaskBufferSize <= t.conf.MaxTaskNum {
		t.conf.TaskBufferSize = DefaultConfig.TaskBufferSize
		t.conf.MaxTaskNum = DefaultConfig.MaxTaskNum
	}

	if t.conf.StartThreadNum == 0 {
		t.conf.StartThreadNum = DefaultConfig.StartThreadNum
	}

	if t.conf.MaxThreadNum == 0 {
		t.conf.MaxThreadNum = DefaultConfig.MaxThreadNum
	}

	if t.conf.StartThreadNum >= t.conf.MaxThreadNum {
		t.conf.StartThreadNum = DefaultConfig.StartThreadNum
		t.conf.MaxThreadNum = DefaultConfig.MaxThreadNum
	}

	if t.conf.ThreadGrowthIntvl == 0 {
		t.conf.ThreadGrowthIntvl = DefaultConfig.ThreadGrowthIntvl
	}

	if t.conf.TimeOut == 0 {
		t.conf.TimeOut = DefaultConfig.TimeOut
	}
}

func (t *TaskPool) process(subProcess bool) {
	for {
		task := <-t.buffer
		task.Func()

		// 如果是子协程，buffer为空就退出
		if subProcess && len(t.buffer) <= t.conf.SubThreadMinTaskNum {
			atomic.AddInt32(&t.currThreadNum, -1)
			return
		}
	}
}
