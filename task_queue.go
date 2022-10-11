package hio

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var taskQueueNodePool = &sync.Pool{
	New: func() interface{} {
		return &taskQueueNode{}
	},
}

type taskQueue struct {
	pool *sync.Pool
	size int32
	head unsafe.Pointer
	tail unsafe.Pointer
}

type taskQueueNode struct {
	next unsafe.Pointer
	fn   func() error
}

func (t *taskQueue) IsEmpty() bool {
	return atomic.LoadInt32(&t.size) == 0
}

func (t *taskQueue) Enqueue(fn func() error) {
	node := t.newNode(fn)
	for {
		tail := t.load(&t.tail)
		next := t.load(&tail.next)

		if tail == t.load(&t.tail) {
			if next == nil {
				if atomic.CompareAndSwapPointer(&tail.next, unsafe.Pointer(next), unsafe.Pointer(node)) {
					atomic.CompareAndSwapPointer(&t.tail, t.tail, unsafe.Pointer(node))
					atomic.AddInt32(&t.size, 1)
					return
				}
			} else {
				atomic.CompareAndSwapPointer(&t.tail, t.tail, tail.next)
			}
		}
	}
}

func (t *taskQueue) Dequeue() func() error {
	for {
		head := t.load(&t.head)
		tail := t.load(&t.tail)
		next := t.load(&head.next)

		if head == t.load(&t.head) {
			if head == tail {
				if next == nil {
					return nil
				}
				atomic.CompareAndSwapPointer(&t.tail, t.tail, tail.next)
			} else {
				fn := next.fn
				if atomic.CompareAndSwapPointer(&t.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
					next.fn = nil
					atomic.AddInt32(&t.size, -1)
					t.releaseNode(head)
					return fn
				}
			}
		}
	}
}

func (t *taskQueue) newNode(fn func() error) *taskQueueNode {
	node := taskQueueNodePool.Get().(*taskQueueNode)
	node.fn = fn
	return node
}

func (t *taskQueue) releaseNode(node *taskQueueNode) {
	node.fn = nil
	node.next = nil
	taskQueueNodePool.Put(node)
}

func (t *taskQueue) load(pointer *unsafe.Pointer) *taskQueueNode {
	return (*taskQueueNode)(atomic.LoadPointer(pointer))
}

func newTaskQueue() *taskQueue {
	queue := new(taskQueue)
	node := queue.newNode(nil)
	queue.head = unsafe.Pointer(node)
	queue.tail = queue.head

	return queue
}
