package process

import (
	"fmt"

	. "github.com/eburlingame/fstop/resources"
)

const NUM_WORKERS = 10

type ImportBatchRequest struct {
	ImportBatchId string
	Images        []ImageImport
}

type ImportTask struct {
	Resources *Resources
	Image     ImageImport
}

func imageWorker(id int, queue chan ImportTask) {
	for tsk := range queue {
		ProcessImageImport(tsk.Resources, tsk.Image)
	}
	fmt.Printf("Worker %d done\n", id)
}

func ImportImageBatch(r *Resources, batch ImportBatchRequest) {
	queue := make(chan ImportTask)
	imgCount := len(batch.Images)

	go func() {
		for _, img := range batch.Images {
			// loop over all items
			queue <- ImportTask{
				Resources: r,
				Image:     img,
			}
		}
		close(queue)
	}()

	numWorkers := NUM_WORKERS
	if imgCount < numWorkers {
		numWorkers = imgCount
	}

	for i := 0; i < numWorkers; i++ {
		go imageWorker(i, queue)
	}
}
