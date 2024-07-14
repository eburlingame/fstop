package handlers

import (
	"fmt"
	"net/http"
	"strings"

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

			smallImageFile := FindSizedImage(img.Files, 500)

			if smallImageFile != nil {
				imagesWithSrcSets = append(imagesWithSrcSets, ImageWithSrcSet{
					ImageId:       img.ImageId,
					SrcSet:        ComputeImageSrcSet(r.Config.S3BaseUrl, (img.Files)),
					SmallImageUrl: PublicImageURL(r.Config.S3BaseUrl, smallImageFile.StoragePath),
					Width:         img.WidthPixels,
					Height:        img.HeightPixels,
					Title:         img.DateTimeOriginal.Format("Monday, January _2, 2006"),
					Description:   metaDescription,
				})
			}
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

		type ImageFile struct {
			ImageId    string
			PublicURL  string
			Width      uint64
			Height     uint64
			IsOriginal bool
		}

		r.Db.ListImageFiles(&files, params.ImageId)
		r.Db.GetImage(&image, params.ImageId)

		if len(files) == 0 {
			c.Status(404)
			return
		}

		renderedFiles := []ImageFile{}
		for _, file := range files {
			if strings.HasSuffix(file.StoragePath, ".webp") {
				renderedFiles = append(renderedFiles, ImageFile{
					ImageId:    file.ImageId,
					Width:      file.Width,
					Height:     file.Height,
					IsOriginal: file.IsOriginal,
					PublicURL:  PublicImageURL(r.Config.S3BaseUrl, file.StoragePath),
				})
			}
		}

		c.HTML(http.StatusOK, "image.html", gin.H{
			"files":        renderedFiles,
			"smallestFile": renderedFiles[0],
			"srcSet":       ComputeImageSrcSet(r.Config.S3BaseUrl, files),
			"isAdmin":      isAdmin,
			"date":         image.DateTimeOriginal.Format("Monday, January _2, 2006"),
			"camera":       fmt.Sprintf("%s, %s", image.CameraModel, image.Lens),
			"meta":         fmt.Sprintf("%s' f/%.1f ISO %.0f", image.ShutterSpeed, image.FNumber, image.ISO),
		})
	}
}
