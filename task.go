package hio

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	taskPool = &sync.Pool{
		New: func() interface{} {
			return &Task{}
		},
	}
	taskNodePool = &sync.Pool{
		New: func() interface{} {
			return &taskNode{}
		},
	}
)

type Task struct {
	fn  func(arg interface{}) error
	arg interface{}
}

func (t *Task) Release() {
	t.fn = nil
	t.arg = nil
	taskPool.Put(t)
}

type taskNode struct {
	next unsafe.Pointer
	task *Task
}

func (t *taskNode) Release() {
	t.task = nil
	t.next = unsafe.Pointer(nil)

	taskNodePool.Put(t)
}

type taskQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

func (t *taskQueue) Enqueue(task *Task) {
	node := newTaskNode(task)

	for {
		tail := t.load(&t.tail)
		next := t.load(&tail.next)

		if tail == t.load(&t.tail) {
			if next == nil {
				if atomic.CompareAndSwapPointer(&tail.next, unsafe.Pointer(next), unsafe.Pointer(node)) {
					atomic.CompareAndSwapPointer(&t.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
					break
				}
			} else {
				atomic.CompareAndSwapPointer(&t.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
			}
		}
	}
}

func (t *taskQueue) Dequeue() (*Task, bool) {
	var res *Task

	for {
		head := t.load(&t.head)
		tail := t.load(&t.tail)
		next := t.load(&head.next)

		if head == t.load(&t.head) {
			if head == tail {
				if tail == nil {
					return nil, false
				} else {
					atomic.CompareAndSwapPointer(&t.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
				}
			} else {
				res = next.task
				if atomic.CompareAndSwapPointer(&t.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
					next.task = nil
					head.Release()
					break
				}
			}
		}
	}

	return res, true
}

func (t *taskQueue) load(p *unsafe.Pointer) *taskNode {
	return (*taskNode)(atomic.LoadPointer(p))
}

func newTaskQueue() *taskQueue {
	q := &taskQueue{}

	node := newTaskNode(nil)
	q.head = unsafe.Pointer(node)
	q.tail = q.head

	return q
}

func NewTask(fn func(interface{}) error, arg interface{}) *Task {
	t := taskPool.Get().(*Task)
	t.fn = fn
	t.arg = arg

	return t
}

func newTaskNode(task *Task) *taskNode {
	node := taskNodePool.Get().(*taskNode)
	node.task = task
	node.next = unsafe.Pointer(nil)

	return node
}
