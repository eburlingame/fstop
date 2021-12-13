package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminUploadGet(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", gin.H{
			"title": "Upload Images",
		})
	}
}

func AdminUploadSignedUrlPostHandler(r *Resources) gin.HandlerFunc {
	type RequestBody struct {
		Filename    string `json:"filename"`
		ContentType string `json:"contentType"`
	}

	return func(c *gin.Context) {
		var body RequestBody
		c.Bind(&body)

		bucketPath := r.config.S3UploadFolder + "/" + body.Filename

		signedUrl, err := r.storage.GetSignedUploadUrl(bucketPath, body.ContentType)
		if err != nil {
			c.String(500, "Error getting signed url: %s", err)
			return
		}

		c.JSON(200, gin.H{
			"method": "put",
			"url":    signedUrl,
			"fields": []string{},
			"file":   gin.H{"type": body.ContentType},
		})
	}
}

func AdminUploadPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Multipart form
		// form, _ := c.MultipartForm()
		// files := form.File["upload"]

		// Uploaded images
		// imageFolder := "./temp/"
		// uploadedImages := make([]Image, len(files))
		// batchId := r.db.AddImportBatch()

		// imageBatch := make([]ImportBatch, len(files))

		// for i, file := range files {
		// 	fileId := Uuid()
		// 	extension := GetExtension(file.Filename)
		// 	localPath := imageFolder + fileId + extension

		// 	// Upload the file to specific dst.
		// 	c.SaveUploadedFile(file, localPath)

		// 	// Extract image EXIF data
		// 	tags, err := extractExif(localPath)
		// 	if err != nil {
		// 		fmt.Printf("Error extracting EXIF data: %s\n", err)
		// 		continue
		// 	}

		// 	// Create the image db entry
		// 	image := Image{
		// 		FileId:        fileId,
		// 		IsProcessed:   false,
		// 		ImportBatchId: batchId,
		// 	}

		// 	// Populate database Image with exif tags
		// 	PopulateImageFromExif(&image, tags)

		// 	// Populate image sizes
		// 	populateImageSize(&image, localPath)

		// 	// Write the image to the database
		// 	r.db.AddImage(&image)
		// 	uploadedImages[i] = image

		// 	// go processImageUpload(r, batchId, imageFolder, fileId, extension, image.MIMEType)
		// 	imageBatch[i] = ImportBatch{
		// 		FileId:    fileId,
		// 		Extension: extension,
		// 		MimeType:  image.MIMEType,
		// 	}
		// }

		// // Kick off the image processing queue
		// go ProcessImageBatch(r, batchId, imageFolder, imageBatch)

		// c.HTML(http.StatusOK, "upload_complete.html", gin.H{
		// 	"title":         "Upload Images",
		// 	"header":        fmt.Sprintf("%d files uploaded:", len(uploadedImages)),
		// 	"ImportBatchId": batchId,
		// })
	}
}
