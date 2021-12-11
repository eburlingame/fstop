package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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

func UploadHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		username := session.Get("authed_user")
		if username == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		// Multipart form
		form, _ := c.MultipartForm()
		files := form.File["upload"]

		for _, file := range files {
			extension := GetExtension(file.Filename)
			fileId := Uuid()

			// Upload the file to specific dst.
			c.SaveUploadedFile(file, "./temp/"+fileId+extension)
		}
		c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	}

}
