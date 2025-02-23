package main

import (
	"fmt"
	"reflect"
	"testing"

	"fsedano.net/pq/priorityqueue"
)

func TestPriorityQueue(t *testing.T) {
	tests := []struct {
		name string
		pq   priorityqueue.PriorityQueuer
	}{
		{"SlicePQ", priorityqueue.NewMultiPriorityQueue()},
		{"RedisPQ", priorityqueue.NewRedisPriorityQueue("localhost:6379", "nBr3nJu6hn", 0)},
	}

	// List of queue names used in tests
	queueNames := []string{
		"addqueue_test",
		"enqueue_test",
		"dequeue_test",
		"isempty_test",
		"listcontents_test",
		"getposition_test",
		"insertattop_test",
		"deleteitem_test",
		"bench_enqueue_test",
		"bench_dequeue_test",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq := tt.pq

			// Cleanup for RedisPQ before running subtests
			if tt.name == "RedisPQ" {
				if redisPQ, ok := pq.(*priorityqueue.RedisPriorityQueue); ok {
					err := redisPQ.ClearQueues(queueNames...)
					if err != nil {
						t.Fatalf("Failed to clear Redis queues: %v", err)
					}
				}
			}

			t.Run("NewPriorityQueue", func(t *testing.T) {
				if pq == nil {
					t.Error("New priority queue should return a non-nil value")
				}
			})

			t.Run("AddQueue", func(t *testing.T) {
				err := pq.AddQueue("addqueue_test")
				if err != nil {
					t.Errorf("AddQueue failed: %v", err)
				}
			})

			t.Run("Enqueue", func(t *testing.T) {
				pq.AddQueue("enqueue_test")
				err := pq.Enqueue("enqueue_test", "item1", 0)
				if err != nil {
					t.Errorf("Enqueue failed: %v", err)
				}

				err = pq.Enqueue("enqueue_test", "item2", 10)
				if err == nil {
					t.Error("Enqueue should fail with priority > 9")
				}

				err = pq.Enqueue("nonexistent", "item3", 0)
				if err == nil && tt.name == "SlicePQ" {
					t.Error("Enqueue should fail with non-existent queue for SlicePQ")
				}
			})

			t.Run("Dequeue", func(t *testing.T) {
				pq.AddQueue("dequeue_test")
				_, err := pq.Dequeue("dequeue_test")
				if err == nil {
					t.Error("Dequeue should fail on empty queue")
				}

				pq.Enqueue("dequeue_test", "low", 5)
				pq.Enqueue("dequeue_test", "high", 0)
				pq.Enqueue("dequeue_test", "medium", 2)

				item, err := pq.Dequeue("dequeue_test")
				if err != nil || item != "high" {
					t.Errorf("Dequeue should return highest priority item first, got %v, err: %v", item, err)
				}

				item, err = pq.Dequeue("dequeue_test")
				if err != nil || item != "medium" {
					t.Errorf("Dequeue should return medium priority next, got %v, err: %v", item, err)
				}
			})

			t.Run("IsEmpty", func(t *testing.T) {
				pq.AddQueue("isempty_test")
				empty, err := pq.IsEmpty("isempty_test")
				if err != nil {
					t.Errorf("IsEmpty failed: %v", err)
				}
				if !empty {
					t.Errorf("New queue should be empty, got %v", empty)
				}

				pq.Enqueue("isempty_test", "item", 0)
				empty, err = pq.IsEmpty("isempty_test")
				if err != nil {
					t.Errorf("IsEmpty failed: %v", err)
				}
				if empty {
					t.Errorf("Queue with item should not be empty, got %v", empty)
				}
			})

			t.Run("ListContents", func(t *testing.T) {
				pq.AddQueue("listcontents_test")
				contents, err := pq.ListContents("listcontents_test")
				if err != nil {
					t.Errorf("ListContents failed: %v", err)
				}
				if len(contents) != 0 {
					t.Errorf("Empty queue should return empty contents, got %v", contents)
				}

				pq.Enqueue("listcontents_test", "high", 0)
				pq.Enqueue("listcontents_test", "medium1", 2)
				pq.Enqueue("listcontents_test", "medium2", 2)

				contents, err = pq.ListContents("listcontents_test")
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

			t.Run("GetPosition", func(t *testing.T) {
				pq.AddQueue("getposition_test")
				priority, pos, err := pq.GetPosition("getposition_test", "missing")
				if err == nil || priority != -1 || pos != -1 {
					t.Errorf("GetPosition should fail for non-existent item, got %d, %d, err: %v", priority, pos, err)
				}

				pq.Enqueue("getposition_test", "first", 0)
				pq.Enqueue("getposition_test", "second", 0)
				pq.Enqueue("getposition_test", "third", 2)

				priority, pos, err = pq.GetPosition("getposition_test", "first")
				if err != nil || priority != 0 || pos != 0 {
					t.Errorf("Wrong position for 'first': got %d, %d, err: %v", priority, pos, err)
				}

				priority, pos, err = pq.GetPosition("getposition_test", "second")
				if err != nil || priority != 0 || pos != 1 {
					t.Errorf("Wrong position for 'second': got %d, %d, err: %v", priority, pos, err)
				}

				priority, pos, err = pq.GetPosition("getposition_test", "third")
				if err != nil || priority != 2 || pos != 0 {
					t.Errorf("Wrong position for 'third': got %d, %d, err: %v", priority, pos, err)
				}
			})

			t.Run("InsertAtTop", func(t *testing.T) {
				pq.AddQueue("insertattop_test")
				err := pq.InsertAtTop("insertattop_test", "item", 10)
				if err == nil {
					t.Error("InsertAtTop should fail with priority > 9")
				}

				err = pq.InsertAtTop("insertattop_test", "urgent", 0)
				if err != nil {
					t.Errorf("InsertAtTop failed: %v", err)
				}

				priority, pos, err := pq.GetPosition("insertattop_test", "urgent")
				if err != nil || priority != 0 || pos != 0 {
					t.Errorf("InsertAtTop item should be at priority 0, position 0, got %d, %d, err: %v", priority, pos, err)
				}

				pq.Enqueue("insertattop_test", "normal", 0)
				pq.Enqueue("insertattop_test", "medium", 2)
				pq.InsertAtTop("insertattop_test", "topmost", 0)
				pq.InsertAtTop("insertattop_test", "urgent_medium", 2)

				item, err := pq.Dequeue("insertattop_test")
				if err != nil || item != "topmost" {
					t.Errorf("InsertAtTop item at priority 0 should be dequeued first, got %v, err: %v", item, err)
				}

				item, err = pq.Dequeue("insertattop_test")
				if err != nil || item != "urgent" {
					t.Errorf("Second priority 0 item should be next, got %v, err: %v", item, err)
				}

				pq.Dequeue("insertattop_test") // Remove "normal"
				item, err = pq.Dequeue("insertattop_test")
				if err != nil || item != "urgent_medium" {
					t.Errorf("InsertAtTop item at priority 2 should be first in its level, got %v, err: %v", item, err)
				}
			})

			t.Run("DeleteItem", func(t *testing.T) {
				pq.AddQueue("deleteitem_test")

				err := pq.DeleteItem("deleteitem_test", "missing")
				if err == nil {
					t.Error("DeleteItem should fail for non-existent item in empty queue")
				}

				pq.Enqueue("deleteitem_test", "item1", 0)
				pq.Enqueue("deleteitem_test", "item2", 0)
				pq.Enqueue("deleteitem_test", "item3", 2)

				err = pq.DeleteItem("deleteitem_test", "item2")
				if err != nil {
					t.Errorf("DeleteItem failed: %v", err)
				}

				contents, err := pq.ListContents("deleteitem_test")
				if err != nil {
					t.Errorf("ListContents failed after delete: %v", err)
				}
				expected := map[int][]interface{}{
					0: {"item1"},
					2: {"item3"},
				}
				if !reflect.DeepEqual(contents, expected) {
					t.Errorf("DeleteItem didn't remove item correctly. Got %v, want %v", contents, expected)
				}

				err = pq.DeleteItem("deleteitem_test", "nonexistent")
				if err == nil {
					t.Error("DeleteItem should fail for non-existent item")
				}

				err = pq.DeleteItem("nonexistent", "item1")
				if err == nil && tt.name == "SlicePQ" {
					t.Error("DeleteItem should fail for non-existent queue in SlicePQ")
				}
			})
		})
	}
}

