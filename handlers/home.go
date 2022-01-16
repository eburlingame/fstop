package handlers

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/eburlingame/fstop/middleware"
	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/resources"

	"github.com/gin-gonic/gin"
)

func HomeGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		files, _ := r.Db.ListLatestPhotos(400, 40, 0)

		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "Main website",
			"files": files,
		})
	}
}

func computeImageSrcSet(files []File) string {
	srcs := []string{}

	for _, file := range files {
		srcs = append(srcs, fmt.Sprintf("%s %dw", file.PublicURL, file.Width))
	}

	return strings.Join(srcs, ", ")
}

func ImageGetHandler(r *Resources) gin.HandlerFunc {

	type UriParams struct {
		ImageId string `uri:"imageId" binding:"required"`
	}

	return func(c *gin.Context) {
		isAdmin := IsLoggedIn(r, c)

		var params UriParams

		err := c.BindUri(&params)
		if err != nil {
			c.Status(404)
			return
		}

		var files []File
		var image Image

		r.Db.ListImageFiles(&files, params.ImageId)
		r.Db.GetImage(&image, params.ImageId)

		c.HTML(http.StatusOK, "image.html", gin.H{
			"files":        files,
			"smallestFile": files[0],
			"srcSet":       computeImageSrcSet(files),
			"isAdmin":      isAdmin,
			"date":         image.DateTimeOriginal.Format("Monday, January _2, 2006"),
			"camera":       fmt.Sprintf("%s, %s", image.CameraModel, image.Lens),
			"meta":         fmt.Sprintf("%s' f/%.1f ISO %.0f", image.ShutterSpeed, image.FNumber, image.ISO),
		})
	}
}
