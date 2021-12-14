package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HomeGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		var files []File

		r.db.ListLatestPhotos(&files, 400, 20, 0)

		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "Main website",
			"files": files,
		})
	}
}

func ImageGetHandler(r *Resources) gin.HandlerFunc {
	type UriParams struct {
		ImageId string `uri:"imageId" binding:"required"`
	}

	return func(c *gin.Context) {
		var params UriParams

		err := c.BindUri(&params)
		if err != nil {
			c.Status(404)
			return
		}

		var file File
		r.db.GetFile(&file, params.ImageId, 1000)

		c.HTML(http.StatusOK, "image.html", gin.H{
			"file": file,
		})
	}
}
