package pool

import (
	"io/ioutil"
	"log"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	// 设定使用最大的cpu核数
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// 测试worker创建
func TestNewWorker(t *testing.T) {
	pool := make(chan *worker)
	worker := newWorker(pool)
	assert.NotNil(t, worker, "worker not nil")
}

func TestWorkerStart(t *testing.T) {
	pool := make(chan *worker)
	worker := newWorker(pool)
	worker.start()
	worker = <-pool
	assert.NotNil(t, worker, "worker show regitser to pool")
}

func TestWokerWorking(t *testing.T) {
	pool := make(chan *worker)
	worker := newWorker(pool)
	worker.start()
	// 做这个是为了确保worker start已经启动
	worker = <-pool
	test := false
	// 等待函数执行完
	done := make(chan bool)
	job := func() {
		test = true
		done <- true
	}
	worker.jobChannel <- job
	<-done
	assert.Equal(t, test, true)
}

func TestNewPool(t *testing.T) {
	pool := NewPool(1000, 50)
	defer pool.Release()
	assert.NotNil(t, pool, "new pool should success")
}

func TestPoolWaitAll(t *testing.T) {
	pool := NewPool(1000, 10000)
	defer pool.Release()
	workCount := 100
	pool.WaitCount(workCount)
	var counter uint64 = 0
	for i := 0; i < workCount; i++ {
		job := func() {
			// 因为是多worker操作，所以必须用atomic原子操作（类似于lock）才能保证值累加是正确的
			atomic.AddUint64(&counter, uint64(1))
			defer pool.JobDone()
		}
		pool.JobQueue <- job
	}
	// 等待任务完成
	pool.WaitAll()
	assert.Equal(t, counter, uint64(workCount))
}

func TestPoolRelease(t *testing.T) {
	// 正在执行和排队的携程数，记录此时还没有工作时的状态
	runGoroutine := runtime.NumGoroutine()
	pool := NewPool(5, 10)
	defer func() {
		pool.Release()
		// 由于release会等待所有携程结束，所以release结束后，应该恢复到原本的状态
		assert.Equal(t, runtime.NumGoroutine(), runGoroutine, "All goroutine released")
	}()
	pool.WaitCount(10000)
	for i := 0; i < 10000; i++ {
		job := func() {
			defer pool.JobDone()
		}
		// 在这里执行的时候go协程数会拉满，10
		pool.JobQueue <- job
	}
	pool.WaitAll()
}

func Benchmark(b *testing.B) {
	// 使用单线程测试
	pool := NewPool(1, 10)
	defer pool.Release()
	// 不让log显示在终端（输出到dev/Null）
	log.SetOutput(ioutil.Discard)
	// 重置计数器，避免前置代码干扰
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.JobQueue <- func() {
			log.Printf("I am worker! Number %d\n", i)
		}
	}
}
