package main

// import (
// 	"fmt"
// 	"os"
// 	"time"

// 	"github.com/h2non/bimg"
// )

// func convertImageToPng(imgBuffer []byte) ([]byte, error) {
// 	// Convert the image to a png format
// 	pngImage, err := bimg.NewImage(imgBuffer).Convert(bimg.PNG)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return pngImage, nil
// }

// func convertImageToJpeg(imgBuffer []byte) ([]byte, error) {
// 	// Convert the image to a jpeg format
// 	jpegImage, err := bimg.NewImage(imgBuffer).Convert(bimg.JPEG)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return jpegImage, nil
// }

// func resize(jpegBuffer []byte, newWidth int, newHeight int) ([]byte, error) {
// 	// Resize the image to the new size
// 	resizedImage, err := bimg.NewImage(jpegBuffer).Resize(newWidth, newHeight)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return resizedImage, nil
// }

// func writeNewImage(r *Resources, jpegBuffer []byte, batchId string, imageFolder string, fileId string, suffix string, width int, height int) error {
// 	resizedFilename := fileId + suffix + ".jpeg"
// 	resizedLocalFilename := imageFolder + resizedFilename
// 	storagePath := r.config.S3MediaFolder + "/" + resizedFilename
// 	jpegContentType := "image/jpeg"

// 	// Resize the image
// 	// jpegBuffer, err := convertImageToPng(jpegBuffer)
// 	// if err != nil {
// 	// 	fmt.Printf("Error writing resized image: %s\n", err)
// 	// 	return err
// 	// }

// 	// Write the image to disk
// 	fmt.Printf("Writing %s\n", resizedLocalFilename)
// 	bimg.Write(resizedLocalFilename, jpegBuffer)

// 	defer os.Remove(resizedLocalFilename)

// 	// Upload the image to the storage adapter
// 	err := r.storage.PutFile(resizedLocalFilename, storagePath, jpegContentType)
// 	if err != nil {
// 		fmt.Printf("Error uploading file to storage: %s\n", err)
// 		return err
// 	}

// 	r.db.AddFile(&File{
// 		FileId:        fileId,
// 		ImportBatchId: batchId,
// 		Filename:      resizedFilename,
// 		StoragePath:   storagePath,
// 		PublicURL:     r.config.S3BaseUrl + storagePath,
// 		IsOriginal:    false,
// 		Width:         uint64(width),
// 		Height:        uint64(height),
// 	})

// 	// Remove the temporary file
// 	return nil
// }

// func processImageUpload(r *Resources, batchId string, imageFolder string, fileId string, extension string, contentType string) {
// 	filename := fileId + extension
// 	localPath := imageFolder + filename
// 	storagePath := r.config.S3MediaFolder + "/" + filename

// 	// Load image into buffer for resizing
// 	start := time.Now()
// 	imgBuffer, err := bimg.Read(localPath)
// 	if err != nil {
// 		fmt.Printf("Unable to read image file %s: %s\n", localPath, err)
// 		return
// 	}
// 	fmt.Printf("Reading image took %s\n", time.Since(start))

// 	start = time.Now()
// 	imgSize, err := bimg.Size(imgBuffer)
// 	if err != nil {
// 		fmt.Printf("Unable to read image size for %s: %s\n", fileId, err)
// 		return
// 	}
// 	fmt.Printf("Sizing image took %s\n", time.Since(start))

// 	// Upload the original file into storage
// 	start = time.Now()
// 	r.storage.PutFile(localPath, storagePath, contentType)
// 	fmt.Printf("Upload original image took %s\n", time.Since(start))

// 	r.db.AddFile(&File{
// 		FileId:        fileId,
// 		ImportBatchId: batchId,
// 		Filename:      filename,
// 		StoragePath:   storagePath,
// 		PublicURL:     r.config.S3BaseUrl + storagePath,
// 		IsOriginal:    true,
// 		Width:         uint64(imgSize.Width),
// 		Height:        uint64(imgSize.Height),
// 	})

// 	start = time.Now()
// 	jpegBuffer, err := convertImageToJpeg(imgBuffer)
// 	fmt.Printf("Converting to jpeg took %s\n", time.Since(start))

// 	if err != nil {
// 		fmt.Printf("Error converting to png: %s\n", err)
// 	}

// 	// Copy the original file size into png format
// 	start = time.Now()
// 	err = writeNewImage(r, jpegBuffer, batchId, imageFolder, fileId, "", imgSize.Width, imgSize.Height)
// 	fmt.Printf("Writing original took %s\n", time.Since(start))

// 	if err != nil {
// 		fmt.Printf("Error resizing image: %s\n", err)
// 	}

// 	// Compute long-edge resize images
// 	for _, size := range sizes {
// 		newWidth, newHeight := ResizeLongEdgeDimensions(imgSize.Width, imgSize.Height, size.LongEdge)

// 		start = time.Now()
// 		resizedJpegBuffer, err := resize(jpegBuffer, newWidth, newHeight)
// 		fmt.Printf("Resizing took %s\n", time.Since(start))

// 		if err != nil {
// 			fmt.Printf("Error resizing image: %s\n", err)
// 			continue
// 		}

// 		start = time.Now()
// 		err = writeNewImage(r, resizedJpegBuffer, batchId, imageFolder, fileId, size.Suffix, newWidth, newHeight)
// 		fmt.Printf("Writing resized image took %s\n", time.Since(start))

// 		if err != nil {
// 			fmt.Printf("Error writing image: %s\n", err)
// 			continue
// 		}
// 	}

// 	// Remove the uploaded file
// 	os.Remove(localPath)

// 	r.db.UpdateImageProcessedStatus(fileId, true)
// 	fmt.Printf("Resizing %s complete\n", localPath)
// }
