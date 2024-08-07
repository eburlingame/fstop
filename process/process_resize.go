package process

import (
	"log"
	"net/http"
	"sync"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/resources"
	. "github.com/eburlingame/fstop/utils"

	"github.com/h2non/bimg"
)

const DEFAULT_QUALITY = 75

func getImageSize(file []byte) (int, int, error) {
	sizes, err := bimg.Size(file)
	if err != nil {
		return 0, 0, err
	}

	meta, err := bimg.Metadata(file)
	if err != nil {
		return 0, 0, err
	}

	if meta.Orientation == 6 || meta.Orientation == 8 {
		return sizes.Height, sizes.Width, nil
	} else {
		return sizes.Width, sizes.Height, nil
	}
}

func resizeImage(file []byte, newWidth int, newHeight int) ([]byte, error) {
	imgBuffer, err := bimg.NewImage(file).Process(bimg.Options{
		Width:   newWidth,
		Height:  newHeight,
		Quality: DEFAULT_QUALITY,
	})

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
	if format == "webp" {
		return bimg.WEBP
	}
	return bimg.JPEG
}

func getResizedStorageFilename(r *Resources, image *ImageImport, size OutputImageSize) string {
	return image.ImageId + size.Suffix + size.Extension
}

func getOriginalStorageFilename(r *Resources, image *ImageImport) string {
	extension := GetExtension(image.OriginalFileKey)
	return image.ImageId + extension
}

func getStoragePath(r *Resources, filename string) string {
	return r.Config.S3MediaFolder + "/" + filename
}

func ProcessImageResize(r *Resources, wg *sync.WaitGroup, image *ImageImport, size OutputImageSize, file []byte) error {
	defer wg.Done()

	outputImage := file

	// Determine image dimensions
	width, height, err := getImageSize(file)
	if err != nil {
		log.Printf("Something went wrong: %s\n", err)
		return err
	}

	log.Printf("Image size %s to %d x %d\n", image.ImageId, width, height)

	originalLongEdge := GetLongestEdge(width, height)
	// Resize to longest edge, if needed
	if size.LongEdge < originalLongEdge {
		width, height = ResizeLongEdgeDimensions(width, height, size.LongEdge)
	}

	log.Printf("Resizing %s to %d x %d\n", image.ImageId, width, height)

	outputImage, err = bimg.NewImage(outputImage).Process(bimg.Options{
		Type:    imageTypeNameToEnum(size.Format),
		Quality: size.Quality,
		Width:   width,
		Height:  height,
	})

	if err != nil {
		log.Printf("Something went wrong: %s\n", err)
		return err
	}

	// Upload to storage
	storageFilename := getResizedStorageFilename(r, image, size)
	storagePath := getStoragePath(r, storageFilename)

	r.Storage.PutFile(outputImage, storagePath, size.ContentType)

	// Insert a FileRecord
	r.Db.AddFile(&File{
		FileId:        Uuid(),
		ImageId:       image.ImageId,
		ImportBatchId: image.ImportBatchId,
		Filename:      storageFilename,
		StoragePath:   storagePath,
		PublicURL:     r.Config.S3BaseUrl + storagePath,
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

	err = r.Storage.PutFile(outputImage, storagePath, http.DetectContentType(file))
	if err != nil {
		log.Printf("Error uploading to S3: %s\n", err)
		return err
	}

	// Insert a FileRecord
	err = r.Db.AddFile(&File{
		FileId:        Uuid(),
		ImageId:       image.ImageId,
		ImportBatchId: image.ImportBatchId,
		Filename:      storageFilename,
		StoragePath:   storagePath,
		PublicURL:     PublicImageURL(r.Config.S3BaseUrl, storagePath),
		IsOriginal:    true,
		Width:         uint64(width),
		Height:        uint64(height),
	})
	if err != nil {
		log.Printf("Error inserting file into database: %s\n", err)
		return err
	}

	return nil
}
