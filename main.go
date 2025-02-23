package main

import (
	"fmt"

	"fsedano.net/pq/priorityqueue"
)

func main() {
	fmt.Println("Slice-based Priority Queue:")
	var slicePQ priorityqueue.PriorityQueuer = priorityqueue.NewMultiPriorityQueue()
	testQueue(slicePQ)

	fmt.Println("\nRedis-based Priority Queue:")
	var redisPQ priorityqueue.PriorityQueuer = priorityqueue.NewRedisPriorityQueue("localhost:6379", "", 0)
	testQueue(redisPQ)
}

func testQueue(pq priorityqueue.PriorityQueuer) {
	pq.AddQueue("tasks")
	pq.Enqueue("tasks", "Process payment", 2)
	pq.Enqueue("tasks", "Emergency shutdown", 0)
	pq.Enqueue("tasks", "Update UI", 5)
	pq.InsertAtTop("tasks", "Critical task", 0)
	pq.InsertAtTop("tasks", "Urgent payment", 2)

	fmt.Println("Initial contents:")
	contents, _ := pq.ListContents("tasks")
	for priority, items := range contents {
		fmt.Printf("Priority %d: %v\n", priority, items)
	}

	pq.DeleteItem("tasks", "Process payment")

	fmt.Println("\nAfter deleting 'Process payment':")
	contents, _ = pq.ListContents("tasks")
	for priority, items := range contents {
		fmt.Printf("Priority %d: %v\n", priority, items)
	}

	fmt.Println("\nDequeuing items:")
	for i := 0; i < 3; i++ {
		item, _ := pq.Dequeue("tasks")
		fmt.Printf("Dequeued: %v\n", item)
	}
}
