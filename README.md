# 一个简单的动态增长协程池

```
type TaskFunc func() error
type Task struct {
	Func TaskFunc
}
```

其中task为任务执行函数，把task put协程池的buffer中即可，异步执行任务。比较简单，暂时不支持返回执行结果。

如果池中buffer中等待执行的task超过一定量，则会新起一个协程来执行任务，当buffer为空时，子协程则会退出。

## example

import "github.com/pippishrimp/task_pool"

```
pool := NewTaskPool(DefaultConfig)
f := func() error {
	fmt.Println("task ...")
	time.Sleep(time.Second)
	return nil
}
task := &Task{Func: f}
pool.PushTask(task)
```

### 参数描述

```
type PoolConfig struct {
	TaskBufferSize      int // 任务channel长度
	MaxTaskNum          int // 启动协程处理最大任务数
	SubThreadMinTaskNum int // 当任务数小于等于该值，子协程退出
	StartThreadNum      int // 启动协程数
	MaxThreadNum        int // 最大协程数
	ThreadGrowthIntvl   int // 协程增长时间间隔，单位秒
}

```

### 默认参数

```
var DefaultConfig = &PoolConfig{
	TaskBufferSize:      1024,
	MaxTaskNum:          896, // 总长度的7/8
	SubThreadMinTaskNum: 0,
	StartThreadNum:      1,
	MaxThreadNum:        50,
	ThreadGrowthIntvl:   1,
}
```

