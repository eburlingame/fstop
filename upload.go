package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiam/exif"
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
	data, err := exif.Read(localPath)

	if err != nil {
		return nil, err
	}

	return data.Tags, nil
}

func parseExifTimestamp(s string) (time.Time, error) {
	// Exif date format: 2021:12:10 20:29:21
	layout := "2006:01:02 15:04:05"

	return time.Parse(layout, s)
}

func UploadHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Multipart form
		form, _ := c.MultipartForm()
		files := form.File["upload"]

		for _, file := range files {
			fileId := Uuid()
			extension := GetExtension(file.Filename)
			localPath := "./temp/" + fileId + extension

			// Upload the file to specific dst.
			c.SaveUploadedFile(file, localPath)

			// Extract image EXIF data
			var image Image

			tags, err := extractExif(localPath)

			ImageFromExif(&image, tags)

			if err != nil {
				fmt.Errorf("%s", err)
				continue
			}

		}
		c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	}

}
