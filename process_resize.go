package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/h2non/bimg"
)

func getImageSize(file []byte) (int, int, error) {
	sizes, err := bimg.Size(file)

	if err != nil {
		return 0, 0, err
	}

	return sizes.Width, sizes.Height, nil
}

func resizeImage(file []byte, newWidth int, newHeight int) ([]byte, error) {
	imgBuffer, err := bimg.NewImage(file).Resize(newWidth, newHeight)
	if err != nil {
		return nil, err
	}

	return imgBuffer, nil
}

func imageTypeNameToEnum(format string) bimg.ImageType {
	if format == "png" {
		return bimg.PNG
	}
	if format == "tiff" {
		return bimg.TIFF
	}
	if format == "heic" {
		return bimg.HEIF
	}
	return bimg.JPEG
}

func convertImageFormat(file []byte, imgTypeName string) ([]byte, error) {
	newType := imageTypeNameToEnum(imgTypeName)

	imgBuffer, err := bimg.NewImage(file).Convert(newType)
	if err != nil {
		return nil, err
	}

	return imgBuffer, nil
}

func getImageType(file []byte) string {
	return bimg.DetermineImageTypeName(file)
}

func getResizedStorageFilename(r *Resources, image *ImageImport, size OutputImageSize) string {
	return image.FileId + size.Suffix + size.Extension
}

func getOriginalStorageFilename(r *Resources, image *ImageImport) string {
	extension := GetExtension(image.UploadFilePath)
	return image.FileId + extension
}

func getStoragePath(r *Resources, filename string) string {
	return r.config.S3MediaFolder + "/" + filename
}

func ProcessImageResize(r *Resources, wg *sync.WaitGroup, image *ImageImport, size OutputImageSize, file []byte) error {
	defer wg.Done()

	outputImage := file

	// Determine image dimensions
	width, height, err := getImageSize(file)
	if err != nil {
		fmt.Printf("Something went wrong: %s\n", err)
		return err
	}

	originalLongEdge := GetLongestEdge(width, height)
	// Resize to longest edge, if needed
	if size.LongEdge < originalLongEdge {
		newWidth, newHeight := ResizeLongEdgeDimensions(width, height, size.LongEdge)
		outputImage, err = resizeImage(outputImage, newWidth, newHeight)

		if err != nil {
			fmt.Printf("Something went wrong: %s\n", err)
			return err
		}

		width = newWidth
		height = newHeight
	}

	// Convert image type, if needed
	if getImageType(outputImage) != size.Format {
		outputImage, err = convertImageFormat(outputImage, size.Format)

		if err != nil {
			fmt.Printf("Something went wrong: %s\n", err)
			return err
		}
	}

	// Upload to storage
	storageFilename := getResizedStorageFilename(r, image, size)
	storagePath := getStoragePath(r, storageFilename)

	r.storage.PutFile(outputImage, storagePath, size.ContentType)

	// Insert a FileRecord
	r.db.AddFile(&File{
		FileId:        image.FileId,
		ImportBatchId: image.ImportBatchId,
		Filename:      storageFilename,
		StoragePath:   storagePath,
		PublicURL:     r.config.S3BaseUrl + storagePath,
		IsOriginal:    false,
		Width:         uint64(width),
		Height:        uint64(height),
	})

	return nil
}

func ProcessImageOriginal(r *Resources, wg *sync.WaitGroup, image *ImageImport, file []byte) error {
	defer wg.Done()

	outputImage := file

	// Determine image dimensions
	width, height, err := getImageSize(file)
	if err != nil {
		return err
	}

	// Upload to storage
	storageFilename := getOriginalStorageFilename(r, image)
	storagePath := getStoragePath(r, storageFilename)

	r.storage.PutFile(outputImage, storagePath, http.DetectContentType(file))

	// Insert a FileRecord
	r.db.AddFile(&File{
		FileId:        image.FileId,
		ImportBatchId: image.ImportBatchId,
		Filename:      storageFilename,
		StoragePath:   storagePath,
		PublicURL:     r.config.S3BaseUrl + storagePath,
		IsOriginal:    true,
		Width:         uint64(width),
		Height:        uint64(height),
	})

	return nil
}
