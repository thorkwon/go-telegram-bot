package queue

import (
	"container/list"
	"time"

	"github.com/thorkwon/go-telegram-bot/utils"
)

var log = utils.GetLogger(utils.GetPackageName())

func init() {
	// utils.EnableDebugLog(utils.GetPackageName())
}

type Task struct {
	chatID int64
	msgID  int
	cnt    int
	idx    int
}

type WorkQueue struct {
	queue      []list.List
	queueLen   int
	currentIdx int
	runFlag    bool
	haveTask   bool
	cb         func(int64, int)
	arg        interface{}
	doneQueue  chan bool
}

func NewWorkQueue(cb func(int64, int)) *WorkQueue {
	obj := &WorkQueue{}

	obj.queueLen = 10
	obj.queue = make([]list.List, obj.queueLen)
	obj.currentIdx = 0
	obj.runFlag = false
	obj.haveTask = false
	obj.doneQueue = make(chan bool)

	obj.setCbFunc(cb)

	return obj
}

func (q *WorkQueue) setCbFunc(cb func(int64, int)) {
	q.cb = cb
}

func (q *WorkQueue) AddTask(chatID int64, msgID int, delay int) {
	n := q.currentIdx + delay
	m := int(n / q.queueLen)
	s := int(n % q.queueLen)
	if s == 0 || s <= q.currentIdx {
		m--
	}

	log.Debugf("=====>> Add Task msgID [%d] delay [%d]", msgID, delay)
	task := &Task{chatID: chatID, msgID: msgID, cnt: m, idx: s}
	q.queue[s].PushBack(task)

	q.haveTask = true
	if !q.runFlag {
		q.runFlag = true
		go q.start()
	}
}

func (q *WorkQueue) start() {
	log.Debug("Work Queue start")
	totalCnt := 0
	for q.runFlag {
		time.Sleep(time.Second)
		q.currentIdx++
		totalCnt++

		if q.currentIdx == q.queueLen {
			q.currentIdx = 0
			if !q.haveTask {
				q.runFlag = false
			}
			q.haveTask = false
		}

		// Must be use pointer (*list.List)
		tasks := &q.queue[q.currentIdx]
		if tasks.Len() != 0 {
			log.Debugf("have tasks %d", tasks.Len())

			task := tasks.Front()
			for task != nil {
				data := task.Value.(*Task)
				log.Debugf("task msgID [%d]\t idx %d : delete cnt %d", data.msgID, data.idx, data.cnt)

				if data.cnt == 0 {
					// remove msg, call cb func
					q.cb(data.chatID, data.msgID)

					log.Debugf("Delete task msgID [%d] in list", data.msgID)
					// delete task
					if task.Next() == nil {
						tasks.Remove(task)
						task = nil
					} else {
						task = task.Next()
						tasks.Remove(task.Prev())
					}
					continue
				}

				q.haveTask = true
				data.cnt--
				task = task.Next()
			}
		}

		// log.Debugf("ccurrentIdx [%d] haveTask [%v] ========= totalCnt [%d]", q.currentIdx, q.haveTask, totalCnt)
	}
	log.Debug("Work Queue finished")
	q.doneQueue <- true
}

func (q *WorkQueue) Stop() {
	if q.runFlag {
		q.runFlag = false
		<-q.doneQueue
	}
}

/*
// Unit test code
func cbTest(a int64, b int) {
	log.Debugf("Call cbTest chatID %d, msgID %d", a, b)
}

func UnitTest() {
	qu := NewWorkQueue(cbTest)

	qu.AddTask(111, 100, 13)
	qu.AddTask(111, 101, 3)

	log.Debug("main wait 15 ===========>")
	time.Sleep(15 * time.Second)

	log.Debug("main add task ==========> 111 102 8")
	qu.AddTask(111, 102, 8)

	qu.AddTask(111, 103, 8)
	qu.AddTask(111, 104, 2)
	time.Sleep(5 * time.Second)
	qu.AddTask(111, 105, 60)
	qu.AddTask(111, 106, 2)
	qu.AddTask(111, 107, 2)
	time.Sleep(10 * time.Second)
}
*/