func BenchmarkEnqueue(b *testing.B) {
	pqs := []struct {
		name string
		pq   priorityqueue.PriorityQueuer
	}{
		{"SlicePQ", priorityqueue.NewMultiPriorityQueue()},
		{"RedisPQ", priorityqueue.NewRedisPriorityQueue("localhost:6379", "", 0)},
	}

	for _, pq := range pqs {
		b.Run(pq.name, func(b *testing.B) {
			// Cleanup for RedisPQ before benchmark
			if redisPQ, ok := pq.pq.(*priorityqueue.RedisPriorityQueue); ok {
				err := redisPQ.ClearQueues("bench_enqueue_test")
				if err != nil {
					b.Fatalf("Failed to clear Redis queue: %v", err)
				}
			}

			pq.pq.AddQueue("bench_enqueue_test")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pq.pq.Enqueue("bench_enqueue_test", fmt.Sprintf("item%d", i), i%10)
			}
		})
	}
}

func BenchmarkDequeue(b *testing.B) {
	pqs := []struct {
		name string
		pq   priorityqueue.PriorityQueuer
	}{
		{"SlicePQ", priorityqueue.NewMultiPriorityQueue()},
		{"RedisPQ", priorityqueue.NewRedisPriorityQueue("localhost:6379", "", 0)},
	}

	for _, pq := range pqs {
		b.Run(pq.name, func(b *testing.B) {
			// Cleanup for RedisPQ before benchmark
			if redisPQ, ok := pq.pq.(*priorityqueue.RedisPriorityQueue); ok {
				err := redisPQ.ClearQueues("bench_dequeue_test")
				if err != nil {
					b.Fatalf("Failed to clear Redis queue: %v", err)
				}
			}

			pq.pq.AddQueue("bench_dequeue_test")
			for i := 0; i < 1000; i++ {
				if i%2 == 0 {
					pq.pq.Enqueue("bench_dequeue_test", fmt.Sprintf("item%d", i), i%10)
				} else {
					pq.pq.InsertAtTop("bench_dequeue_test", fmt.Sprintf("item%d", i), i%10)
				}
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pq.pq.Dequeue("bench_dequeue_test")
			}
		})
	}
}
