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

func populateImageSize(image *Image, localPath string) error {
	// Load image into buffer for resizing
	imgBuffer, err := bimg.Read(localPath)
	if err != nil {
		fmt.Printf("Unable to read image file %s: %s\n", localPath, err)
		return err
	}

	// Store the image dimensions
	imgSize, err := bimg.Size(imgBuffer)
	if err != nil {
		fmt.Printf("Unable to read image size for %s: %s\n", localPath, err)
		return err
	}

	image.HeightPixels = uint64(imgSize.Height)
	image.WidthPixels = uint64(imgSize.Width)

	return nil
}

func AdminUploadPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Multipart form
		form, _ := c.MultipartForm()
		files := form.File["upload"]

		// Uploaded images
		imageFolder := "./temp/"
		uploadedImages := make([]Image, len(files))
		batchId := r.db.AddUploadBatch()

		imageBatch := make([]BatchProcessImage, len(files))

		for i, file := range files {
			fileId := Uuid()
			extension := GetExtension(file.Filename)
			localPath := imageFolder + fileId + extension

			// Upload the file to specific dst.
			c.SaveUploadedFile(file, localPath)

			// Extract image EXIF data
			tags, err := extractExif(localPath)
			if err != nil {
				fmt.Printf("Error extracting EXIF data: %s\n", err)
				continue
			}

			// Create the image db entry
			image := Image{
				FileId:        fileId,
				IsProcessed:   false,
				UploadBatchId: batchId,
			}

			// Populate database Image with exif tags
			PopulateImageFromExif(&image, tags)

			// Populate image sizes
			populateImageSize(&image, localPath)

			// Write the image to the database
			r.db.AddImage(&image)
			uploadedImages[i] = image

			// go processImageUpload(r, batchId, imageFolder, fileId, extension, image.MIMEType)
			imageBatch[i] = BatchProcessImage{
				FileId:    fileId,
				Extension: extension,
				MimeType:  image.MIMEType,
			}
		}

		// Kick off the image processing queue
		go ProcessImageBatch(r, batchId, imageFolder, imageBatch)

		c.HTML(http.StatusOK, "upload_complete.html", gin.H{
			"title":         "Upload Images",
			"header":        fmt.Sprintf("%d files uploaded:", len(uploadedImages)),
			"uploadBatchId": batchId,
		})
	}
}

func AdminUploadStatusGetHandler(r *Resources) gin.HandlerFunc {
	type UriParams struct {
		BatchId string `uri:"batchId" binding:"required"`
	}

	return func(c *gin.Context) {
		var params UriParams

		err := c.BindUri(&params)
		if err != nil {
			c.Status(404)
			return
		}

		var images []Image
		r.db.GetImagesInUploadBatch(&images, params.BatchId)

		type Status struct {
			IsProcessed bool
			URL         string
		}

		statuses := make([]Status, len(images))
		allProcessed := true

		for i, img := range images {
			statuses[i].IsProcessed = img.IsProcessed

			if img.IsProcessed {
				var file File

				r.db.GetFile(&file, img.FileId, 100)
				statuses[i].URL = file.PublicURL
			} else {
				allProcessed = false
			}
		}

		c.HTML(http.StatusOK, "upload_status_table.html", gin.H{
			"polling":       !allProcessed,
			"statuses":      statuses,
			"uploadBatchId": params.BatchId,
		})
	}
}
