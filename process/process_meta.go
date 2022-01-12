package process

import (
	"fmt"
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

func ensureTempDirExists() {
	os.MkdirAll(os.TempDir(), os.ModePerm)
}

func ProcessImageMeta(r *Resources, wg *sync.WaitGroup, image *ImageImport, file []byte) error {
	defer wg.Done()

	ensureTempDirExists()

	extension := GetExtension(image.UploadFilePath)
	tempPath := os.TempDir() + image.ImageId + extension

	// Write to temporary file
	err := bimg.Write(tempPath, file)
	if err != nil {
		fmt.Printf("Error writing temporary file: %s\n", err)
		return err
	}
	defer os.Remove(tempPath)

	// Extract image EXIF data
	tags, err := extractExif(tempPath)
	if err != nil {
		fmt.Printf("Error extracting EXIF data: %s\n", err)
		return err
	}

	// Create the image db entry
	imageRecord := Image{
		ImageId:          image.ImageId,
		ImportBatchId:    image.ImportBatchId,
		OriginalFilename: filepath.Base(image.UploadFilePath),
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

	// Add the image to the correct album, if set
	if image.AlbumId != "" {
		r.Db.AddImageToAlbum(image.AlbumId, image.ImageId)
	}

	return nil
}
