# pool


轻量级 Goroutine 线程池

learn by https://github.com/ivpusic/grpool


## Docs
https://github.com/github.com/gamefu/pool

## Installation
```
go get github.com/gamefu/pool
```

## Simple example
```Go
package main

import (
  "fmt"
  "runtime"
  "time"

  "github.com/gamefu/pool"
)

func main() {
  // 第一个参数为worker数量，第二个为job队列长度
  pool := grpool.NewPool(100, 50)

  // 任务执行完后释放所有资源
  defer pool.Release()

  // 提交任务到队列
  for i := 0; i < 10; i++ {
    count := i

    pool.JobQueue <- func() {
      fmt.Printf("I am worker! Number %d\n", count)
    }
  }

  // 模拟等待任务执行完
  time.Sleep(1 * time.Second)
}
```

## Example with waiting jobs to finish
```Go
package main

import (
  "fmt"
  "runtime"

  "github.com/ivpusic/grpool"
)

func main() {
  pool := grpool.NewPool(100, 50)
  defer pool.Release()

  // 需要执行多少次任务才会结束
  pool.WaitCount(10)

  for i := 0; i < 10; i++ {
    count := i

    pool.JobQueue <- func() {
      // 执行完一次任务
      defer pool.JobDone()

      fmt.Printf("hello %d\n", count)
    }
  }

  // 任务执行完之前会一直等待挂起
  pool.WaitAll()
}
```

## License
*MIT*