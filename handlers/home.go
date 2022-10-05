package handlers

import (
	"fmt"
	"net/http"

	. "github.com/eburlingame/fstop/middleware"
	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/resources"
	. "github.com/eburlingame/fstop/utils"

	"github.com/gin-gonic/gin"
)

func HomeGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		images, _ := r.Db.ListLatestImages(40, 0)

		var imagesWithSrcSets []ImageWithSrcSet

		for _, img := range images {
			metaDescription := fmt.Sprintf("%s, %s (%s' f/%.1f ISO %.0f)", img.CameraModel, img.Lens, img.ShutterSpeed, img.FNumber, img.ISO)

			imagesWithSrcSets = append(imagesWithSrcSets, ImageWithSrcSet{
				ImageId:       img.ImageId,
				SrcSet:        ComputeImageSrcSet((img.Files)),
				SmallImageUrl: FindSizedImage(img.Files, 500).PublicURL,
				Width:         img.WidthPixels,
				Height:        img.HeightPixels,
				Title:         img.DateTimeOriginal.Format("Monday, January _2, 2006"),
				Description:   metaDescription,
			})
		}

		c.HTML(http.StatusOK, "home.html", gin.H{
			"title":  "Main website",
			"images": imagesWithSrcSets,
		})
	}
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
			"srcSet":       ComputeImageSrcSet(files),
			"isAdmin":      isAdmin,
			"date":         image.DateTimeOriginal.Format("Monday, January _2, 2006"),
			"camera":       fmt.Sprintf("%s, %s", image.CameraModel, image.Lens),
			"meta":         fmt.Sprintf("%s' f/%.1f ISO %.0f", image.ShutterSpeed, image.FNumber, image.ISO),
		})
	}
}
