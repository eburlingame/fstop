package process

import (
	"fmt"
	"sync"

	. "github.com/eburlingame/fstop/resources"
)

type OutputImageSize struct {
	LongEdge    int
	Suffix      string
	Extension   string
	Format      string
	ContentType string
}
type ImageImport struct {
	ImageId        string
	ImportBatchId  string
	UploadFilePath string
	Sizes          []OutputImageSize
}

func ProcessImageImport(r *Resources, image ImageImport) {
	fmt.Printf("Processing image %s\n", image.UploadFilePath)

	wg := new(sync.WaitGroup)

	fileContents, err := r.Storage.GetFile(image.UploadFilePath)
	if err != nil || len(fileContents) == 0 {
		fmt.Printf("Error getting image from storage: %s\n", err)
		return
	}

	wg.Add(1)
	go ProcessImageMeta(r, wg, &image, fileContents)

	wg.Add(1)
	go ProcessImageOriginal(r, wg, &image, fileContents)

	wg.Add(len(image.Sizes))
	for _, size := range image.Sizes {
		go ProcessImageResize(r, wg, &image, size, fileContents)
	}

	wg.Wait()

	r.Db.UpdateImageProcessedStatus(image.ImageId, true)

	fmt.Printf("Import of %s complete. Removing from upload directory.\n", image.UploadFilePath)
	r.Storage.DeleteFile(image.UploadFilePath)
}
