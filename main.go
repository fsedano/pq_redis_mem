package main

import (
	"fmt"
	"sync"
)

// PriorityQueuer defines the interface for priority queue operations
type PriorityQueuer interface {
	AddQueue(name string) error
	Enqueue(queueName string, value interface{}, priority int) error
	Dequeue(queueName string) (interface{}, error)
	IsEmpty(queueName string) (bool, error)
	ListContents(queueName string) (map[int][]interface{}, error)
	GetPosition(queueName string, value interface{}) (int, int, error)
	InsertAtTop(queueName string, value interface{}, priority int) error
}

// Item represents an element in the priority queue
type Item struct {
	Value    interface{}
	Priority int
}

// PriorityQueue represents a single priority queue with multiple priority levels
type PriorityQueue struct {
	queues [][]Item // Slice of slices: each inner slice is a priority level
	mutex  sync.Mutex
}

// MultiPriorityQueue manages multiple named priority queues
type MultiPriorityQueue struct {
	queues map[string]*PriorityQueue
	mutex  sync.Mutex
}

// NewMultiPriorityQueue creates a new multi-priority queue system
func NewMultiPriorityQueue() PriorityQueuer {
	return &MultiPriorityQueue{
		queues: make(map[string]*PriorityQueue),
	}
}

// NewPriorityQueue creates a new single priority queue with 10 priority levels
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		queues: make([][]Item, 10),
	}
	for i := range pq.queues {
		pq.queues[i] = make([]Item, 0)
	}
	return pq
}

// AddQueue adds a new named priority queue to the system
func (mpq *MultiPriorityQueue) AddQueue(name string) error {
	mpq.mutex.Lock()
	defer mpq.mutex.Unlock()

	if _, exists := mpq.queues[name]; exists {
		return fmt.Errorf("queue '%s' already exists", name)
	}

	mpq.queues[name] = NewPriorityQueue()
	return nil
}

// Enqueue adds an item to a specific queue with given priority
func (mpq *MultiPriorityQueue) Enqueue(queueName string, value interface{}, priority int) error {
	if priority < 0 || priority > 9 {
		return fmt.Errorf("priority must be between 0 and 9")
	}

	mpq.mutex.Lock()
	pq, exists := mpq.queues[queueName]
	mpq.mutex.Unlock()

	if !exists {
		return fmt.Errorf("queue '%s' does not exist", queueName)
	}

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.queues[priority] = append(pq.queues[priority], Item{Value: value, Priority: priority})
	return nil
}

// Dequeue removes and returns the highest priority item from a specific queue
func (mpq *MultiPriorityQueue) Dequeue(queueName string) (interface{}, error) {
	mpq.mutex.Lock()
	pq, exists := mpq.queues[queueName]
	mpq.mutex.Unlock()

	if !exists {
		return nil, fmt.Errorf("queue '%s' does not exist", queueName)
	}

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	for i := 0; i < 10; i++ {
		if len(pq.queues[i]) > 0 {
			item := pq.queues[i][0]
			pq.queues[i] = pq.queues[i][1:]
			return item.Value, nil
		}
	}

	return nil, fmt.Errorf("queue '%s' is empty", queueName)
}

// IsEmpty checks if a specific queue is empty
func (mpq *MultiPriorityQueue) IsEmpty(queueName string) (bool, error) {
	mpq.mutex.Lock()
	pq, exists := mpq.queues[queueName]
	mpq.mutex.Unlock()

	if !exists {
		return false, fmt.Errorf("queue '%s' does not exist", queueName)
	}

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	for i := 0; i < 10; i++ {
		if len(pq.queues[i]) > 0 {
			return false, nil
		}
	}
	return true, nil
}

// ListContents returns all items in a specific queue, organized by priority
func (mpq *MultiPriorityQueue) ListContents(queueName string) (map[int][]interface{}, error) {
	mpq.mutex.Lock()
	pq, exists := mpq.queues[queueName]
	mpq.mutex.Unlock()

	if !exists {
		return nil, fmt.Errorf("queue '%s' does not exist", queueName)
	}

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	contents := make(map[int][]interface{})
	for priority := 0; priority < 10; priority++ {
		if len(pq.queues[priority]) > 0 {
			values := make([]interface{}, len(pq.queues[priority]))
			for i, item := range pq.queues[priority] {
				values[i] = item.Value
			}
			contents[priority] = values
		}
	}

	return contents, nil
}

// GetPosition returns the position of a specific value in its priority level
func (mpq *MultiPriorityQueue) GetPosition(queueName string, value interface{}) (int, int, error) {
	mpq.mutex.Lock()
	pq, exists := mpq.queues[queueName]
	mpq.mutex.Unlock()

	if !exists {
		return -1, -1, fmt.Errorf("queue '%s' does not exist", queueName)
	}

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	for priority := 0; priority < 10; priority++ {
		for pos, item := range pq.queues[priority] {
			if fmt.Sprintf("%v", item.Value) == fmt.Sprintf("%v", value) {
				return priority, pos, nil
			}
		}
	}

	return -1, -1, fmt.Errorf("value '%v' not found in queue '%s'", value, queueName)
}

// InsertAtTop inserts an item at the beginning of the specified priority level
func (mpq *MultiPriorityQueue) InsertAtTop(queueName string, value interface{}, priority int) error {
	if priority < 0 || priority > 9 {
		return fmt.Errorf("priority must be between 0 and 9")
	}

	mpq.mutex.Lock()
	pq, exists := mpq.queues[queueName]
	mpq.mutex.Unlock()

	if !exists {
		return fmt.Errorf("queue '%s' does not exist", queueName)
	}

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.queues[priority] = append([]Item{{Value: value, Priority: priority}}, pq.queues[priority]...)
	return nil
}

// Updated main function using the interface
func main() {
	var mpq PriorityQueuer = NewMultiPriorityQueue()

	// Add a queue
	mpq.AddQueue("tasks")

	// Add items
	mpq.Enqueue("tasks", "Process payment", 2)
	mpq.Enqueue("tasks", "Emergency shutdown", 0)
	mpq.Enqueue("tasks", "Update UI", 5)
	mpq.Enqueue("tasks", "Backup data", 2)

	// Insert at top of priority 0 and 2
	mpq.InsertAtTop("tasks", "Critical task", 0)
	mpq.InsertAtTop("tasks", "Urgent payment", 2)

	// List contents
	contents, _ := mpq.ListContents("tasks")
	fmt.Println("Queue contents:")
	for priority, items := range contents {
		fmt.Printf("Priority %d: %v\n", priority, items)
	}

	// Demonstrate order
	fmt.Println("\nDequeuing items:")
	for i := 0; i < 3; i++ {
		item, _ := mpq.Dequeue("tasks")
		fmt.Printf("Dequeued: %v\n", item)
	}
}
