package main

import (
	"fmt"
	"net/http"

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

func extractExif(localPath string) error {
	data, err := exif.Read(localPath)

	if err != nil {
		return err
	}

	for key, val := range data.Tags {
		fmt.Printf("%s = %s\n", key, val)
	}

	return nil
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
			err := extractExif(localPath)
			if err != nil {
				fmt.Errorf("%s", err)
				continue
			}

		}
		c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	}

}
