package process

import (
	"log"
	"runtime"
	"time"

	"github.com/eburlingame/fstop/resources"
)

const NUM_WORKERS = 4

func worker(resources *resources.Resources) {
	queue := resources.Queue

	log.Println("Worker started")
	for {
		taskId, task, err := queue.Receive()
		if err != nil {
			log.Fatalln(err)
		}

		if task != nil {
			ProcessImageImport(resources, *task)

			if err := queue.Done(*taskId); err != nil {
				log.Fatalln(err)
			}
		}

		runtime.Gosched()
		time.Sleep(1 * time.Second)
	}
}

func InitWorkers(resources *resources.Resources) {
	for w := 1; w <= NUM_WORKERS; w++ {
		go worker(resources)
	}
}
