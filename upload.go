package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/h2non/bimg"
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
			tags, err := extractExif(localPath)
			if err != nil {
				fmt.Errorf("%s", err)
				continue
			}

			var image Image
			PopulateImageFromExif(&image, tags)

			buffer, _ := bimg.Read(localPath)
			imgSize, _ := bimg.Size(buffer)

			image.HeightPixels = uint32(imgSize.Height)
			image.WidthPixels = uint32(imgSize.Width)

			r.db.AddImage(&image)
		}
		c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	}

}
