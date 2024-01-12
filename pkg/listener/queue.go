package listener

import (
	"fmt"
	"strings"
	"sync"

	"github.com/KyberNetwork/evmlistener/pkg/types"
)

// Queue holds values in a slice.
type Queue struct {
	values  []*types.Block
	start   int
	maxSize int
	size    int

	seq uint64
	mu  sync.Mutex
}

// NewQueue instantiates a new empty queue with the specified size of maximum number of elements that it can hold.
// This max size of the buffer cannot be changed.
func NewQueue(maxSize int) *Queue {
	if maxSize < 1 {
		panic("Invalid maxSize, should be at least 1")
	}

	queue := &Queue{maxSize: maxSize}
	queue.clear()

	return queue
}

func (q *Queue) insertAt(value *types.Block, idx int) {
	q.values[(q.start+idx)%q.maxSize] = value
	q.size++
}

func (q *Queue) insert(seq uint64, value *types.Block) {
	if seq < q.seq {
		return
	}

	if q.isFull() {
		q.dequeue()
	}

	// Ignore missing values at the start of queue.
	if int(seq-q.seq) >= q.maxSize {
		for i := 0; i <= int(seq-q.seq)-q.maxSize; i++ {
			q.dequeue()
		}
	}

	q.insertAt(value, int(seq-q.seq))
}

// Insert inserts new block into queue relative to current seq.
func (q *Queue) Insert(seq uint64, value *types.Block) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.insert(seq, value)
}

func (q *Queue) dequeue() (*types.Block, bool) {
	if q.empty() {
		return nil, false
	}

	value, ok := q.values[q.start], true
	if value != nil {
		q.values[q.start] = nil
		q.size--
	} else {
		ok = false
	}

	q.start++
	if q.start >= q.maxSize {
		q.start = 0
	}
	q.seq++

	return value, ok
}

// Dequeue removes the first element of the queue and returns it, or nil if queue is empty.
// Second return parameter is true, unless the queue was empty and there was nothing to dequeue.
func (q *Queue) Dequeue() (*types.Block, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.dequeue()
}

// Peek returns first element of the queue without removing it, or nil if queue is empty.
// Second return parameter is true, unless the queue was empty and there was nothing to peek.
func (q *Queue) Peek() (*types.Block, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.empty() {
		return nil, false
	}

	value := q.values[q.start]
	if value == nil {
		return nil, false
	}

	return value, true
}

func (q *Queue) empty() bool {
	return q.size == 0
}

// Empty returns true if queue does not contain any elements.
func (q *Queue) Empty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.empty()
}

func (q *Queue) isFull() bool {
	return q.size == q.maxSize
}

// Full returns true if the queue is full.
func (q *Queue) Full() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.isFull()
}

// Size returns number of elements within the queue.
func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.size
}

func (q *Queue) clear() {
	if q.values == nil {
		q.values = make([]*types.Block, q.maxSize)
	} else {
		for i := range q.values {
			q.values[i] = nil
		}
	}

	q.start = 0
	q.size = 0
	q.seq = 1
}

// Clear removes all elements from the queue.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.clear()
}

// Values returns all elements in the queue.
func (q *Queue) Values() []*types.Block {
	q.mu.Lock()
	defer q.mu.Unlock()

	values := make([]*types.Block, 0, q.size)
	for i := 0; i < q.maxSize; i++ {
		v := q.values[(q.start+i)%q.maxSize]
		if v != nil {
			values = append(values, v)
		}
	}

	return values
}

// String returns a string representation of container.
func (q *Queue) String() string {
	str := "CircularBuffer\n"

	qValues := q.Values()
	values := make([]string, 0, len(qValues))
	for _, value := range qValues {
		values = append(values, fmt.Sprintf("%+v", value))
	}

	str += strings.Join(values, ", ")

	return str
}

// SequenceNumber returns current sequence number of queue.
func (q *Queue) SequenceNumber() uint64 {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.seq
}

// SetSequenceNumber sets current sequence number of queue.
func (q *Queue) SetSequenceNumber(seq uint64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.seq = seq
}
