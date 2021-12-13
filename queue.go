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

func processMessages(queue chan Task) {
	for {
		tsk := <-queue
		fmt.Printf("%v\n", tsk)

		time.Sleep(time.Second * time.Duration(rand.Float32()*5))
	}
}

func QueueGetHandler(r *Resources) gin.HandlerFunc {

	return func(c *gin.Context) {

		queue := make(chan Task)

		for i := 0; i < 10; i++ {
			go func(i int) {
				queue <- Task{
					Id: fmt.Sprintf("Task #%d", i),
				}
			}(i)
		}

		for i := 0; i < 3; i++ {
			go processMessages(queue)
		}

		c.String(200, "%s", "Hi")
	}
}
