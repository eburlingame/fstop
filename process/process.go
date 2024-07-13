package process

import (
	"log"
	"sync"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/resources"
)

func ProcessImageImport(r *Resources, image ImageImport) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in ProcessImageImport", r)
		}
	}()

	log.Printf("Processing image %s\n", image.UploadFilePath)

	wg := new(sync.WaitGroup)

	fileContents, err := r.Storage.GetFile(image.UploadFilePath)
	if err != nil || len(fileContents) == 0 {
		log.Printf("Error getting image from storage: %s\n", err)
		return
	}

	wg.Add(1)
	log.Printf("Processing image metadata, imageId: %s\n", image.ImageId)
	go ProcessImageMeta(r, wg, &image, fileContents)

	wg.Add(1)
	log.Printf("Processing image original, imageId: %s\n", image.ImageId)
	go ProcessImageOriginal(r, wg, &image, fileContents)

	wg.Add(len(image.Sizes))
	log.Printf("Processing image resizes, imageId: %s\n", image.ImageId)
	for _, size := range image.Sizes {
		go ProcessImageResize(r, wg, &image, size, fileContents)
	}

	wg.Wait()

	log.Printf("Updating processed status, imageId: %s\n", image.ImageId)
	r.Db.UpdateImageProcessedStatus(image.ImageId, true)

	log.Printf("Import of %s complete. Removing from upload directory.\n", image.UploadFilePath)
	r.Storage.DeleteFile(image.UploadFilePath)
}
