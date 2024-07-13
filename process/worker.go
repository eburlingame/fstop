package process

import (
	"log"
	"runtime"
	"time"

	"github.com/eburlingame/fstop/resources"
)

const NUM_WORKERS = 4

func worker(worker_num int, resources *resources.Resources) {
	queue := resources.Queue

	log.Printf("Worker %d started", worker_num)
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
	for w := 0; w < NUM_WORKERS; w++ {
		go worker(w, resources)
	}
}
