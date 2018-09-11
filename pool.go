package pool

// 定义别名
type Job func()

// work
type worker struct {
	workerPool chan *worker
	jobChannel chan Job
	stop       chan bool
}

func newWorker(workerPool chan *worker) *worker {
	worker := &worker{
		workerPool: workerPool,
		// 用来接收传过来的job， 一个协程一次只能处理一个job
		jobChannel: make(chan Job),
		stop:       make(chan bool),
	}
	return worker
}

// worker 启动
func (w *worker) start() {
	var job Job
	go func() {
		for {
			// worker有空闲，写进线程池等待调度
			w.workerPool <- w
			// 阻塞等待任务分配
			select {
			case job = <-w.jobChannel:
				job()
			case <-w.stop:
				// 为了确保这个worker关掉后才会关掉下一个
				w.stop <- true
				return
			}
		}
	}()
}

// 调度器
type dispatcher struct {
	workerPool chan *worker
	jobQueue   chan Job
	stop       chan bool
}

// 初始化调度器
func newDispatcher(workerPool chan *worker, jobQueue chan Job) *dispatcher {
	dispatcher := &dispatcher{
		workerPool: workerPool,
		jobQueue:   jobQueue,
		stop:       make(chan bool),
	}
	// 根据worker数量声明worker
	for i := 0; i < cap(workerPool); i++ {
		worker := newWorker(workerPool)
		worker.start()
	}
	return dispatcher
}

// 调度
func (d *dispatcher) dispatch() {
	for {
		// 等待任务进行调度
		select {
		case job := <-d.jobQueue:
			worker := <-d.workerPool
			worker.jobChannel <- job
		case <-d.stop:
			// 关闭所有worker
			for i := 0; i < cap(d.workerPool); i++ {
				// 将worker从线程池弹出
				worker := <-d.workerPool
				// 给worker发送stop指令
				worker.stop <- true
				// 阻塞等待确保上一个worker已经关闭
				<-worker.stop
			}
			d.stop <- true
			return
		}
	}
}

type Pool struct {
	// 工作队列
	JobQueue chan Job
	// 调度器
	dispatcher *dispatcher
}

// 初始化线程池
func NewPool(workerNum int, jobQueueLen int) (pool *Pool) {
	workerPool := make(chan *worker, workerNum)
	jobQueue := make(chan Job, jobQueueLen)
	pool = &Pool{
		JobQueue:   jobQueue,
		dispatcher: newDispatcher(workerPool, jobQueue),
	}
	return
}

// 释放线程池所有资源
func (p *Pool) Release() {
	p.dispatcher.stop <- true
	// 同样为了确保dispatcher资源已经全部释放
	<-p.dispatcher.stop
}
