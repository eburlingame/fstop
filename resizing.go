package main

import (
	"fmt"
	"os"

	"github.com/h2non/bimg"
)

func convertImageToPng(imgBuffer []byte) ([]byte, error) {
	// Convert the image to a png format
	pngImage, err := bimg.NewImage(imgBuffer).Convert(bimg.PNG)
	if err != nil {
		return nil, err
	}

	return pngImage, nil
}

func resizeLocalImage(pngImageBuffer []byte, destFilename string, newWidth int, newHeight int) error {
	// Resize the image to the new size
	resizedImage, err := bimg.NewImage(pngImageBuffer).Resize(newWidth, newHeight)
	if err != nil {
		return err
	}

	// Write image to disk
	fmt.Printf("Writing %s\n", destFilename)
	return bimg.Write(destFilename, resizedImage)
}

func resizeAndUploadImage(r *Resources, pngImageBuffer []byte, imageFolder string, fileId string, suffix string, width int, height int) error {
	resizedFilename := fileId + suffix + ".png"
	resizedLocalFilename := imageFolder + resizedFilename
	storagePath := r.config.S3MediaFolder + "/" + resizedFilename
	pngContentType := "image/png"

	// Resize the image
	err := resizeLocalImage(pngImageBuffer, resizedLocalFilename, width, height)
	defer os.Remove(resizedLocalFilename)

	if err != nil {
		fmt.Printf("Error writing resized image: %s\n", err)
		return err
	}

	// Upload the image to the storage adapter
	err = r.storage.PutFile(resizedLocalFilename, storagePath, pngContentType)
	if err != nil {
		fmt.Printf("Error uploading file to storage: %s\n", err)
		return err
	}

	r.db.AddFile(&File{
		FileId:      fileId,
		Filename:    resizedFilename,
		StoragePath: storagePath,
		PublicURL:   r.config.S3BaseUrl + storagePath,
		IsOriginal:  false,
		Width:       uint64(width),
		Height:      uint64(height),
	})

	// Remove the temporary file
	return nil
}

type OutputImageSize struct {
	LongEdge int
	Suffix   string
}

func processImageUpload(r *Resources, imageFolder string, fileId string, extension string, contentType string) {
	filename := fileId + extension
	localPath := imageFolder + filename
	storagePath := r.config.S3MediaFolder + "/" + filename

	sizes := []OutputImageSize{
		{
			LongEdge: 200,
			Suffix:   "_thumb",
		},
		{
			LongEdge: 600,
			Suffix:   "_small",
		},
		{
			LongEdge: 1080,
			Suffix:   "_medium",
		},
		{
			LongEdge: 1920,
			Suffix:   "_large",
		},
	}

	// Load image into buffer for resizing
	imgBuffer, err := bimg.Read(localPath)
	if err != nil {
		fmt.Printf("Unable to read image file %s: %s\n", localPath, err)
		return
	}

	imgSize, err := bimg.Size(imgBuffer)
	if err != nil {
		fmt.Printf("Unable to read image size for %s: %s\n", fileId, err)
		return
	}

	// Upload the original file into storage
	r.storage.PutFile(localPath, storagePath, contentType)
	r.db.AddFile(&File{
		FileId:      fileId,
		Filename:    filename,
		StoragePath: storagePath,
		PublicURL:   r.config.S3BaseUrl + storagePath,
		IsOriginal:  true,
		Width:       uint64(imgSize.Width),
		Height:      uint64(imgSize.Height),
	})

	pngImageBuffer, err := convertImageToPng(imgBuffer)
	if err != nil {
		fmt.Printf("Error converting to png: %s\n", err)
	}

	// Copy the original file size into png format
	err = resizeAndUploadImage(r, pngImageBuffer, imageFolder, fileId, "", imgSize.Width, imgSize.Height)
	if err != nil {
		fmt.Printf("Error resizing image: %s\n", err)
	}

	// Compute long-edge resize images
	for _, size := range sizes {
		newWidth, newHeight := ResizeLongEdgeDimensions(imgSize.Width, imgSize.Height, size.LongEdge)

		err := resizeAndUploadImage(r, pngImageBuffer, imageFolder, fileId, size.Suffix, newWidth, newHeight)
		if err != nil {
			fmt.Printf("Error resizing image: %s\n", err)
			continue
		}
	}

	// Remove the uploaded file
	os.Remove(localPath)

	r.db.UpdateImageProcessedStatus(fileId, true)
	fmt.Printf("Resizing %s complete\n", localPath)
}