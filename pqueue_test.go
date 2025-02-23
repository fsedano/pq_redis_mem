package main

import (
	"fmt"
	"reflect"
	"testing"

	"fsedano.net/pq/priorityqueue" // Adjust this import path
)

func TestPriorityQueue(t *testing.T) {
	tests := []struct {
		name string
		pq   priorityqueue.PriorityQueuer
	}{
		{"SlicePQ", priorityqueue.NewMultiPriorityQueue()},
		{"RedisPQ", priorityqueue.NewRedisPriorityQueue("localhost:6379", "", 0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("NewPriorityQueue", func(t *testing.T) {
				if tt.pq == nil {
					t.Error("New priority queue should return a non-nil value")
				}
			})

			t.Run("AddQueue", func(t *testing.T) {
				err := tt.pq.AddQueue("test")
				if err != nil {
					t.Errorf("AddQueue failed: %v", err)
				}
			})

			t.Run("Enqueue", func(t *testing.T) {
				tt.pq.AddQueue("test")
				err := tt.pq.Enqueue("test", "item1", 0)
				if err != nil {
					t.Errorf("Enqueue failed: %v", err)
				}

				err = tt.pq.Enqueue("test", "item2", 10)
				if err == nil {
					t.Error("Enqueue should fail with priority > 9")
				}

				err = tt.pq.Enqueue("nonexistent", "item3", 0)
				if err == nil && tt.name == "SlicePQ" { // Redis doesn't require queue existence
					t.Error("Enqueue should fail with non-existent queue for SlicePQ")
				}
			})

			t.Run("Dequeue", func(t *testing.T) {
				tt.pq.AddQueue("test")
				_, err := tt.pq.Dequeue("test")
				if err == nil {
					t.Error("Dequeue should fail on empty queue")
				}

				tt.pq.Enqueue("test", "low", 5)
				tt.pq.Enqueue("test", "high", 0)
				tt.pq.Enqueue("test", "medium", 2)

				item, err := tt.pq.Dequeue("test")
				if err != nil || item != "high" {
					t.Errorf("Dequeue should return highest priority item first, got %v", item)
				}

				item, err = tt.pq.Dequeue("test")
				if err != nil || item != "medium" {
					t.Errorf("Dequeue should return medium priority next, got %v", item)
				}
			})

			t.Run("IsEmpty", func(t *testing.T) {
				tt.pq.AddQueue("test")
				empty, err := tt.pq.IsEmpty("test")
				if err != nil || !empty {
					t.Error("New queue should be empty")
				}

				tt.pq.Enqueue("test", "item", 0)
				empty, err = tt.pq.IsEmpty("test")
				if err != nil || empty {
					t.Error("Queue with item should not be empty")
				}
			})

			t.Run("ListContents", func(t *testing.T) {
				tt.pq.AddQueue("test")
				contents, err := tt.pq.ListContents("test")
				if err != nil || len(contents) != 0 {
					t.Error("Empty queue should return empty contents")
				}

				tt.pq.Enqueue("test", "high", 0)
				tt.pq.Enqueue("test", "medium1", 2)
				tt.pq.Enqueue("test", "medium2", 2)

				contents, err = tt.pq.ListContents("test")
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
				tt.pq.AddQueue("test")
				priority, pos, err := tt.pq.GetPosition("test", "missing")
				if err == nil || priority != -1 || pos != -1 {
					t.Error("GetPosition should fail for non-existent item")
				}

				tt.pq.Enqueue("test", "first", 0)
				tt.pq.Enqueue("test", "second", 0)
				tt.pq.Enqueue("test", "third", 2)

				priority, pos, err = tt.pq.GetPosition("test", "first")
				if err != nil || priority != 0 || pos != 0 {
					t.Errorf("Wrong position for 'first': %d, %d", priority, pos)
				}

				priority, pos, err = tt.pq.GetPosition("test", "second")
				if err != nil || priority != 0 || pos != 1 {
					t.Errorf("Wrong position for 'second': %d, %d", priority, pos)
				}

				priority, pos, err = tt.pq.GetPosition("test", "third")
				if err != nil || priority != 2 || pos != 0 {
					t.Errorf("Wrong position for 'third': %d, %d", priority, pos)
				}
			})

			t.Run("InsertAtTop", func(t *testing.T) {
				tt.pq.AddQueue("test")
				err := tt.pq.InsertAtTop("test", "item", 10)
				if err == nil {
					t.Error("InsertAtTop should fail with priority > 9")
				}

				err = tt.pq.InsertAtTop("test", "urgent", 0)
				if err != nil {
					t.Errorf("InsertAtTop failed: %v", err)
				}

				priority, pos, err := tt.pq.GetPosition("test", "urgent")
				if err != nil || priority != 0 || pos != 0 {
					t.Errorf("InsertAtTop item should be at priority 0, position 0, got %d, %d", priority, pos)
				}

				tt.pq.Enqueue("test", "normal", 0)
				tt.pq.Enqueue("test", "medium", 2)
				tt.pq.InsertAtTop("test", "topmost", 0)
				tt.pq.InsertAtTop("test", "urgent_medium", 2)

				item, err := tt.pq.Dequeue("test")
				if err != nil || item != "topmost" {
					t.Errorf("InsertAtTop item at priority 0 should be dequeued first, got %v", item)
				}

				item, err = tt.pq.Dequeue("test")
				if err != nil || item != "urgent" {
					t.Errorf("Second priority 0 item should be next, got %v", item)
				}

				tt.pq.Dequeue("test") // Remove "normal"
				item, err = tt.pq.Dequeue("test")
				if err != nil || item != "urgent_medium" {
					t.Errorf("InsertAtTop item at priority 2 should be first in its level, got %v", item)
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
			pq.pq.AddQueue("test")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pq.pq.Enqueue("test", fmt.Sprintf("item%d", i), i%10)
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
			pq.pq.AddQueue("test")
			for i := 0; i < 1000; i++ {
				if i%2 == 0 {
					pq.pq.Enqueue("test", fmt.Sprintf("item%d", i), i%10)
				} else {
					pq.pq.InsertAtTop("test", fmt.Sprintf("item%d", i), i%10)
				}
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pq.pq.Dequeue("test")
			}
		})
	}
}
