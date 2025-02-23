// priorityqueue_test.go
package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMultiPriorityQueue(t *testing.T) {
	// Test creation of new MultiPriorityQueue
	t.Run("NewMultiPriorityQueue", func(t *testing.T) {
		var mpq PriorityQueuer = NewMultiPriorityQueue()
		if mpq == nil {
			t.Error("NewMultiPriorityQueue should return a non-nil value")
		}
	})

	// Test adding queues
	t.Run("AddQueue", func(t *testing.T) {
		var mpq PriorityQueuer = NewMultiPriorityQueue()

		// Add new queue
		err := mpq.AddQueue("test")
		if err != nil {
			t.Errorf("AddQueue failed: %v", err)
		}

		// Try adding duplicate queue
		err = mpq.AddQueue("test")
		if err == nil {
			t.Error("AddQueue should fail on duplicate queue name")
		}
	})

	// Test enqueue operation
	t.Run("Enqueue", func(t *testing.T) {
		var mpq PriorityQueuer = NewMultiPriorityQueue()
		mpq.AddQueue("test")

		// Valid priority
		err := mpq.Enqueue("test", "item1", 0)
		if err != nil {
			t.Errorf("Enqueue failed: %v", err)
		}

		// Invalid priority
		err = mpq.Enqueue("test", "item2", 10)
		if err == nil {
			t.Error("Enqueue should fail with priority > 9")
		}

		// Non-existent queue
		err = mpq.Enqueue("nonexistent", "item3", 0)
		if err == nil {
			t.Error("Enqueue should fail with non-existent queue")
		}
	})

	// Test dequeue operation
	t.Run("Dequeue", func(t *testing.T) {
		var mpq PriorityQueuer = NewMultiPriorityQueue()
		mpq.AddQueue("test")

		// Empty queue
		_, err := mpq.Dequeue("test")
		if err == nil {
			t.Error("Dequeue should fail on empty queue")
		}

		// Add items and test priority order
		mpq.Enqueue("test", "low", 5)
		mpq.Enqueue("test", "high", 0)
		mpq.Enqueue("test", "medium", 2)

		item, err := mpq.Dequeue("test")
		if err != nil || item != "high" {
			t.Errorf("Dequeue should return highest priority item first, got %v", item)
		}

		item, err = mpq.Dequeue("test")
		if err != nil || item != "medium" {
			t.Errorf("Dequeue should return medium priority next, got %v", item)
		}
	})

	// Test IsEmpty
	t.Run("IsEmpty", func(t *testing.T) {
		var mpq PriorityQueuer = NewMultiPriorityQueue()
		mpq.AddQueue("test")

		empty, err := mpq.IsEmpty("test")
		if err != nil || !empty {
			t.Error("New queue should be empty")
		}

		mpq.Enqueue("test", "item", 0)
		empty, err = mpq.IsEmpty("test")
		if err != nil || empty {
			t.Error("Queue with item should not be empty")
		}
	})

	// Test ListContents
	t.Run("ListContents", func(t *testing.T) {
		var mpq PriorityQueuer = NewMultiPriorityQueue()
		mpq.AddQueue("test")

		// Empty queue
		contents, err := mpq.ListContents("test")
		if err != nil || len(contents) != 0 {
			t.Error("Empty queue should return empty contents")
		}

		// Add items
		mpq.Enqueue("test", "high", 0)
		mpq.Enqueue("test", "medium1", 2)
		mpq.Enqueue("test", "medium2", 2)

		contents, err = mpq.ListContents("test")
		if err != nil {
			t.Errorf("ListContents failed: %v", err)
		}

		expected := map[int][]interface{}{
			0: {"high"},
			2: {"medium1", "medium2"},
		}
		if !reflect.DeepEqual(contents, expected) {
			t.Errorf("ListContents wrong result. Got %v, want %v", contents, expected)
		}
	})

	// Test GetPosition
	t.Run("GetPosition", func(t *testing.T) {
		var mpq PriorityQueuer = NewMultiPriorityQueue()
		mpq.AddQueue("test")

		// Non-existent item
		priority, pos, err := mpq.GetPosition("test", "missing")
		if err == nil || priority != -1 || pos != -1 {
			t.Error("GetPosition should fail for non-existent item")
		}

		// Add items and check positions
		mpq.Enqueue("test", "first", 0)
		mpq.Enqueue("test", "second", 0)
		mpq.Enqueue("test", "third", 2)

		priority, pos, err = mpq.GetPosition("test", "first")
		if err != nil || priority != 0 || pos != 0 {
			t.Errorf("Wrong position for 'first': %d, %d", priority, pos)
		}

		priority, pos, err = mpq.GetPosition("test", "second")
		if err != nil || priority != 0 || pos != 1 {
			t.Errorf("Wrong position for 'second': %d, %d", priority, pos)
		}

		priority, pos, err = mpq.GetPosition("test", "third")
		if err != nil || priority != 2 || pos != 0 {
			t.Errorf("Wrong position for 'third': %d, %d", priority, pos)
		}
	})

	// Test InsertAtTop
	t.Run("InsertAtTop", func(t *testing.T) {
		var mpq PriorityQueuer = NewMultiPriorityQueue()
		mpq.AddQueue("test")

		// Test invalid priority
		err := mpq.InsertAtTop("test", "item", 10)
		if err == nil {
			t.Error("InsertAtTop should fail with priority > 9")
		}

		// Insert into empty queue at priority 0
		err = mpq.InsertAtTop("test", "urgent", 0)
		if err != nil {
			t.Errorf("InsertAtTop failed: %v", err)
		}

		// Check it's at priority 0, position 0
		priority, pos, err := mpq.GetPosition("test", "urgent")
		if err != nil || priority != 0 || pos != 0 {
			t.Errorf("InsertAtTop item should be at priority 0, position 0, got %d, %d", priority, pos)
		}

		// Add more items at different priorities
		mpq.Enqueue("test", "normal", 0)
		mpq.Enqueue("test", "medium", 2)
		mpq.InsertAtTop("test", "topmost", 0)
		mpq.InsertAtTop("test", "urgent_medium", 2)

		// Verify priority 0 order
		item, err := mpq.Dequeue("test")
		if err != nil || item != "topmost" {
			t.Errorf("InsertAtTop item at priority 0 should be dequeued first, got %v", item)
		}

		item, err = mpq.Dequeue("test")
		if err != nil || item != "urgent" {
			t.Errorf("Second priority 0 item should be next, got %v", item)
		}

		// Verify priority 2 order
		mpq.Dequeue("test") // Remove "normal" from priority 0
		item, err = mpq.Dequeue("test")
		if err != nil || item != "urgent_medium" {
			t.Errorf("InsertAtTop item at priority 2 should be first in its level, got %v", item)
		}

		// Test non-existent queue
		err = mpq.InsertAtTop("nonexistent", "test", 0)
		if err == nil {
			t.Error("InsertAtTop should fail with non-existent queue")
		}
	})
}

// Benchmark for Enqueue operation
func BenchmarkEnqueue(b *testing.B) {
	var mpq PriorityQueuer = NewMultiPriorityQueue()
	mpq.AddQueue("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mpq.Enqueue("test", fmt.Sprintf("item%d", i), i%10)
	}
}

// Benchmark for Dequeue operation
func BenchmarkDequeue(b *testing.B) {
	var mpq PriorityQueuer = NewMultiPriorityQueue()
	mpq.AddQueue("test")

	// Fill queue
	for i := 0; i < 1000; i++ {
		if i%2 == 0 {
			mpq.Enqueue("test", fmt.Sprintf("item%d", i), i%10)
		} else {
			mpq.InsertAtTop("test", fmt.Sprintf("item%d", i), i%10)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mpq.Dequeue("test")
	}
}
