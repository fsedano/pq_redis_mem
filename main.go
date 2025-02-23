package main

import (
	"fmt"
	"log"

	"fsedano.net/pq/priorityqueue" // Adjust this import path based on your project structure
)

func main() {
	fmt.Println("Slice-based Priority Queue:")
	var slicePQ priorityqueue.PriorityQueuer = priorityqueue.NewMultiPriorityQueue()
	testQueue(slicePQ)

	fmt.Println("\nRedis-based Priority Queue:")
	var redisPQ priorityqueue.PriorityQueuer = priorityqueue.NewRedisPriorityQueue("localhost:6379", "JBYonWxOus", 0)
	testQueue(redisPQ)
}

func testQueue(pq priorityqueue.PriorityQueuer) {
	pq.AddQueue("tasks")
	pq.Enqueue("tasks", "Process payment 3", 2)
	pq.Enqueue("tasks", "Emergency shutdown 3", 0)
	pq.Enqueue("tasks", "Update UI 3", 5)
	pq.InsertAtTop("tasks", "Critical task 3", 0)
	pq.InsertAtTop("tasks", "Urgent payment 3", 2)

	contents, _ := pq.ListContents("tasks")
	fmt.Println("Queue contents:")
	for priority, items := range contents {
		fmt.Printf("Priority %d: %v\n", priority, items)
	}

	fmt.Println("\nDequeuing items:")
	//for i := 0; i < 3; i++ {
	for {
		item, err := pq.Dequeue("tasks")
		fmt.Printf("Dequeued: %v\n", item)
		if err != nil {
			log.Printf("Error deq: %s", err)
			break
		}
	}
}
