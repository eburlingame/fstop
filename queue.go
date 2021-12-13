package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

type Task struct {
	Id string
}

func processMessages(id int, queue chan Task) {
	for tsk := range queue {
		fmt.Printf("%v\n", tsk)
		time.Sleep(time.Second * time.Duration(rand.Float32()*5))
	}
	fmt.Printf("Worker %d done", id)
}

func QueueGetHandler(r *Resources) gin.HandlerFunc {

	return func(c *gin.Context) {
		queue := make(chan Task)

		go func() {
			for i := 0; i < 10; i++ {
				// loop over all items
				queue <- Task{
					Id: fmt.Sprintf("Task #%d", i),
				}
			}
			close(queue)
		}()

		for i := 0; i < 3; i++ {
			go processMessages(i, queue)
		}

		c.String(200, "%s", "Hi")
	}
}
