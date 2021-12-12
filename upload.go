package main

import (
	"fmt"
	"net/http"

	"github.com/barasher/go-exiftool"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/h2non/bimg"
)

func AdminUploadGet(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		username := session.Get("authed_user")
		if username == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		c.HTML(http.StatusOK, "upload.html", gin.H{
			"title":    "Upload Images",
			"username": username,
		})
	}
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

func convertImageToPng(imgBuffer []byte) ([]byte, error) {
	// Convert the image to a png format
	pngImage, err := bimg.NewImage(imgBuffer).Convert(bimg.PNG)
	if err != nil {
		return nil, err
	}

	return pngImage, nil
}

func resizeImage(pngImageBuffer []byte, destFilaname string, newWidth int, newHeight int) error {
	// Resize the image to the new size
	resizedImage, err := bimg.NewImage(pngImageBuffer).Resize(newWidth, newHeight)
	if err != nil {
		return err
	}

	// Write image to disk
	_ = resizedImage
	fmt.Printf("Writing %s\n", destFilaname)
	// return bimg.Write(destFilaname, resizedImage)
	return nil
}

type OutputImageSize struct {
	LongEdge int
	Suffix   string
}

func UploadHandler(r *Resources) gin.HandlerFunc {

	sizes := []OutputImageSize{
		{
			LongEdge: 200,
			Suffix:   "thumb",
		},
		{
			LongEdge: 600,
			Suffix:   "small",
		},
		{
			LongEdge: 1080,
			Suffix:   "medium",
		},
		{
			LongEdge: 1920,
			Suffix:   "large",
		},
	}
	_ = sizes

	return func(c *gin.Context) {
		// Multipart form
		form, _ := c.MultipartForm()
		files := form.File["upload"]

		for _, file := range files {
			fileId := Uuid()
			extension := GetExtension(file.Filename)
			imageFolder := "./temp/"
			localPath := imageFolder + fileId + extension

			// Upload the file to specific dst.
			c.SaveUploadedFile(file, localPath)

			// Extract image EXIF data
			tags, err := extractExif(localPath)
			if err != nil {
				fmt.Printf("Error extracting EXIF data: %s\n", err)
				continue
			}

			// Populate database Image with exif tags
			var image Image
			PopulateImageFromExif(&image, tags)

			// Load image into buffer for resizing
			imgBuffer, err := bimg.Read(localPath)
			if err != nil {
				fmt.Printf("Unable to read image file %s: %s\n", localPath, err)
				continue
			}

			// Store the image dimensions
			imgSize, err := bimg.Size(imgBuffer)
			if err != nil {
				fmt.Printf("Unable to read image size for %s: %s\n", localPath, err)
				continue
			}

			image.HeightPixels = uint32(imgSize.Height)
			image.WidthPixels = uint32(imgSize.Width)

			pngImageBuffer, err := convertImageToPng(imgBuffer)
			if err != nil {
				fmt.Printf("Error converting to png: %s\n", err)
			}

			// Compute long-edge resize images
			for _, size := range sizes {
				destFile := imageFolder + fileId + "_" + size.Suffix + ".png"

				newWidth, newHeight := ResizeLongEdgeDimensions(imgSize.Width, imgSize.Height, size.LongEdge)
				err := resizeImage(pngImageBuffer, destFile, newWidth, newHeight)

				if err != nil {
					fmt.Printf("Error writing resized image: %s\n", err)
					continue
				}
			}

			// Copy the original file size into png format
			destFile := imageFolder + fileId + ".png"
			err = resizeImage(imgBuffer, destFile, imgSize.Width, imgSize.Height)
			if err != nil {
				fmt.Printf("Error writing resized image: %s\n", err)
				continue
			}

			r.db.AddImage(&image)
		}

		c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	}
}
