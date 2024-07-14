package process

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/resources"
	. "github.com/eburlingame/fstop/utils"

	"github.com/barasher/go-exiftool"
	"github.com/h2non/bimg"
)

func populateImageSize(image *Image, buffer []byte) error {
	imgSize, err := bimg.Size(buffer)
	if err != nil {
		log.Printf("Unable to read image size %s\n", err)
		return err
	}

	image.HeightPixels = uint64(imgSize.Height)
	image.WidthPixels = uint64(imgSize.Width)

	return nil
}

func extractExif(localPath string) (map[string]string, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		log.Printf("Error when intializing: %v\n", err)
		return nil, err
	}
	defer et.Close()

	fileInfos := et.ExtractMetadata(localPath)

	valueMap := map[string]string{}
	for tagName := range fileInfos[0].Fields {
		valueMap[tagName], _ = fileInfos[0].GetString(tagName)
	}

	return valueMap, nil
}

func ensureTempDirExists() {
	os.MkdirAll(os.TempDir(), os.ModePerm)
}

func ProcessImageMeta(r *Resources, wg *sync.WaitGroup, image *ImageImport, file []byte) error {
	defer wg.Done()

	ensureTempDirExists()

	extension := GetExtension(image.OriginalFileKey)
	tempPath := os.TempDir() + "/" + image.ImageId + extension

	// Write to temporary file
	log.Printf("Writing temporary file %s\n", tempPath)
	err := bimg.Write(tempPath, file)
	if err != nil {
		log.Printf("Error writing temporary file: %s\n", err)
		return err
	}
	defer os.Remove(tempPath)

	// Extract image EXIF data
	log.Printf("Extracting EXIF data %s\n", tempPath)
	tags, err := extractExif(tempPath)
	if err != nil {
		log.Printf("Error extracting EXIF data: %s\n", err)
		return err
	}

	// Create the image db entry
	imageRecord := Image{
		ImageId:          image.ImageId,
		ImportBatchId:    image.ImportBatchId,
		OriginalFilename: filepath.Base(image.OriginalFileKey),
	}

	// Populate database Image with exif tags
	log.Printf("Populating image from exif, imageId: %s\n", imageRecord.ImageId)
	PopulateImageFromExif(&imageRecord, tags)

	// Populate image sizes
	log.Printf("Populating image sizes, imageId: %s\n", imageRecord.ImageId)
	err = populateImageSize(&imageRecord, file)
	if err != nil {
		log.Printf("Error populating image sizes: %s\n", err)
		return err
	}

	// Write the image to the database
	log.Printf("Inserting image into database, imageId: %s\n", imageRecord.ImageId)
	err = r.Db.AddImage(&imageRecord)
	if err != nil {
		log.Printf("Error inserting image into database: %s\n", err)
		return err
	}

	// Add the image to the correct album, if set
	if image.AlbumId != "" {
		log.Printf("Adding image to album %s\n", image.AlbumId)
		r.Db.AddImageToAlbum(image.AlbumId, image.ImageId)
	}

	return nil
}
