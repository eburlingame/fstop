package process

import (
	"fmt"
	"os"
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
		fmt.Printf("Unable to read image size %s\n", err)
		return err
	}

	image.HeightPixels = uint64(imgSize.Height)
	image.WidthPixels = uint64(imgSize.Width)

	return nil
}

func extractExif(localPath string) (map[string]string, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		fmt.Printf("Error when intializing: %v\n", err)
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

func ProcessImageMeta(r *Resources, wg *sync.WaitGroup, image *ImageImport, file []byte) error {
	defer wg.Done()

	extension := GetExtension(image.UploadFilePath)
	tempPath := "temp/" + image.ImageId + extension

	// Write to temporary file
	bimg.Write(tempPath, file)
	defer os.Remove(tempPath)

	// Extract image EXIF data
	tags, err := extractExif(tempPath)
	if err != nil {
		fmt.Printf("Error extracting EXIF data: %s\n", err)
		return err
	}

	// Create the image db entry
	imageRecord := Image{
		ImageId:       image.ImageId,
		IsProcessed:   false,
		ImportBatchId: image.ImportBatchId,
	}

	// Populate database Image with exif tags
	PopulateImageFromExif(&imageRecord, tags)

	// Populate image sizes
	err = populateImageSize(&imageRecord, file)
	if err != nil {
		return err
	}

	// Write the image to the database
	r.Db.AddImage(&imageRecord)

	return nil
}
