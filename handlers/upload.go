package handlers

import (
	"net/http"

	. "github.com/eburlingame/fstop/resources"

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

		bucketPath := r.Config.S3UploadFolder + "/" + body.Filename

		signedUrl, err := r.Storage.GetSignedUploadUrl(bucketPath, body.ContentType)
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
